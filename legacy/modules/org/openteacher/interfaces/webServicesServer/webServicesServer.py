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

import contextlib
import json
import logging

logger = logging.getLogger(__file__)

class WebServicesServerModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WebServicesServerModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "webServicesServer"
		self.requires = (
			self._mm.mods(type="webDatabase"),
			self._mm.mods(type="translator"),
			self._mm.mods(type="metadata"),
		)
		self.uses = (
			self._mm.mods(type="load"),
			self._mm.mods(type="save"),
		)
		self.filesWithTranslations = ("serverImpl.py", "register-template.html")

	def enable(self):
		try:
			self._server = self._mm.import_("serverImpl")
		except ImportError, e:
			logger.debug(e)
			return

		self.app = self._server.app
		try:
			with open(self._mm.resourcePath("web-config.json")) as f:
				self.app.config.update(json.load(f))
		except IOError:
			#cleanup & remain disabled
			del self._server
			del self.app
			return

		self.app.config["REGISTER_TEMPLATE_PATH"] = self._mm.resourcePath("register-template.html")
		self.app.config["NEWS_TEMPLATE_PATH"] = self._mm.resourcePath("news-template.html")

		modules = next(iter(self._mm.mods(type="modules")))
		self.app.config["TRANSLATIONS_DIR"] = self._mm.resourcePath("translations")

		def gettextFunctions(language):
			translator = modules.default("active", type="translator")
			translationDir = self._mm.resourcePath("translations")
			return translator.gettextFunctions(translationDir, language)
		self._server.gettextFunctions = gettextFunctions
		self._server.createWebDatabase = modules.default("active", type="webDatabase").createWebDatabase
		self._server.loaders = modules.sort("active", type="load")
		self._server.savers = modules.sort("active", type="save")
		self._server.metadata = modules.default("active", type="metadata").metadata

		self._server.initialize_endpoints()

		self.active = True

	def disable(self):
		self.active = False

		del self.app
		del self._server

def init(moduleManager):
	return WebServicesServerModule(moduleManager)
