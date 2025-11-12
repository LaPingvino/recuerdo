#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011, 2013, Marten de Vries
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

class PercentsCalculatorModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(PercentsCalculatorModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "percentsCalculator"
		self.requires = (
			self._mm.mods(type="javaScriptEvaluator"),
			self._mm.mods("javaScriptImplementation", type="mapfunc"),
			self._mm.mods("javaScriptImplementation", type="sumfunc"),
		)
		self.javaScriptImplementation = True

	def enable(self):
		modules = next(iter(self._mm.mods(type="modules")))
		self._js = modules.default("active", type="javaScriptEvaluator").createEvaluator()
		self.code = modules.default("active", "javaScriptImplementation", type="mapfunc").code
		self.code += modules.default("active", "javaScriptImplementation", type="sumfunc").code
		with open(self._mm.resourcePath("percentsCalculator.js")) as f:
			self.code += f.read()
		self._js.eval(self.code)

		self.active = True

	def calculateAveragePercents(self, tests):
		"""Calculates the average score of all ``tests`` in percents"""

		return self._js.global_["calculateAveragePercents"](tests)

	def calculatePercents(self, test):
		"""Calculates the score of the user in the passed-in ``test``
		   in percents.

		"""
		return self._js.global_["calculatePercents"](test)

	def disable(self):
		self.active = False

		del self._js
		del self.code

def init(moduleManager):
	return PercentsCalculatorModule(moduleManager)
