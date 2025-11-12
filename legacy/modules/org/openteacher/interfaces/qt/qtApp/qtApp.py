#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2012, 2017, Marten de Vries
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

class QtAppModule(object):
	"""When this module is enabled, there is guaranteed to be a
	   QApplication instance. It **doesn't** guarantee that that
	   QApplication did initialize the GUI, though.

	"""
	def __init__(self, moduleManager, *args, **kwargs):
		super(QtAppModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "qtApp"
		self.uses = (
			self._mm.mods(type="ui"),
		)

	def enable(self):
		try:
			from PyQt5 import QtCore
		except ImportError:
			return
		if not QtCore.QCoreApplication.instance():
			#not handled by the ui module. That means that we need a
			#non-gui qapplication
			self._app = QtCore.QCoreApplication([])
		self.active = True

	def disable(self):
		self.active = False

		with contextlib.suppress(AttributeError):
			del self._app

def init(moduleManager):
	return QtAppModule(moduleManager)
