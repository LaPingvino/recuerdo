#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2013, 2017, Marten de Vries
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

SEGMENTS_RE = re.compile(r"(?<!\\)[0-9]+\. ")
SEPARATOR_RE = re.compile(r"(?<!\\)[,;]")

class WordsStringParserModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WordsStringParserModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "wordsStringParser"
		self.priorities = {
			"default": 10,
		}

	def parse(self, text):
		"""Parses a string into a questions or answers list as used
		   internally by OpenTeacher. See for examples the unit tests.

		"""
		obligatorySegments = SEGMENTS_RE.split(text)
		obligatorySegments = [x.strip() for x in obligatorySegments]
		if obligatorySegments[0]:
			# https://bugs.launchpad.net/openteacher/+bug/1233809
			obligatorySegments = [text]
		obligatorySegments = [x for x in obligatorySegments if x != u""]
		item = []
		for segment in obligatorySegments:
			words = SEPARATOR_RE.split(segment)
			words = [word.strip() for word in words]
			words = [word for word in words if word != u""]
			if words:
				item.append(words)
		return item

	def enable(self):
		self.active = True

	def disable(self):
		self.active = False

def init(moduleManager):
	return WordsStringParserModule(moduleManager)
