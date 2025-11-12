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

import unittest

class TestModule(object):
	"""This module runs the test suite included in the jseval module.
	   (Which is currently developed alongside OT, but useful for other
	   programs too which is why the tests are separated.)

	"""
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "testSuite"
		self.requires = (
			self._mm.mods(type="javaScriptEvaluator"),
		)

	def enable(self):
		global jseval
		try:
			import jseval.tests
		except ImportError:
			return
		self.testSuite = unittest.defaultTestLoader.loadTestsFromModule(jseval.tests)
		self.active = True

	def disable(self):
		self.active = False

		del self.testSuite

def init(moduleManager):
	return TestModule(moduleManager)
