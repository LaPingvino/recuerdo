#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011, Milan Boers
#	Copyright 2011-2012, 2014, Marten de Vries
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

class PdfSaverModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(PdfSaverModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "save"
		self.priorities = {
			"default": 784,
		}

		self.requires = (
			self._mm.mods(type='print', dataType='words'),
		)
		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.filesWithTranslations = ("pdf.py",)

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
			#TRANSLATORS: This is the name of a document exchange standard.
			#TRANSLATORS: Please just use the English name of it, unless the
			#TRANSLATORS: standard is known under another name in your
			#TRANSLATORS: language (or you have a very good reason yourself
			#TRANSLATORS: for translating it). For more information on PDF:
			#TRANSLATORS: http://en.wikipedia.org/wiki/PDF
			"pdf": _("Portable Document Format"),
		}}

	def enable(self):
		global QtPrintSupport, QtWebKitWidgets
		try:
			from PyQt5 import QtPrintSupport, QtWebKitWidgets
		except ImportError:
			return
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

	def save(self, type, lesson, path):
		printer = QtPrintSupport.QPrinter()
		printer.setPageMargins(25, 25, 25, 25, QtPrintSupport.QPrinter.Millimeter)
		printer.setOutputFileName(path)
		printer.setOutputFormat(QtPrintSupport.QPrinter.PdfFormat)

		self._print(type, lesson, printer)
		lesson.path = None

	@property
	def _print(self):
		return self._modules.default(type='print', dataType='words').print_

def init(moduleManager):
	return PdfSaverModule(moduleManager)
