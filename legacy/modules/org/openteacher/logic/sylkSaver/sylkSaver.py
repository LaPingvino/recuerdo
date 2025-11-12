#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2014, Marten de Vries
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

import contextlib
import itertools

class SylkSaverModule(object):
	"""Document format description:
	   https://en.wikipedia.org/wiki/SYmbolic_LinK_%28SYLK%29

	"""
	def __init__(self, moduleManager, *args, **kwargs):
		super(SylkSaverModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "sylkSaver"
		self.requires = (
			self._mm.mods(type="metadata"),
			self._mm.mods(type="wordsStringComposer"),
		)
		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.filesWithTranslations = ["sylkSaver.py",]

	def enable(self):
		#Translations
		self._modules = set(self._mm.mods(type="modules")).pop()
		self._metadata = self._modules.default("active", type="metadata").metadata

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

	def _retranslate(self):
		global _, ngettext
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
		del self._metadata

	@property
	def _compose(self):
		return self._modules.default(
			"active",
			type="wordsStringComposer"
		).compose

	def _rowRepresentation(self, columns):
		y = next(self._rowNumber)
		for x, column in enumerate(columns, start=1):
			yield u'C;X%s;Y%s;K"%s"' % (x, y, column)

	def save(self, lesson, path):
		self._rowNumber = itertools.count(1)

		programId = self._metadata["name"] + self._metadata["version"]
		#start of file
		rows = [u"ID;P" + programId]

		#title
		with contextlib.suppress(KeyError):
			titleRow = [lesson.list["title"]]
			emptyRow = []
			rows.extend(self._rowRepresentation(titleRow))
			rows.extend(self._rowRepresentation(emptyRow))

		#headers
		headerRow = [_("Questions"), _("Answers"), _("Comment"), _("Comment after answering")]
		rows.extend(self._rowRepresentation(headerRow))

		#items
		items = lesson.list.get("items", [])
		for item in items:
			row = [
				self._compose(item.get("questions", [])),
				self._compose(item.get("answers", [])),
				item.get("comment", u""),
				item.get("commentAfterAnswering", u"")
			]
			rows.extend(self._rowRepresentation(row))

		#footer
		rows.append(u"E")

		#save file
		data = u"\n".join(rows).encode("UTF-8")
		with open(path, "w") as f:
			f.write(data)

def init(moduleManager):
	return SylkSaverModule(moduleManager)
