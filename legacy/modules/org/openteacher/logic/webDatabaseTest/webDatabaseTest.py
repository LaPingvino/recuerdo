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

import unittest
import json
import traceback

class TestCase(unittest.TestCase):
	@property
	def _mods(self):
		if self.mode != "web-database":
			self.skipTest("only run in the 'web-database' test mode (not even in 'all'), since it requires a CouchDB database to be installed & credentials for it to be in the config file of this test module. Also, it requires user interaction to let it tear down the generated db again.")
		return self._mm.mods("active", type="webDatabase")

	def test(self):
		for mod in self._mods:
			with open(self._mm.resourcePath("web-config.json")) as f:
				config = json.load(f)
			couch = mod.createWebDatabase(
				config["COUCHDB_HOST"],
				config["COUCHDB_ADMIN_USERNAME"],
				config["COUCHDB_ADMIN_PASSWORD"],
			)

			#set up
			TEST_USER = "test"
			TEST_PASSW = "hsdfjkh3$4"

			TEST2_USER = "test2"
			TEST2_PASSW = "s2Sjgjk*92"

			couch.create_anonymous_user()
			couch.new_user(TEST_USER, TEST_PASSW)
			testAuth = requests.auth.HTTPBasicAuth(TEST_USER, TEST_PASSW)

			listId = couch.req("post", "/private_test/_design/lists/_update/set_last_edited_to_now", {
				"type": "list",
				"shares": ["testShare", "testShareB"],
				"items": [
					{
						"id": 0,
						"questions": [["een"]],
						"answers": [["a single", "one"], ["a"]],
						"comment": u"Silly example...",
					},
					{
						"id": 1,
						"questions": [["twee"]],
						"answers": [["two"]],
					}
				],
				"title": "something",
				"questionLanguage": "Dutch",
				"answerLanguage": "English",
			}, auth=testAuth).json()["_id"]
			assert couch.req("post", "/private_test", {
				"type": "test",
				"listId": listId,

				"finished": True, 
				"results": [
					{
						"itemId": 0, 
						"result": "right", 
						"givenAnswer": "one, a",
						"active": {
							"start":"2013-07-22T18:14:15.015Z",
							"end":"2013-07-22T18:14:23.637Z",
						},
					}, 
					{
						"itemId": 1, 
						"result": "wrong", 
						"givenAnswer": "ttwo"
					},
				], 
				"pauses": [],
			}).status_code == 201

			assert couch.req("post", "/private_test/_design/lists/_update/set_last_edited_to_now", {
				"type": "list",
				"shares": ["testShare"],
				"title": "a",
			}, auth=testAuth).status_code == 201
			assert couch.req("post", "/private_test/_design/lists/_update/set_last_edited_to_now", {
				"type": "list",
				"shares": ["testShare"],
				"title": "b",
			}, auth=testAuth).status_code == 201
			assert couch.req("post", "/private_test/_design/lists/_update/set_last_edited_to_now", {
				"type": "list",
				"shares": ["testShareB"],
				"title": "c",
			}, auth=testAuth).status_code == 201
			assert couch.req("post", "/private_test/_design/lists/_update/set_last_edited_to_now", {
				"type": "list",
				"shares": [],
				"title": "d",
			}, auth=testAuth).status_code == 201
			assert couch.req("post", "/private_test", {
				"_id": "org.openteacher.test",
				"type": "setting",
				"value": 123,
			}, auth=testAuth).status_code == 201

			#see: https://wiki.apache.org/couchdb/View_collation#Complex_keys
			try:
				assert len(couch.req("get", '/private_test/_design/lists/_view/by_title?startkey=["a"]&endkey=["b", {}]').json()["rows"]) == 2
				assert couch.req("get", "/private_test/_design/lists/_show/print/" + listId).status_code == 200
				assert len(couch.req("get", "/shared_lists_test/_design/shares/_view/share_names?group=true").json()["rows"]) == 2
				assert len(couch.req("get", '/shared_lists_test/_design/shares/_view/by_name?startkey=["testShareB"]&endkey=["testShareB", {}, {}]').json()["rows"]) == 2
				assert len(couch.req("get", '/private_test/_design/tests/_view/by_list_id?startkey=["%s", {}]&endkey=["%s"]&descending=true' % (listId, listId)).json()["rows"]) == 1
			except:
				traceback.print_exc()

			couch.new_user(TEST2_USER, TEST2_PASSW)
			test2Auth = requests.auth.HTTPBasicAuth(TEST2_USER, TEST2_PASSW)

			try:
				assert couch.req("get", "/private_test/_design/lists/_view/by_title", auth=test2Auth).status_code == 401
				assert couch.req("get", "/shared_lists_test/_design/shares/_view/share_names?group=true").status_code == 200
				assert couch.req("get", '/shared_lists_test/_design/shares/_view/by_name?startkey=["testShareB"]&endkey=["testShareB", {}, {}]').status_code == 200
				assert couch.req("get", '/shared_lists_test/_design/shares/_list/feed/by_name?startkey=["testShareB"]&endkey=["testShareB", {}, {}]').status_code == 200
			except:
				traceback.print_exc()

			raw_input("Set up the database & ran db tests. Press enter to tear it down again... ")

			#and tear down
			couch.delete_user(TEST2_USER)
			couch.delete_user(TEST_USER)
			couch.delete_user("anonymous")

			couch.req("post", "/_users/_compact")
			couch.req("post", "/_replicator/_compact")

class TestModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "test"
		self.requires = (
			self._mm.mods(type="webDatabase"),
		)

	def enable(self):
		global requests
		try:
			import requests
		except ImportError:
			return
		self.TestCase = TestCase
		self.TestCase._mm = self._mm
		self.active = True

	def disable(self):
		self.active = False
		del self.TestCase

def init(moduleManager):
	return TestModule(moduleManager)
