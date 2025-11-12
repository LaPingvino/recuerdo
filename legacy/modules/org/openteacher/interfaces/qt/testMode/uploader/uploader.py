#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2012, Milan Boers
#	Copyright 2012, Marten de Vries
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
import uuid
import superjson

class TestModeUploaderModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModeUploaderModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager
		
		self.type = "testModeUploader"
		x = 560
		self.priorities = {
			"all": x,
			"teacher": x,
			"code-documentation": x,
			"test-suite": x,
			"default": -1,
		}
		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.requires = (
			self._mm.mods(type="fileDialogs"),
			self._mm.mods(type="testModeConnection"),
			self._mm.mods(type="testMenu"),
			self._mm.mods(type="load", loads={"otwd": ["words"]})
		)
		self.filesWithTranslations = ("uploader.py",)

	def enable(self):
		self._modules = set(self._mm.mods(type="modules")).pop()
		self._fileDialogs = self._modules.default("active", type="fileDialogs")

		self._testMenu = self._modules.default("active", type="testMenu").menu

		self._action = self._testMenu.addAction(self.priorities["default"])
		self._action.triggered.handle(self.upload)

		#setup translation
		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

	def _retranslate(self):
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

		self._action.text = _("Upload lesson")

	def disable(self):
		self.active = False

		del self._modules
		del self._fileDialogs
		del self._testMenu
		del self._action
	
	def upload(self):
		# First, login
		self.connection = self._modules.default("active", type="testModeConnection").getConnection()
		self.connection.loggedIn.handle(self.upload_)
		
		self.loginid = uuid.uuid4()
		self.connection.login(self.loginid)
	
	def upload_(self, loginid):
		# Check if this is indeed from the request I sent out
		if loginid == self.loginid:
			loadModule = self._modules.default("active", type="load", loads={"otwd": ["words"]})

			# Get filename
			result = self._fileDialogs.getLoadPath(os.path.expanduser("~"), [("otwd", loadModule.name)])

			# Load the list
			list = loadModule.load(result)["list"]

			# Save it to a json string
			listJson = superjson.dumps(
				list, #the list to save
				separators=(',',':'), #compact encoding
				default=self._serialize
			)
			
			fb = self.connection.post("tests",{"list": listJson})
			print fb

	def _serialize(self, obj):
		try:
			return obj.strftime("%Y-%m-%dT%H:%M:%S.%f")
		except AttributeError:
			raise TypeError("The type '%s' isn't JSON serializable." % obj.__class__)

def init(moduleManager):
	return TestModeUploaderModule(moduleManager)
