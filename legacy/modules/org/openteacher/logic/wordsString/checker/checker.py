#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2013, Marten de Vries
#	Copyright 2011, Cas Widdershoven
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

class WordsStringCheckerModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WordsStringCheckerModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "wordsStringChecker"
		self.priorities = {
			"default": 10,
		}

	def _checkSingleCompulsoryAnswerGiven(self, givenAnswer, word):
		"""Called when the user only comma separates answers (there
		   might be compulsory answers among them.)

		"""
		result = {"result": "right"}
		difference = set(givenAnswer[0])
		for compulsoryAnswer in word["answers"]:
			oldDifference = difference.copy()
			difference -= set(compulsoryAnswer)
			if oldDifference == difference:
				result = {"result": "wrong"}
				break
		if result["result"] == "right" and len(difference) != 0:
			result = {"result": "wrong"}
		return result

	def _checkMultipleCompulsoryAnswersGiven(self, givenAnswer, word):
		"""The normal case: checks if enough compulsory answers are
		   given.

		"""
		result = {"result": "wrong"}
		compulsoryAnswerCount = 0
		for compulsoryGivenAnswer in givenAnswer:
			for compulsoryAnswer in word["answers"]:
				difference = set(compulsoryGivenAnswer) - set(compulsoryAnswer)
				if len(difference) == 0:
					compulsoryAnswerCount += 1

		if compulsoryAnswerCount == len(word["answers"]):
			result = {"result": "right"}
		return result

	def check(self, givenAnswer, word):
		"""Checks if the ``givenAnswer`` (the answer of the user) is
		   correct given the correct ``word``. It uses two strategies to
		   determine if an answer is correct:
		   - one that works if the user doesn't use compulsory answers
		   - one that works if the user does.

		   See for examples the unit tests.
		"""
		if len(givenAnswer) == 1:
			result = self._checkSingleCompulsoryAnswerGiven(givenAnswer, word)
		else:
			result = self._checkMultipleCompulsoryAnswersGiven(givenAnswer, word)

		result["itemId"] = word["id"]

		return result

	def enable(self):
		self._modules = set(self._mm.mods(type="modules")).pop()

		self.active = True

	def disable(self):
		self.active = False

		del self._modules

def init(moduleManager):
	return WordsStringCheckerModule(moduleManager)
