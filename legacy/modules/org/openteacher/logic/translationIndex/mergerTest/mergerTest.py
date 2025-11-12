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

class TestCase(unittest.TestCase):
	"""Tests TranslationIndexesMerger. No tests for invalid input
	   currently, but since all real life input is most likely generated
	   via the TranslationIndexBuilder anyway it's not really worth it.

	"""
	@property
	def _mods(self):
		return self._mm.mods("active", type="translationIndexesMerger")

	def testMergeTwoLangs(self):
		for m in self._mods:
			self.assertEqual(m.mergeIndexes(
				{"de": {"no": "nein"}},
				{"nl": {"no": "nee"}},
			), {
				"de": {"no": "nein"},
				"nl": {"no": "nee"},
			})

	def testSimpleMergeSingleLang(self):
		for m in self._mods:
			self.assertEqual(m.mergeIndexes(
				{"de": {"no": "nein"}},
				{"de": {"yes": "ja"}},
			), {
				"de": {
					"no": "nein",
					"yes": "ja"
				},
			})

	def testOverwritingEntry(self):
		for m in self._mods:
			#last overwrites first
			self.assertEqual(m.mergeIndexes(
				{"nl": {"yes": "ja"}},
				{"nl": {"yes": "zeker"}}
			), {
				"nl": {"yes": "zeker"},
			})

class TestModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "test"
		self.requires = (
			self._mm.mods(type="translationIndexesMerger"),
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
