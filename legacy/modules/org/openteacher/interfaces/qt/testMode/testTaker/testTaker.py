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

import uuid
import urllib2
import superjson

def getTestChooser():
	class TestChooser(QtWidgets.QWidget):
		# Parameter: the chosen test (tests/<id>)
		testChosen = QtCore.pyqtSignal(dict)
		def __init__(self, testSelecter, *args, **kwargs):
			super(TestChooser, self).__init__(*args, **kwargs)

			self.testSelecter = testSelecter

			layout = QtWidgets.QVBoxLayout()

			self.comm = QtWidgets.QLabel()
			layout.addWidget(self.comm)

			layout.addWidget(testSelecter)

			self.button = QtWidgets.QPushButton()
			self.button.clicked.connect(lambda: self.testChosen.emit(self.testSelecter.getCurrentTest()))
			layout.addWidget(self.button)

			self.setLayout(layout)

		def retranslate(self):
			self.comm.setText(_("Select the test you want to take"))
			self.button.setText(_("Take test"))
	return TestChooser

class TestModeTestTakerModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModeTestTakerModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "testModeTestTaker"
		x = 504
		self.priorities = {
			"all": x,
			"student@school": x,
			"code-documentation": x,
			"default": -1,
		}

		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.requires = (
			self._mm.mods(type="ui"),
			self._mm.mods(type="event"),
			self._mm.mods(type="wordsTestTeacher"),
			self._mm.mods(type="testModeConnection"),
			self._mm.mods(type="testMenu"),
		)
		self.filesWithTranslations = ("testTaker.py",)

	def enable(self):
		global QtCore, QtWidgets
		try:
			from PyQt5 import QtCore, QtWidgets
		except ImportError:
			return
		global TestChooser
		TestChooser = getTestChooser()

		self._modules = set(self._mm.mods(type="modules")).pop()

		self._testMenu = self._modules.default("active", type="testMenu").menu

		self._action = self._testMenu.addAction(self.priorities["default"])
		self._action.triggered.handle(self.showTestTaker)

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

		self._action.text = _("Take test")
		if hasattr(self, "testChooser"):
			self.testChooser.retranslate()
		if hasattr(self, "testChooseTab"):
			self.testChooseTab.title = _("Choose test")
		if hasattr(self, "teachTab"):
			self.teachTab.title = _("Taking test")

	def disable(self):
		self.active = False

		self._action.remove()

		del self._modules
		del self._action
		del self._testMenu

	def showTestTaker(self):
		# First, login
		self.connection = self._modules.default("active", type="testModeConnection").getConnection()
		self.connection.loggedIn.handle(self.showTestTaker_)
		self.loginid = uuid.uuid4()
		self.connection.login(self.loginid)

	def showTestTaker_(self, loginid):
		# Check if this is indeed from the request I sent out
		if loginid == self.loginid:
			uiModule = self._modules.default("active", type="ui")

			#teachWidget = self._modules.default("active", type="wordsTestTeacher").createWordsTeacher()
			testSelecter = self._modules.default("active", type="testModeTestSelecter").getTestSelecter(True)
			self.testChooser = TestChooser(testSelecter)
			self.testChooser.testChosen.connect(self.takeTest)

			self.testChooseTab = uiModule.addCustomTab(self.testChooser)
			self.testChooseTab.closeRequested.handle(self.testChooseTab.close)
			self._retranslate()

	def takeTest(self, testInfo):
		"""testInfo: dictionary with info from /tests/<id>"""

		# Close the test chooser.
		self.testChooseTab.close()

		self.teachWidget = self._modules.default("active", type="wordsTestTeacher").createWordsTeacher()
		self.teachWidget.updateList(testInfo["list"])

		uiModule = self._modules.default("active", type="ui")

		self.teachTab = uiModule.addCustomTab(self.teachWidget)
		self.teachTab.closeRequested.handle(self.teachTab.close)
		self._retranslate()

		self.teachWidget.lessonDone.connect(lambda: self.handIn(testInfo["answers"]))

	def handIn(self, answersUrl):
		"""Hand in the test"""

		answeredList = self.teachWidget.getAnsweredList()
		dialogShower = self._modules.default("active", type="dialogShower")

		r = self.connection.post(answersUrl, {"list" : superjson.dumps(answeredList)})

		# Check if there was an error.
		if type(r) == urllib2.HTTPError:
			# Show error message
			dialogShower.showError.send(self.teachTab, _("An error occured. The server could not be reached. Check your network connection."))
		else:
			# Show thank you message
			dialogShower.showBigMessage.send(_("Your answers have successfully been handed in!"))
			# Close the teach widget
			self.teachTab.close()

def init(moduleManager):
	return TestModeTestTakerModule(moduleManager)
