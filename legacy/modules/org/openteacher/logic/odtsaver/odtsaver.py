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

class OdtSaverModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(OdtSaverModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "odtSaver"

		self.requires = (
			self._mm.mods(type="htmlGenerator", dataType="words"),
			self._mm.mods(type="ui"),
		)

	def save(self, lesson, path):
		html = self._modules.default(
			"active",
			type="htmlGenerator",
			dataType="words",
		).generate(lesson, margin="0.5em", coloredRows=False)

		doc = QtGui.QTextDocument()
		doc.setHtml(html)

		#odf -> OpenDocument Format
		writer = QtGui.QTextDocumentWriter(path, "odf")
		writer.write(doc)

	def enable(self):
		global QtGui
		try:
			from PyQt5 import QtGui
		except ImportError:
			return
		self._modules = set(self._mm.mods(type="modules")).pop()

		self.active = True

	def disable(self):
		self.active = False

		del self._modules

def init(moduleManager):
	return OdtSaverModule(moduleManager)
