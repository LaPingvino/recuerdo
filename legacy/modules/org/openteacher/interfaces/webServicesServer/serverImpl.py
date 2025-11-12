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

import flask
import recaptcha.client.captcha
import feedparser
import functools
import collections
import superjson
import tempfile
import os
import datetime
import json
import contextlib
import datetime
import locale
import hmac
import hashlib
import requests

locale.setlocale(locale.LC_ALL, "C")

#Imports handled by the module:
#
#import createWebDatabase
#import gettextFunctions
#import loaders
#import savers
#import metadata

app = flask.Flask(__name__)

class DummyLesson(object):
	def __init__(self, list, *args, **kwargs):
		super(DummyLesson, self).__init__(*args, **kwargs)

		self.list = list
		self.resources = {}

#utils
def get_couch():
	try:
		return flask.g.couch
	except AttributeError:
		flask.g.couch = createWebDatabase(
			app.config["COUCHDB_HOST"],
			app.config["COUCHDB_ADMIN_USERNAME"],
			app.config["COUCHDB_ADMIN_PASSWORD"],
		)
		return get_couch()

feed = {}
def get_feed():
	last_updated = feed.get("last_update", datetime.datetime.min)
	if datetime.datetime.now() - last_updated > datetime.timedelta(minutes=30):
		feed["data"] = feedparser.parse(metadata["newsFeedUrl"])
		feed["last_update"] = datetime.datetime.now()
	return feed["data"]

def json_default(item):
	try:
		return item.isoformat()
	except AttributeError, e:
		raise TypeError("Can't provide a default value")

def jsonify(data):
	dataJson = superjson.dumps(data, indent=4, default=json_default)
	resp = flask.make_response(dataJson)
	resp.headers["Content-Type"] = "application/json"
	return resp

def json_err(msg, code=400):
	resp = jsonify({"error": msg})
	resp.status_code = code
	return resp

def auth_err(msg):
	resp = json_err(msg, 401)
	resp.headers["WWW-Authenticate"] = 'Basic realm="%s Web"' % metadata["name"]
	return resp

def requires_auth(f):
	@functools.wraps(f)
	def decorated(*args, **kwargs):
		auth = flask.request.authorization
		if not auth and flask.request.form.get("username"):
			auth = {
				"username": flask.request.form.get("username"),
				"password": flask.request.form.get("password", ""),
			}
		if not auth:
			return auth_err("authentication_required")
		if auth["username"] == "anonymous":
			return auth_err("The user 'anonymous' isn't allowed to use the services API.")
		if not get_couch().check_auth(auth["username"], auth["password"]):
			return auth_err("wrong_username_or_password")
		return f(*args, **kwargs)
	return decorated

#Thanks Armin Ronacher! See: http://flask.pocoo.org/snippets/56/
def crossdomain(origin=None, methods=None, headers=None, credentials="false", max_age=21600, attach_to_all=True, automatic_options=True):
	if not origin:
		return lambda f: f
	if methods is not None:
		methods = ', '.join(sorted(x.upper() for x in methods))
	if headers is not None and not isinstance(headers, basestring):
		headers = ', '.join(x.upper() for x in headers)
	if not isinstance(origin, basestring):
		origin = ', '.join(origin)
	if isinstance(max_age, datetime.timedelta):
		max_age = max_age.total_seconds()

	def get_methods():
		if methods is not None:
			return methods

		options_resp = app.make_default_options_response()
		return options_resp.headers['allow']

	def decorator(f):
		def wrapped_function(*args, **kwargs):
			if automatic_options and flask.request.method == 'OPTIONS':
				resp = app.make_default_options_response()
			else:
				resp = flask.make_response(f(*args, **kwargs))
			if not attach_to_all and flask.request.method != 'OPTIONS':
				return resp

			h = resp.headers

			h['Access-Control-Allow-Origin'] = origin
			h['Access-Control-Allow-Methods'] = get_methods()
			h['Access-Control-Max-Age'] = str(max_age)
			if headers is not None:
				h['Access-Control-Allow-Headers'] = headers
			if credentials:
				h['Access-Control-Allow-Credentials'] = "true"
			return resp

		f.provide_automatic_options = False
		return functools.update_wrapper(wrapped_function, f)
	return decorator

def allow_cross_domain_from_trusted_origins(f):
	return crossdomain(origin=app.config["TRUSTED_ORIGINS"], credentials=True, headers="authorization")(f)

def load_exts(_cache=[]):
	if not _cache:
		for loader in loaders:
			for ext, dataTypes in loader.loads.iteritems():
				if "words" in dataTypes:
					_cache.append(ext)
	return _cache

def save_exts(_cache=[]):
	if not _cache:
		for saver in savers:
			_cache.extend(saver.saves.get("words", []))
	return _cache

def jsonp(f):
	@functools.wraps(f)
	def wrapper(*args, **kwargs):
		resp = f(*args, **kwargs)
		try:
			callback = flask.request.args["callback"]
		except KeyError:
			return resp
		else:
			content = str(callback) + "(" + superjson.dumps(resp) + ")"
			resp = flask.make_response(content)
			resp.headers["Content-Type"] = "application/javascript"
			return resp
	return wrapper

def initialize_endpoints():
	with app.app_context():
		if not get_couch().check_auth("anonymous", "anonymous"):
			get_couch().create_anonymous_user()

	#services
	@app.route("/", methods=["OPTIONS", "GET"])
	@allow_cross_domain_from_trusted_origins
	def services():
		data = collections.OrderedDict([
			("welcome", "%s Web services" % metadata["name"]),
			("info", "All api entry points require HTTP Basic Authentication."),
			("web_entry_points", [
				flask.url_for("register"),
				flask.url_for("news"),
			]),
			("api_entry_points", [
				flask.url_for("deregister"),
				flask.url_for("load"),
				flask.url_for("supported_load_extensions"),
				flask.url_for("save"),
				flask.url_for("supported_save_extensions"),
				flask.url_for("build_ci"),
			])
		])
		return jsonify(data)

	@app.route("/news")
	@jsonp
	def news():
		def nice_time(struct):
			obj = datetime.datetime(*struct[:6])
			return obj.strftime("%H:%M:%S at %B %d, %Y")
		with open(app.config["NEWS_TEMPLATE_PATH"]) as f:
			return flask.render_template_string(f.read(), feed=get_feed(), nice_time=nice_time)

	@app.route("/register")
	def register():
		try:
			redirect = flask.request.args["redirect"]
		except KeyError:
			return "<h1>Required GET URL parameter: redirect.</h1>"
		try:
			language = flask.request.args["language"]
		except KeyError:
			return "<h1>Required GET URL parameter: language. ('C' will suffice if your app is available in English only.)</h1>"
		screenshotOnly = flask.request.args.get("screenshotonly", "false")

		_, ngettext = gettextFunctions(language)

		error = {
			"invalid_captcha": _("Invalid captcha. Please try again."),
			"unsafe_password": _("Your password should at least be %s characters long, and contain special (non alphanumeric) characters. Please try again.") % 8,
			"username_taken": _("The username you requested is already taken. Please try again.")
		}.get(flask.request.args.get("error"), u"")

		if screenshotOnly == "true":
			#fetching a captcha is too much overhead when the website is
			#just shown for 'screenshot' purposes.
			captcha = ""
		else:
			publicKey = app.config["RECAPTCHA_PUBLIC_KEY"]
			captcha = recaptcha.client.captcha.displayhtml(publicKey)

		data = {"captcha": captcha, "redirect": redirect, "error": error, "_": _, "ngettext": ngettext, "language": language}
		with open(app.config["REGISTER_TEMPLATE_PATH"]) as f:
			return flask.render_template_string(f.read(), **data)

	@app.route("/register/send", methods=["OPTIONS", "POST"])
	def register_send():
		redirect_url = flask.request.form["redirect"]
		language = flask.request.form["language"]
		def error(e):
			return flask.redirect(flask.url_for("register") + "?error=" + e + "&redirect=" + redirect_url + "&language=" + language)

		try:
			challenge = flask.request.form["recaptcha_challenge_field"]
			response = flask.request.form["recaptcha_response_field"]
		except KeyError:
			return error("invalid_captcha")

		private_key = app.config["RECAPTCHA_PRIVATE_KEY"]
		ip = flask.request.remote_addr
		valid = recaptcha.client.captcha.submit(challenge, response, private_key, ip).is_valid
		if not valid:
			return error("invalid_captcha")

		username = flask.request.form["username"]
		password = flask.request.form["password"]
		try:
			get_couch().new_user(username, password)
		except ValueError, e:
			return error(str(e))

		return flask.redirect(redirect_url + "?status=ok")

	@app.route("/deregister", methods=["OPTIONS", "POST"])
	@allow_cross_domain_from_trusted_origins
	@requires_auth
	def deregister():
		auth = flask.request.authorization
		try:
			get_couch().delete_user(auth.username)
		except ValueError, e:
			return json_err(str(e))
		return jsonify({"result": "ok"})

	@app.route("/load/supported_extensions", methods=["GET", "OPTIONS"])
	@allow_cross_domain_from_trusted_origins
	@requires_auth
	def supported_load_extensions():
		return jsonify({"result": load_exts()})

	@app.route("/load", methods=["OPTIONS", "POST"])
	@allow_cross_domain_from_trusted_origins
	@requires_auth
	def load():
		try:
			f = flask.request.files["file"]
		except KeyError, e:
			return json_err("Please upload a file (name='file')")
		ext = os.path.splitext(f.filename)[1]
		#strip the . in .otwd
		if not ext[1:] in load_exts():
			return json_err("Invalid file type for the uploaded file.")

		fd, path = tempfile.mkstemp(ext)
		os.close(fd)
		f.save(path)

		resp = json_err("Couldn't load file")
		for loader in loaders:
			if loader.getFileTypeOf(path) == "words":
				try:
					result = loader.load(path)
				except Exception:
					continue
				else:
					resp = jsonify(result["list"])
					break

		os.remove(path)
		return resp

	@app.route("/save/supported_extensions", methods=["GET", "OPTIONS"])
	@allow_cross_domain_from_trusted_origins
	@requires_auth
	def supported_save_extensions():
		return jsonify({"result": save_exts()})

	@app.route("/save", methods=["OPTIONS", "POST"])
	@allow_cross_domain_from_trusted_origins
	@requires_auth
	def save():
		try:
			list = json.loads(flask.request.form["list"])
		except KeyError:
			return json_err("Please specify the 'list' field in the body form data.")
		except ValueError:
			return json_err("The posted list is invalid json.")

		dateFormat = "%Y-%m-%dT%H:%M:%S.%fZ"
		for item in list.get("items", []):
			with contextlib.suppress(KeyError):
				item["created"] = datetime.datetime.strptime(item["created"], dateFormat)
		for test in list.get("tests", []):
			for result in test.get("results", []):
				with contextlib.suppress(KeyError):
					result["active"]["start"] = datetime.datetime.strptime(result["active"]["start"], dateFormat)
					result["active"]["end"] = datetime.datetime.strptime(result["active"]["end"], dateFormat)
		lesson = DummyLesson(list)

		try:
			filename = flask.request.form["filename"]
		except KeyError:
			return json_err("Please specify the 'filename' field in the body form data.")

		#[1:] tears of the . in .txt
		ext = os.path.splitext(filename)[1][1:]
		if not ext in save_exts():
			return json_err("Unsupported filename extension. (%s)" % ext)

		fd, path = tempfile.mkstemp("." + ext)
		os.close(fd)
		for saver in savers:
			if ext in saver.saves.get("words", []):
				saver.save("words", lesson, path)
				resp = flask.send_file(path, as_attachment=True, attachment_filename=filename)
				break

		os.remove(path)
		return resp

	@app.route("/build-ci", methods=["POST"])
	def build_ci():
		signature = flask.request.headers.get("X-Hub-Signature", "")
		try:
			contentLength = int(flask.request.headers.get("Content-Length"))
		except ValueError:
			return json_err("Invalid content length")
		if contentLength > 1024 * 1024:
			# 1 MB
			return json_err("Body too large")
		body = flask.request.get_data()
		expected = "sha1=" + hmac.new(app.config["SIGNATURE_SECRET"], body, hashlib.sha1).hexdigest()
		if not hmac.compare_digest(signature.encode('UTF-8'), expected):
			return json_err("Invalid signature", 401)
		travis = requests.post("https://api.travis-ci.org/repo/openteacher%2Fopenteacher-ci/requests", json={
			"request": {
				"branch": "master",
				"message": "Build triggered by Launchpad branch update."
			},
		}, headers={
			"User-Agent": "OpenTeacher web services server",
			"Content-Type": "application/json",
			"Accept": "application/json",
			"Travis-API-Version": "3",
			"Authorization": "token %s" % app.config["TRAVIS_TOKEN"],
		})
		appveyor = requests.post("https://ci.appveyor.com/api/builds", json={
			"accountName": "openteacher",
			"projectSlug": "openteacher-ci",
			"branch": "master",
		}, headers={
			"Authorization": "Bearer %s" % app.config["APPVEYOR_TOKEN"],
		})
		return jsonify({
			"travis": travis.json(),
			"appveyor": appveyor.json(),
		})
