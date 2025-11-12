#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2017, Marten de Vries
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

class LaTeXSaverModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(LaTeXSaverModule, self).__init__(*args, **kwargs)

		self._mm = moduleManager
		self.type = "save"
		self.priorities = {
			"default": 783,
		}
		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.requires = (
			self._mm.mods(type="wordsStringComposer"),
			self._mm.mods(type="metadata"),
		)
		self.filesWithTranslations = ("latex.py",)

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
			#TRANSLATORS: LaTeX is the name of a document preparation or
			#TRANSLATORS: typesetting system
			"tex": _("LaTeX"),
		}}

	def enable(self):
		global pyratemp
		try:
			import pyratemp
			import latexcodec
		except ImportError: # pragma: no cover
			return #remain inactive
		self._modules = set(self._mm.mods(type="modules")).pop()

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

	def disable(self):
		self.active = False

		del self._modules
		del self.saves

	def _compose(self, text):
		return self._modules.default(
			"active",
			type="wordsStringComposer"
		).compose(text).encode('latex', errors='replace')

	def save(self, type, lesson, path):
		class EvalPseudoSandbox(pyratemp.EvalPseudoSandbox):
			def __init__(self2, *args, **kwargs):
				pyratemp.EvalPseudoSandbox.__init__(self2, *args, **kwargs)

				self2.register("compose", self._compose)

		templatePath = self._mm.resourcePath("template.tex")
		t = pyratemp.Template(
			open(templatePath).read(),
			eval_class=EvalPseudoSandbox
		)
		content = t(list=lesson.list)
		with open(path, "w") as f:
			f.write(content.encode("UTF-8"))
		lesson.path = None

def init(moduleManager):
	return LaTeXSaverModule(moduleManager)
