#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2012, Marten de Vries
#	Copyright 2011, Milan Boers
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

import json

class WordsHtmlGeneratorModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WordsHtmlGeneratorModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "htmlGenerator"
		self.dataType = "words"
		self.requires = (
			self._mm.mods(type="javaScriptEvaluator"),
			self._mm.mods("javaScriptImplementation", type="wordsStringComposer"),
			self._mm.mods(type="jsLib", name="tmpl"),
		)
		self.javaScriptImplementation = True

	def enable(self):
		global pyratemp
		try:
			import pyratemp
		except ImportError:
			return #remain inactive
		self._modules = set(self._mm.mods(type="modules")).pop()

		with open(self._mm.resourcePath("words.js")) as f:
			self.code = f.read()
		with open(self._mm.resourcePath("template.html")) as f:
			self.code += "var WORDS_HTML_GENERATOR_TEMPLATE = %s;\n\n" % json.dumps(f.read())
		self.code += self._modules.default("active", "javaScriptImplementation", type="wordsStringComposer").code
		self.code += self._modules.default("active", type="jsLib", name="tmpl").code

		self._js = self._modules.default("active", type="javaScriptEvaluator").createEvaluator()
		self._js.eval(self.code)

		self.active = True

	def disable(self):
		self.active = False

		del self._modules
		del self.code
		del self._js

	def generate(self, lesson, margin=None, coloredRows=None):
		"""Generates a html document which provides an overview of all
		   the questions and answers in `lesson`. It includes an inline
		   stylesheet.

		   - `margin` specifies the margin for the page (should be valid
		     css, e.g. 1em or 1px)
		   - `coloredRows` specifies if the odd rows should have a
		     different background colors than the even ones.

		"""
		return self._js.global_["generateWordsHtml"](lesson.list, margin=margin, coloredRows=coloredRows)

def init(moduleManager):
	return WordsHtmlGeneratorModule(moduleManager)
