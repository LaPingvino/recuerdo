#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2012-2013, 2017, Marten de Vries
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

class WordListStringComposerModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WordListStringComposerModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "wordListStringComposer"
		self.requires = (
			self._mm.mods(type="wordsStringComposer"),
		)

	def composeList(self, container):
		items = container['list'].get('items', [])
		result = []
		for item in items:
			questions = self._escape(self._compose(item.get('questions', [])))
			answers = self._escape(self._compose(item.get('answers', [])))
			result.append('%s = %s' % (questions, answers))
		result.append('')
		return '\n'.join(result) or '\n'

	def _escape(self, data):
		# (+ 1 because we want the '=', not the thing that isn't a slash)
		data = self._equalsRe.sub(r'\1\=', data)
		# same for tab
		data = self._tabRe.sub(r'\1\\t', data)

		return data

	def enable(self):
		modules = set(self._mm.mods(type="modules")).pop()

		self._compose = modules.default("active", type="wordsStringComposer").compose
		self._equalsRe = re.compile(r"([^\\])=")
		self._tabRe = re.compile(r"([^\\])\t")
		self.active = True

	def disable(self):
		self.active = False
		del self._tabRe
		del self._equalsRe
		del self._compose

def init(moduleManager):
	return WordListStringComposerModule(moduleManager)
