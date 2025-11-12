#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2013, Marten de Vries
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

from etree import ElementTree
import itertools

class FmdLoaderModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(FmdLoaderModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "load"
		self.priorities = {
			"default": 756,
		}
		
		self.requires = (
			self._mm.mods(type="wordsStringParser"),
		)
		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.filesWithTranslations = ("fmd.py",)

	def _retranslate(self):
		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			_, ngettext = unicode, lambda a, b, n: a if n == 1 else b
		else:
			_, ngettext = translator.gettextFunctions(
				self._mm.resourcePath("translations")
			)
		#TRANSLATORS: Fresh Memory is the name of program similar to
		#TRANSLATORS: OpenTeacher of which we can import its files.
		#TRANSLATORS: See also: http://freshmemory.sourceforge.net/
		self.name = _("Fresh Memory")

	def enable(self):
		self._modules = set(self._mm.mods(type="modules")).pop()
		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.mimetype = "application/x-fm-dictionary"
		self.loads = {"fmd": ["words"]}

		self.active = True

	def disable(self):
		self.active = False

		del self._modules
		del self.name
		del self.loads
		del self.mimetype

	def getFileTypeOf(self, path):
		if path.endswith(".fmd"):
			return "words"

	def load(self, path):
		wsp = self._modules.default("active", type="wordsStringParser")

		with open(path) as f:
			root = ElementTree.parse(f).getroot()

		wordList = {
			"items": [],
		}

		wordList["questionLanguage"] = root.findtext("fields/field[1]") or u""
		wordList["answerLanguage"] = root.findtext("fields/field[2]") or u""

		#counter is used as word id
		counter = itertools.count()

		for entryTree in root.findall("entries/e"):
			wordList["items"].append({
				"id": next(counter),
				"questions": wsp.parse(entryTree.findtext("f[1]") or u""),
				"answers": wsp.parse(entryTree.findtext("f[2]")  or u""),
				"comment": entryTree.findtext("f[3]")  or u"",
			})

		return {
			"resources": {},
			"list": wordList,
		}

def init(moduleManager):
	return FmdLoaderModule(moduleManager)
