#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2013-2014, Marten de Vries
#
#	This file is part of OpenTeacher.
#
#	OpenTeacher is free software: you can redistribute it and/or modify
#	it under the terms of the GNU General Public License as published by
#	the Free Software Foundation, either version 3 of the License, or
#	(at your option) any later version.
#
#	OpenTeacher is distributed in the hope that it will be useful,
#	but WITHOUT ANY WARRANTY; without even the implied warranty of
#	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#	GNU General Public License for more details.
#
#	You should have received a copy of the GNU General Public License
#	along with OpenTeacher.  If not, see <http://www.gnu.org/licenses/>.

import requests
import json
import uuid
import logging
import os

join = os.path.join

logger = logging.getLogger(__name__)

def _only_access_for(username):
	return {
		"admins": {
			"names": [],
			"roles": [],
		},
		"members": {
			"names": [username],
			"roles": [],
		}
	}

ADMIN_ONLY_VALIDATE_DOC_UPDATE = """
(function (newDoc, oldDoc, userCtx) {
	if (userCtx.roles.indexOf("_admin") === -1) {
		throw({"unauthorized": "need to be admin to make changes."});
	}
});
"""

class WebCouch(object):
	def __init__(self, host, username, password, dbSkeletonDir, isSafeHtmlCode, generateWordsHtml, *args, **kwargs):
		super(WebCouch, self).__init__(*args, **kwargs)

		self._host = host
		self._username = username
		self._password = password

		self._codeDir = dbSkeletonDir
		self._validationLib = isSafeHtmlCode + "\n" + self._getJs("validationLib.js")
		self._presentationLib = generateWordsHtml + "\n" + self._getJs("presentationLib.js")

	def _getJs(self, endpoint):
		with open(join(self._codeDir, endpoint)) as f:
			return f.read()

	def _design_from(self, endpoint, additionalData={}):
		base = join(self._codeDir, endpoint)
		parts = os.listdir(base)
		design_doc = {}

		if "validate_doc_update.js" in parts:
			design_doc["validate_doc_update"] = self._getJs(join(endpoint, "validate_doc_update.js"))

		for type in ["updates", "lists", "shows"]:
			if type in parts:
				design_doc[type] = {}

				allOfType = os.listdir(join(base, type))
				for oneOfTypeJs in allOfType:
					oneOfType = oneOfTypeJs[:-len(".js")]
					design_doc[type][oneOfType] = self._getJs(join(endpoint, type, oneOfTypeJs))

		if "views" in parts:
			design_doc["views"] = self._gatherViews(join(base, "views"), endpoint)

		design_doc.update(additionalData)
		return design_doc

	def _gatherViews(self, path, endpoint):
		viewsObj = {}

		views = os.listdir(path)
		for view in views:
			viewObj = {
				"map": self._getJs(join(endpoint, "views", view, "map.js")),
			}
			reducePath = join(endpoint, "views", view, "reduce.js")
			if os.path.exists(join(self._codeDir, reducePath)):
				viewObj["reduce"] = self._getJs(reducePath)
			viewsObj[view] = viewObj
		return viewsObj

	def req(self, method, endpoint, data=None, auth=None):
		if data:
			data = json.dumps(data)
		if not auth:
			auth = requests.auth.HTTPBasicAuth(self._username, self._password)
		headers = {"Content-Type": "application/json"}
		func = getattr(requests, method)
		return func(self._host + endpoint, auth=auth, headers=headers, data=data)

	def create_anonymous_user(self):
		self._create_user("anonymous", "anonymous", anonymous=True)

	def _create_user(self, username, password, anonymous):
		try:
			#user
			assert self.req("post", "/_users", {
				"_id": "org.couchdb.user:" + username,
				"type": "user",
				"name": username,
				"password": password,
				"roles": [],
			}).status_code == 201

			#private
			assert self.req("put", "/private_" + username).status_code == 201
			assert self.req(
				"put",
				"/private_" + username + "/_security",
				_only_access_for(username)
			).status_code == 200
			private_design_doc = self._design_from("private", {
				"validation_lib": self._validationLib,
			})
			if anonymous:
				private_design_doc["validate_doc_update"] = ADMIN_ONLY_VALIDATE_DOC_UPDATE
			assert self.req("put", "/private_" + username + "/_design/private", private_design_doc).status_code == 201

			lists_design_doc = self._design_from("lists", {
				"presentation_lib": self._presentationLib,
			})
			assert self.req("put", "/private_" + username + "/_design/lists", lists_design_doc).status_code == 201

			tests_design_doc = self._design_from("tests")
			assert self.req("put", "/private_" + username + "/_design/tests", tests_design_doc).status_code == 201

			#shared_lists
			assert self.req("put", "/shared_lists_" + username).status_code == 201
			assert self.req("put", "/shared_lists_" + username + "/_design/shares", self._design_from("shared_lists", {
				"presentation_lib": self._presentationLib,
			})).status_code == 201

			#replicator
			assert self.req("put", "/_replicator/private_to_shared_lists_" + username, {
				"source": "private_" + username,
				"target": "shared_lists_" + username,
				"continuous": True,
				"user_ctx": {
					"roles": ["_admin"],
				},
				#Workaround for https://issues.apache.org/jira/browse/COUCHDB-1415
				"random_value": str(uuid.uuid4()),
			}).status_code == 201

			#add one 'default' list
			assert self.req("post", "/private_" + username + "/_design/lists/_update/set_last_edited_to_now", {
				"type": "list",
				"shares": [],
				"items": [
					{
						"id": 0,
						"questions": [["een"]],
						"answers": [["one"]],
					},
					{
						"id": 1,
						"questions": [["twee"]],
						"answers": [["two"]],
					},
					{
						"id": 1,
						"questions": [["drie"]],
						"answers": [["three"]],
					}
				],
				"title": "Example list",
			}).status_code == 201

		except AssertionError, e:
			logger.debug(e, exc_info=True)
			try:
				self.delete_user(username)
			except ValueError:
				#delete as far as possible, but it'll probably crash since
				#create_user didn't succeed.
				pass
			raise ValueError("username_taken")

	def new_user(self, username, password):
		if len(password) < 8 or password.isalnum():
			raise ValueError("unsafe_password")

		self._create_user(username, password, anonymous=False)

	def check_auth(self, username, password):
		session = self.req("get", "/_session", auth=requests.auth.HTTPBasicAuth(username, password)).json()
		try:
			return username == session["userCtx"]["name"]
		except KeyError:
			return False

	def delete_user(self, username):
		try:
			#no assertion, because this could actually fail, because the
			#order is different than in new_user. And it has to be this way
			#because the DB's that are replicated shouldn't be deleted
			#already when the replication is done.
			rev = json.loads(self.req("head", "/_replicator/private_to_shared_lists_" + username).headers["Etag"])
			self.req("delete", "/_replicator/private_to_shared_lists_" + username + "?rev=" + rev)

			rev = json.loads(self.req("head", "/_users/org.couchdb.user:" + username).headers["Etag"])
			assert self.req("delete", "/_users/org.couchdb.user:" + username + "?rev=" + rev).status_code == 200

			assert self.req("delete", "/private_" + username).status_code == 200
			assert self.req("delete", "/shared_lists_" + username).status_code == 200
		except (AssertionError, KeyError), e:
			logger.debug(e, exc_info=True)
			raise ValueError("Error deleting user, aborted.")
