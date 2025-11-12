#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2013-2014, Marten de Vries
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

import logging
logger = logging.getLogger(__name__)

class WebCouchDatabaseModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WebCouchDatabaseModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "webDatabase"
		self.requires = (
			self._mm.mods("javaScriptImplementation", type="safeHtmlChecker"),
			self._mm.mods("javaScriptImplementation", type="htmlGenerator"),
		)

	def enable(self):
		try:
			self._couch = self._mm.import_("couch")
		except ImportError, e:
			logger.debug(e)
			return
		self._modules = next(iter(self._mm.mods(type="modules")))

		self.active = True

	def createWebDatabase(self, host, username, password):
		kwargs = {
			"dbSkeletonDir": self._mm.resourcePath("db_skeleton"),
			"isSafeHtmlCode": self._modules.default("active", "javaScriptImplementation", type="safeHtmlChecker").code,
			"generateWordsHtml": self._modules.default("active", "javaScriptImplementation", type="htmlGenerator", dataType="words").code
		}
		return self._couch.WebCouch(host, username, password, **kwargs)

	def disable(self):
		self.active = False

		del self._couch
		del self._modules

def init(moduleManager):
	return WebCouchDatabaseModule(moduleManager)
