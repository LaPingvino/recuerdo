#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011, Milan Boers
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

# There are a couple of old (non-relevant) file formats that became broken over
# time. Probably some obscure LibreOffice bug (or bugs). Anyway, they could be
# re-enabled if those bugs ever get fixed.

import tempfile
import distutils.spawn
import os
import subprocess
import shutil
import logging

loformatsLogger = logging.getLogger("libreofficeformats")

class LibreofficeFormatsSaverModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(LibreofficeFormatsSaverModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "save"
		self.priorities = {
			"default": 621,
		}

		self.requires = (
			self._mm.mods(type="odtSaver"),
			self._mm.mods(type="sylkSaver"),
		)
		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.filesWithTranslations = ("libreofficeFormats.py",)

	def enable(self):
		if not distutils.spawn.find_executable("soffice"):# pragma: no cover
			#remain inactive
			return

		self._modules = set(self._mm.mods(type="modules")).pop()

		#Values are LibreOffice filters, see for more on them:
		#http://cgit.freedesktop.org/libreoffice/core/tree/filter/source/config/fragments/filters
		self._textFilters = {
			"doc": "MS Word 97",
			"rtf": "Rich Text Format",
			"docx": "Office Open XML Text",
			"uot": "UOF text",
#			"sxw": "StarOffice XML (Writer)",
		}
		self._spreadSheetFilters = {
			"ods": "calc8",
			"xls": "MS Excel 97",
			"xlsx": "Calc Office Open XML",
			"dif": "DIF",
#			"dbf": "dBase",
			"uos": "UOF spreadsheet",
#			"sxc": "StarOffice XML (Calc)",
		}
		self._filters = self._textFilters.copy()
		self._filters.update(self._spreadSheetFilters)

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

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
			#TRANSLATORS: This is a file format.
			"doc": _("Microsoft Word 97/2000/XP/2003"),
			#TRANSLATORS: This is a file format.
			"rtf": _("Rich Text Format"),
			#TRANSLATORS: This is a file format.
			"docx": _("Office Open XML Text"),
			#TRANSLATORS: This is a file format.
			"uot": _("Unified Office Format text"),
			#TRANSLATORS: This is a file format.
#			"sxw": _("StarOffice XML (Writer)"),
			#TRANSLATORS: This is a file format.
			"ods": _("OpenDocument Spreadsheet"),
			#TRANSLATORS: This is a file format.
			"xls": _("Microsoft Excel 97/2000/XP/2003"),
			#TRANSLATORS: This is a file format.
			"xlsx": _("Office Open XML Spreadsheet"),
			#TRANSLATORS: This is a file format.
			"dif": _("Data Interchange Format"),
			#TRANSLATORS: This is a file format.
#			"dbf": _("dBASE"),
			#TRANSLATORS: This is a file format.
			"uos": _("Unified Office Format spreadsheet"),
			#TRANSLATORS: This is a file format.
#			"sxc": _("StarOffice XML (Calc)"),
		}}

	def disable(self):
		self.active = False

		del self._modules
		del self.saves
		del self._textFilters
		del self._spreadSheetFilters
		del self._filters

	def _toInterimFile(self, tempDir, ext, lesson):
		if ext in self._textFilters:
			saver = self._modules.default("active", type="odtSaver")
			name = "document.odt"
		if ext in self._spreadSheetFilters:
			saver = self._modules.default("active", type="sylkSaver")
			name = "document.slk"
		saver.save(lesson, os.path.join(tempDir, name))
		return name

	def save(self, type, lesson, path):
		tempDir = tempfile.mkdtemp()
		ext = os.path.splitext(path)[1].strip(".")
		fileFormat = ext + ":" + self._filters[ext]

		name = self._toInterimFile(tempDir, ext, lesson)

		cwd = os.getcwd()
		os.chdir(tempDir)
		try:
			process = subprocess.Popen([
				"soffice",
				#to allow for multiple LO instances to run.
				"-env:UserInstallation=file://%s/lo-ui" % tempDir,
				"--headless",
				"--convert-to", fileFormat,
				name,
			], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
			loformatsLogger.debug("soffice output\n\tstdout: %s\n\tstderr: %s" % process.communicate())
		finally:
			os.chdir(cwd)

		try:
			shutil.move(os.path.join(tempDir, "document." + ext), path)
		finally:
			shutil.rmtree(tempDir)

		lesson.path = None

def init(moduleManager):
	return LibreofficeFormatsSaverModule(moduleManager)
