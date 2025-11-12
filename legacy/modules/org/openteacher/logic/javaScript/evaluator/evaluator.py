#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2013, Marten de Vries
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

class JSEvaluatorModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(JSEvaluatorModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "javaScriptEvaluator"
		self.requires = (
			self._mm.mods(type="qtApp"),
		)

	def createEvaluator(self):
		__doc__ = "Returns a:\n\n" + jseval.JSEvaluator.__doc__
		return jseval.JSEvaluator()

	def enable(self):
		global jseval
		try:
			import jseval
		except ImportError:
			return
		self.active = True

	def disable(self):
		self.active = False

def init(moduleManager):
	return JSEvaluatorModule(moduleManager)
