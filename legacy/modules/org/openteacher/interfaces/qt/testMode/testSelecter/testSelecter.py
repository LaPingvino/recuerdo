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
import weakref
import json

def installQtClasses():
	global TestSelecter

	class TestSelecter(QtWidgets.QListWidget):
		# Parameter: The current test (tests/<id>)
		testChosen = QtCore.pyqtSignal(dict)
		def __init__(self, connection, filter=False, *args, **kwargs):
			super(TestSelecter, self).__init__(*args, **kwargs)

			self.connection = connection
			# Are tests you already handed in answers for filtered out?
			self.filter = filter
			self.currentRowChanged.connect(self._currentRowChanged)

			#retranslate fills the list
			self.retranslate()

		def _addTests(self):
			# Get all tests
			self.testsInfo = self.connection.get("tests")
			self.testsInfos = []

			for test in self.testsInfo:
				if self.filter:
					# Get urls of all answers for this test
					answersUrls = self.connection.get(test["url"] + "/answers")
					# Get userids of all answers for this test
					answersIds = map(os.path.basename, answersUrls)

					if str(self.connection.userId) in answersIds:
						break

				# Get name of this test
				testInfo = self.connection.get(test["url"])
				testInfo["list"] = json.loads(testInfo["list"])
				self.testsInfos.append(testInfo)

				if "title" in testInfo["list"]:
					self.addItem(testInfo["list"]["title"])
				else:
					self.addItem(_("Unnamed"))

		def retranslate(self):
			self.clear()
			self._addTests()

		def _currentRowChanged(self, index):
			self.testChosen.emit(self.testsInfos[index])

		def getCurrentTest(self):
			return self.testsInfos[self.currentRow()]
	return TestSelecter

class TestModeTestSelecterModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModeTestSelecterModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "testModeTestSelecter"
		self.priorities = {
			"default": 532,
		}

		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.requires = (
			self._mm.mods(type="ui"),
			self._mm.mods(type="event"),
			self._mm.mods(type="testModeConnection"),
		)
		self.filesWithTranslations = ("testSelecter.py",)

	def enable(self):
		global QtCore, QtWidgets
		try:
			from PyQt5 import QtCore, QtWidgets
		except ImportError:
			return
		installQtClasses()

		self._modules = set(self._mm.mods(type="modules")).pop()
		self.connection = self._modules.default("active", type="testModeConnection").getConnection()
		self._selecters = set()

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

		for ref in self._selecters:
			selecter = ref()
			if selecter is not None:
				selecter.retranslate()

	def getTestSelecter(self, filter=False):
		ts = TestSelecter(self.connection, filter)
		self._selecters.add(weakref.ref(ts))
		return ts

	def disable(self):
		self.active = False

		del self._modules
		del self.connection
		del self._selecters

def init(moduleManager):
	return TestModeTestSelecterModule(moduleManager)
