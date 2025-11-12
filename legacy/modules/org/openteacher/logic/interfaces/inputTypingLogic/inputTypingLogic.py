#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2013, Marten de Vries
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

import datetime
import contextlib

class Controller(object):
	def __init__(self, Event, parse, compose, check, *args, **kwargs):
		super(Controller, self).__init__(*args, **kwargs)

		self._parse = parse
		self._compose = compose
		self._check = check

		self._showingCorrection = False

		self._installEvents(Event)

	@property
	def lessonType(self):
		return self._lessonType

	@lessonType.setter
	def lessonType(self, v):
		with contextlib.suppress(AttributeError):
			self._lessonType.newItem.unhandle(self._onNewWord)
		if self._showingCorrection:
			self.correctionShowingDone()

		self._lessonType = v

		self._lessonType.newItem.handle(self._onNewWord)
		self.disableCorrectAnyway.send()

	def _installEvents(self, Event):
		self.clearInput = Event()
		self.enableInput = Event()
		self.disableInput = Event()
		self.focusInput = Event()

		self.showCorrection = Event()
		self.hideCorrection = Event()

		self.enableCheck = Event()
		self.disableCheck = Event()
		self.enableSkip = Event()
		self.disableSkip = Event()
		self.enableCorrectAnyway = Event()
		self.disableCorrectAnyway = Event()

	def _onNewWord(self, word):
		self._lastWord = word

		self._activity = {"start": datetime.datetime.now()}

		self.clearInput.send()
		self.focusInput.send()

	def _enableUi(self):
		self.enableCheck.send()
		self.enableSkip.send()
		self.enableInput.send()

	def _disableUi(self):
		self.disableCheck.send()
		self.disableSkip.send()
		self.disableInput.send()

	def checkTriggered(self, inputContent):
		self._assertLessonTypeSet()
		self._assertLessonActive()
		self._assertNotShowingCorrection()

		result = self._checkAnswerAndSaveResult(inputContent)
		if result == "wrong":
			self._resultWrong()
		else:
			self._resultRight()

	def _assertLessonTypeSet(self):
		try:
			self.lessonType
		except AttributeError:
			#just so it's clear what this function's doing.
			raise

	def _assertLessonActive(self):
		try:
			self._lastWord
		except AttributeError:
			raise ValueError("No lesson active")

	def userIsTyping(self):
		self._assertLessonTypeSet()
		self._assertLessonActive()
		self._assertNotShowingCorrection()

		self._activity["end"] = datetime.datetime.now()

	def _resultWrong(self):
		self._disableUi()
		self.enableCorrectAnyway.send()

		self._showingCorrection = True
		correctAnswer = self._compose(self._lastWord["answers"])
		self.showCorrection.send(correctAnswer)

	def _resultRight(self):
		self.disableCorrectAnyway.send()
		self._tellLessonTypeAboutTheResult()

	def _checkAnswerAndSaveResult(self, userInput):
		givenAnswer = self._parse(userInput)

		self._lastResult = self._check(givenAnswer, self._lastWord)
		self._lastResult.update({
			"givenAnswer": userInput,
		})
		return self._lastResult["result"]

	def _assertNotShowingCorrection(self):
		if self._showingCorrection:
			raise ValueError("Showing a correction -> calling this now makes no sense")

	def _tellLessonTypeAboutTheResult(self):
		if not "end" in self._activity:
			#assume now
			self.userIsTyping()
		self._lastResult["active"] = self._activity
		self.lessonType.setResult(self._lastResult)

	def correctionShowingDone(self):
		self._assertShowingCorrection()
		self._showingCorrection = False

		self.hideCorrection.send()
		self._tellLessonTypeAboutTheResult()
		self._enableUi()

	def _assertShowingCorrection(self):
		if not self._showingCorrection:
			raise ValueError("Not showing a correction!")

	def skipTriggered(self):
		self._assertLessonTypeSet()
		self._assertLessonActive()
		self._assertNotShowingCorrection()

		self.lessonType.skip()

	def correctAnywayTriggered(self):
		self._assertLessonTypeSet()
		self._assertLessonActive()

		if self._showingCorrection:
			self.correctionShowingDone()
		self._lastResult.update({
			"result": "right",
			"givenAnswer": _("Corrected: %s") % self._lastResult["givenAnswer"],
		})
		self.lessonType.correctLastAnswer(self._lastResult)
		self.disableCorrectAnyway.send()

class InputTypingLogicModule(object):
	"""This module offers an object that can be used to control the part
	   of a GUI where the user types his/her answer in in a test.

	"""
	def __init__(self, moduleManager, *args, **kwargs):
		super(InputTypingLogicModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "inputTypingLogic"
		self.requires = (
			self._mm.mods(type="event"),
			self._mm.mods(type="wordsStringComposer"),
			self._mm.mods(type="wordsStringParser"),
			self._mm.mods(type="wordsStringChecker"),
		)

	def createController(self, *args, **kwargs):
		return Controller(self._createEvent, self._parse, self._compose, self._check, *args, **kwargs)

	def enable(self):
		self._modules = next(iter(self._mm.mods(type="modules")))
		self._createEvent = self._modules.default(type="event").createEvent
		self._parse = self._modules.default("active", type="wordsStringParser").parse
		self._compose = self._modules.default("active", type="wordsStringComposer").compose
		self._check = self._modules.default("active", type="wordsStringChecker").check

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

	def _retranslate(self):
		#Install translator inside the whole of these file
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

	def disable(self):
		self.active = False

		del self._modules
		del self._createEvent
		del self._parse
		del self._compose
		del self._check

def init(moduleManager):
	return InputTypingLogicModule(moduleManager)
