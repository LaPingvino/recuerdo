#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011, Milan Boers
#	Copyright 2011-2012, 2014 Marten de Vries
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

class SylkSaverModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(SylkSaverModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "save"
		self.priorities = {
			"default": 621,
		}
		
		self.requires = (
			self._mm.mods(type="sylkSaver"),
		)
		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.filesWithTranslations = ("sylk.py",)

	def _retranslate(self):
		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			_, ngettext = unicode, lambda a, b, n: a if n == 1 else b
		else:
			_, ngettext = translator.gettextFunctions(
				self._mm.resourcePath("translations")
			)
		self.saves = {"words": {
			#TRANSLATORS: This is one of the file formats OpenTeacher
			#TRANSLATORS: saves to. For more info, see:
			#TRANSLATORS: https://en.wikipedia.org/wiki/SYmbolic_LinK_%28SYLK%29
			"slk": _("SYLK"),
		}}

	def enable(self):		
		self._modules = set(self._mm.mods(type="modules")).pop()

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

	def save(self, type, lesson, path):
		self._modules.default("active", type="sylkSaver").save(lesson, path)

	def disable(self):
		self.active = False

		del self._modules
		del self.saves

def init(moduleManager):
	return SylkSaverModule(moduleManager)
