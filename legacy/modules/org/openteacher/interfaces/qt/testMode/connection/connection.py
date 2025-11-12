#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2012, Milan Boers
#	Copyright 2012-2013, Marten de Vries
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

import os
import urllib
import urllib2
import socket
import json

MAXVERSION = 1.0

def getConnectWidget():
	class ConnectWidget(QtWidgets.QWidget):
		"""We use HTTP so there's not actually a continuous connection, but
		   this class is basically here to check whether a server exists.

		"""
		def __init__(self, connection, *args, **kwargs):
			super(ConnectWidget, self).__init__(*args, **kwargs)

			self.connection = connection

			self.connectLayout = QtWidgets.QFormLayout()
			self.setLayout(self.connectLayout)

			self.serverField = QtWidgets.QLineEdit()
			self.serverLabel = QtWidgets.QLabel()
			self.connectLayout.addRow(self.serverLabel, self.serverField)

			self.connectButton = QtWidgets.QPushButton()
			self.connectButton.clicked.connect(lambda: self.connection.connect(str(self.serverField.text())))
			self.connectLayout.addRow(self.connectButton)

			self.serverField.returnPressed.connect(self.connectButton.click)

			self.retranslate()

		def retranslate(self):
			self.serverLabel.setText(_("Server IP or hostname:"))
			self.connectButton.setText(_("Connect"))
	return ConnectWidget

def getLoginWidget():
	class LoginWidget(QtWidgets.QWidget):
		def __init__(self, connection, loginid, *args, **kwargs):
			super(LoginWidget, self).__init__(*args, **kwargs)

			self.connection = connection

			self.loginLayout = QtWidgets.QFormLayout()
			self.setLayout(self.loginLayout)

			self.usernameLabel = QtWidgets.QLabel()
			self.usernameField = QtWidgets.QLineEdit()
			self.loginLayout.addRow(self.usernameLabel, self.usernameField)

			self.passwordLabel = QtWidgets.QLabel()
			self.passwordField = QtWidgets.QLineEdit()
			self.passwordField.setEchoMode(QtWidgets.QLineEdit.Password)
			self.loginLayout.addRow(self.passwordLabel, self.passwordField)

			self.checkButton = QtWidgets.QPushButton()
			self.checkButton.clicked.connect(lambda: connection.checkLogin(
				str(self.usernameField.text()),
				str(self.passwordField.text()),
				loginid
			))
			self.loginLayout.addRow(self.checkButton)

			self.passwordField.returnPressed.connect(self.checkButton.click)

		def retranslate(self):
			self.usernameLabel.setText(_("Username:"))
			self.passwordLabel.setText(_("Password:"))
			self.checkButton.setText(_("Login"))
	return LoginWidget

def getConnectLoginWidget():
	class ConnectLoginWidget(QtWidgets.QWidget):
		def __init__(self, connection, loginid, *args, **kwargs):
			super(ConnectLoginWidget, self).__init__(*args, **kwargs)

			self.layout = QtWidgets.QStackedLayout()
			self.setLayout(self.layout)

			self.connectWidget = ConnectWidget(connection)
			connection.connected.handle(self.afterConnect)
			self.layout.addWidget(self.connectWidget)

			self.loginWidget = LoginWidget(connection, loginid)
			self.layout.addWidget(self.loginWidget)

			self.retranslate()

		def afterConnect(self):
			self.layout.setCurrentWidget(self.loginWidget)

		def retranslate(self):
			self.connectWidget.retranslate()
			self.loginWidget.retranslate()
	return ConnectLoginWidget

class Connection(object):
	def __init__(self, modules, *args, **kwargs):
		super(Connection, self).__init__(*args, **kwargs)

		self._modules = modules

		# Connected event
		self.connected = self._modules.default(type="event").createEvent()
		# LoggedIn event
		self.loggedIn = self._modules.default(type="event").createEvent()

		# Setup opener
		self.opener = urllib2.build_opener()
		self.server = None
		self.serverName = None
		self.auth = False

	def retranslate(self):
		if hasattr(self, "loginTab"):
			self.connectLoginWidget.retranslate()
			self.loginTab.title = _("Login")

	@property
	def connectedLoggedIn(self):
		"""'Connected' to the server and logged in?"""

		return self.server != None and self.auth == True

	def connect(self, hostname):
		"""'Connect' to server"""

		try:
			assert hostname
			#Replace hostname by IP (so DNS is not needed at every
			#request. Will speed things up.)
			hostname = socket.gethostbyname(hostname)
		except (socket.gaierror, AssertionError):
			# Could not connect
			dialogShower = self._modules.default(type="dialogShower")
			dialogShower.showError.send(self.loginTab, _("Could not connect to the server. Possibly wrong hostname."))
		else:
			print(hostname)
			# Everything OK, Connected
			self.server = hostname
			# Try to fetch the index page
			index = self.get("https://%s:8080/" % (hostname))
			self.serverName = index["name"]

			self.connected.send()

	def get(self, path):
		"""Get path"""

		try:
			req = json.load(self.opener.open("https://%s:8080/%s" % (self.server, path)))
			return req
		except urllib2.HTTPError, e:
			return e

	def post(self, path, data):
		data = urllib.urlencode(data)
		try:
			request = urllib2.Request("https://%s:8080/%s" % (self.server, path), data)
			return json.load(self.opener.open(request))
		except urllib2.HTTPError, e:
			return e

	def put(self, path, data):
		data = urllib.urlencode(data)
		try:
			request = urllib2.Request("https://%s:8080/%s" % (self.server, path), data)
			request.get_method = lambda: 'PUT'
			return json.load(self.opener.open(request))
		except urllib2.HTTPError, e:
			return e

	def login(self, loginid):
		"""Method to log in. You need to send a uuid along, so your
		   loggedIn doesn't react to other requests.

		"""
		if self.connectedLoggedIn:
			self._afterLogin(loginid)
		else:
			self.connectLoginWidget = ConnectLoginWidget(self, loginid)

			module = self._modules.default("active", type="ui")
			self.loginTab = module.addCustomTab(self.connectLoginWidget)
			#set tab title by retranslating
			self.retranslate()

			self.loginTab.closeRequested.handle(self.loginTab.close)

	def checkLogin(self, username, passwd, loginid):
		"""Checks if login is right (and implicitly fetches the token)
		   (to be used only inside this module)

		"""
		index = "https://%s:8080/" % self.server

		loginCreds = u"%s:%s" % (username, passwd)
		loginCreds = str(loginCreds.encode("utf-8"))
		loginCreds = loginCreds.encode("base64").strip()

		self.opener.addheaders = [("Authorization", "Basic %s" % loginCreds)]

		me = self.get("users/me")

		if type(me) == urllib2.HTTPError:
			# User was not logged in
			dialogShower = self._modules.default("active", type="dialogShower")
			dialogShower.showError.send(self.loginTab, _("Could not login. Wrong username or password."))
		else:
			self.userId = int(os.path.basename(me))
			self._afterLogin(loginid)

	def _afterLogin(self, loginid):
		"""Logged in, set var to server"""

		self.auth = True
		self.loginTab.close()
		self.loggedIn.send(loginid)

class TestModeConnectionModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModeConnectionModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "testModeConnection"
		self.priorities = {
			"default": 444,
		}

		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.requires = (
			self._mm.mods(type="event"),
			self._mm.mods(type="ui"),
		)
		self.filesWithTranslations = ("connection.py",)

	def _retranslate(self):
		#Translations
		global _
		global ngettext

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			_, ngettext = unicode, lambda a, b, n: a if n == 1 else b
		else:
			_, ngettext = translator.gettextFunctions(
				self._mm.resourcePath("translations")
			)

		self.connection.retranslate()

	def enable(self):
		global QtWidgets
		try:
			from PyQt5 import QtWidgets
		except ImportError:
			return
		global ConnectLoginWidget, ConnectWidget, LoginWidget
		ConnectLoginWidget = getConnectLoginWidget()
		ConnectWidget = getConnectWidget()
		LoginWidget = getLoginWidget()

		self._modules = set(self._mm.mods(type="modules")).pop()

		self.connection = Connection(self._modules)

		#setup translation
		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

	def disable(self):
		self.active = False

		del self.connection
		del self._modules

	def getConnection(self):
		return self.connection

def init(moduleManager):
	return TestModeConnectionModule(moduleManager)
