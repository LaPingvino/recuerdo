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
import bisect

class TestCase(unittest.TestCase):
	@property
	def _mods(self):
		#test the native python bisect module too, to double check this
		#tests are right.
		return [bisect] + list(self._mm.mods("active", type="bisectfunc"))

	def testLarger(self):
		for m in self._mods:
			self.assertEqual(m.bisect([1, 2, 3], 4), 3)
			self.assertEqual(m.bisect([1, 2, 3], 2000), 3)

	def testMiddle(self):
		for m in self._mods:
			self.assertEqual(m.bisect([1, 2, 3], 2), 2)

	def testSmaller(self):
		for m in self._mods:
			self.assertEqual(m.bisect([1, 2, 3], 0), 0)
			self.assertEqual(m.bisect([1, 2, 3], -100), 0)

	def testFloat(self):
		for m in self._mods:
			self.assertEqual(m.bisect([1, 2, 3], 1.4), 1)

class TestModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "test"
		self.requires = (
			self._mm.mods(type="bisectfunc"),
		)

	def enable(self):
		self.TestCase = TestCase
		self.TestCase._mm = self._mm
		self.active = True

	def disable(self):
		self.active = False
		del self.TestCase

def init(moduleManager):
	return TestModule(moduleManager)
