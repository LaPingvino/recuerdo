#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2012, Marten de Vries
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

import re
import datetime

class WordListStringParserModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WordListStringParserModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "wordListStringParser"
		self.requires = (
			self._mm.mods(type="wordsStringParser"),
		)
		self.priorities = {
			"default": 10,
		}

	def parseList(self, text, parseLenient=False):
		list = {
			"items": [],
			"tests": [],
		}
		now = datetime.datetime.now()

		counter = 0
		for line in text.split("\n"):
			if line.strip() == u"":
				continue
			try:
				questionText, answerText = re.split(r"(?<!\\)[=\t]", line, maxsplit=1)
			except ValueError:
				if parseLenient:
					continue
				raise ValueError("Missing equals sign or tab")

			word = {
				"id": counter,
				"questions": self._parse(questionText),
				"answers": self._parse(answerText),
				"created": now,
			}
			list["items"].append(word)

			counter += 1
		return {
			"resources": {},
			"list": list,
		}

	def enable(self):
		modules = set(self._mm.mods(type="modules")).pop()
		self._parse = modules.default("active", type="wordsStringParser").parse

		self.active = True

	def disable(self):
		self.active = False

		del self._parse

def init(moduleManager):
	return WordListStringParserModule(moduleManager)
