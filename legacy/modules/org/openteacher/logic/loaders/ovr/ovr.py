#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2014, Marten de Vries
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

import itertools

class OverhoringsprogrammaTalenLoaderModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(OverhoringsprogrammaTalenLoaderModule, self).__init__(*args, **kwargs)
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
		self.filesWithTranslations = ("ovr.py",)

	@property
	def _parse(self):
		return self._modules.default("active", type="wordsStringParser").parse

	def _retranslate(self):
		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			_, ngettext = unicode, lambda a, b, n: a if n == 1 else b
		else:
			_, ngettext = translator.gettextFunctions(
				self._mm.resourcePath("translations")
			)
		#TRANSLATORS: This is the name of a file type OpenTeacher can
		#TRANSLATORS: read. It's named after the program with the same
		#TRANSLATORS: name. For more info on the program, see:
		#TRANSLATORS: http://aaronweb.net/projects/overhoor/ (Dutch)
		self.name = _("Overhoringsprogramma Talen")

	def enable(self):
		self._modules = set(self._mm.mods(type="modules")).pop()
		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.mimetype = "application/x-overhoringsprogrammatalen"
		self.loads = {"ovr": ["words"]}

		self.active = True

	def disable(self):
		self.active = False

		del self._modules
		del self.name
		del self.loads
		del self.mimetype

	def getFileTypeOf(self, path):
		if path.endswith(".ovr"):
			return "words"

	def _readLine(self, f):
		data = next(f)
		return unicode(data, encoding="iso-8859-1").strip()

	def _skipLine(self, f):
		next(f)

	def load(self, path):
		"""Loads 'Overhoringsprogramma Talen' (*.ovr) lists. This parser
		   is based on inspection of files, not on official
		   documentation.

		"""

		with open(path, "r") as f:
			#do something with f

			list = {
				"questionLanguage": self._readLine(f),
				"answerLanguage": self._readLine(f),
			}
			for i in range(4):
				#no idea what these lines represent
				self._skipLine(f)

			list["items"] = self._loadItems(f)

		return {
			"resources": {},
			"list": list,
		}

	def _loadItems(self, f):
		items = []
		counter = itertools.count()
		while True:
			try:
				questions = self._readLine(f)
				answers = []
				#load answers untill the stop sequence is loaded
				while answers[-2:] != [u"-", u"0"]:
					answers.append(self._readLine(f))
				#cut off the stop sequence
				answers = answers[:-2]

				answers = [tuple(answers)]
			except StopIteration:
				break

			items.append({
				"id": next(counter),
				"questions": self._parse(questions),
				"answers": answers,
			})

		return items

def init(moduleManager):
	return OverhoringsprogrammaTalenLoaderModule(moduleManager)
