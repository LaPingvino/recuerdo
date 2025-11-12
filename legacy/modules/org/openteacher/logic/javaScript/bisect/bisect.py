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

class JSBisectModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(JSBisectModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "bisectfunc"
		self.javaScriptImplementation = True
		self.requires = (
			self._mm.mods(type="javaScriptEvaluator"),
		)

	bisect = property(lambda self: self._js.global_["bisect"])

	def enable(self):
		with open(self._mm.resourcePath("bisect.js")) as f:
			self.code = f.read()

		modules = next(iter(self._mm.mods(type="modules")))
		self._js = modules.default("active", type="javaScriptEvaluator").createEvaluator()
		self._js.eval(self.code)

		self.active = True

	def disable(self):
		self.active = False

		del self.code
		del self._js

def init(moduleManager):
	return JSBisectModule(moduleManager)
