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
import datetime

class TestCase(unittest.TestCase):
	def _assertSafe(self, html):
		for mod in self._mm.mods("active", type="safeHtmlChecker"):
			self.assertTrue(mod.isSafeHtml(html), u"'%s' is marked unsafe" % html)

	def _assertUnsafe(self, html):
		try:
			self._assertSafe(html)
		except AssertionError:
			pass
		else:
			self.fail(u"'%s' is marked safe" % html)

	def testScripting(self):
		self._assertUnsafe(u"<script>alert(0);</script>")

	def testB(self):
		self._assertSafe(u" <b>test</b>")

	def testArg(self):
		self._assertUnsafe("<script type='text/javascript'>")

	def testDoubleTag(self):
		self._assertSafe(u"<em script>")
		self._assertUnsafe(u" <script em>")

	def checkWhitespaceInsideTag(self):
		self._assertSafe("<   	em >")
		self._assertUnsafe("<  	script >")

	def testUnicode(self):
		self._assertSafe(u"úñí©óðé")

	def testEmpty(self):
		self._assertSafe(u"")

	def testPieceOfText(self):
		self._assertSafe(u"""
			<h1>A small story...</h1>
			<p>I can't <em><strong>emphasise</strong></em> enough that
			<b>&lt;b&gt; tags</b> are outdated. Same for <i>the tags</i>
			causing the latest effect. <span>Span</span> tags are the
			new things.
			<h3>test</h3>
		""")

	def testFastClosing(self):
		self._assertSafe("<br />")
		self._assertUnsafe("<a />")
		self._assertSafe("<br/>")
		self._assertUnsafe("<a/>")

	def testWeirdCasing(self):
		self._assertSafe("<SpAN>test</sPAn")
		self._assertUnsafe("<StYlE />")

	def testClosingWithDifferentTag(self):
		#by choice, unsafe:
		self._assertUnsafe("<a></script>")
		self._assertSafe("<b></i>")

	def testImg(self):
		self._assertUnsafe("<p>some text <img src='#' /></p>")

	def testNonString(self):
		def test():
			pass
		self._assertUnsafe(23)
		self._assertUnsafe(datetime.datetime.now())
		self._assertUnsafe([])
		self._assertUnsafe({})
		self._assertUnsafe(None)
		self._assertUnsafe(False)
		self._assertUnsafe(test)

	def testAttribute(self):
		self._assertUnsafe("<b onclick='alert(1);'>test</b>")
		#by choice.
		self._assertUnsafe("<b style='color:#333'>abc</test>")

class TestModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "test"
		self.requires = (
			self._mm.mods(type="safeHtmlChecker"),
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
