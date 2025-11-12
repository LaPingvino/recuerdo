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

import json
import os
import posixpath

class TranslationIndexJSONWriterModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TranslationIndexJSONWriterModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "translationIndexJSONWriter"

		self.requires = (
			self._mm.mods(type="friendlyTranslationNames"),
		)

	@property
	def _languages(self, cache={}):
		if "languages" not in cache:
			cache["languages"] = self._modules.default("active", type="friendlyTranslationNames").friendlyNames
		return cache["languages"]

	def writeJSONIndex(self, index, path, url):
		"""Writes a translation ``index`` to JSON files under the
		   ``path`` directory. ``path`` shouldn't exist yet. After
		   running this method, the file ``path``/index.js will exist
		   with links to the other files in the directory. Those links
		   will start with ``url``, so make sure that in the resulting
		   web app the contents of ``path`` are served at ``url``

		"""
		os.mkdir(path)

		options = dict(separators=(",", ":"), encoding="UTF-8")
		indexFileData = {}
		for langCode, translations in index.iteritems():
			jsLangCode = langCode.replace("_", "-")
			filename = jsLangCode + ".json"
			with open(os.path.join(path, filename), "w") as f:
				json.dump(translations, f, **options)
			indexFileData[jsLangCode] = {
				"name": self._languages[langCode],
				"url": posixpath.join(url, filename)
			}

		indexFileData["en"] = {"name": self._languages["C"]}

		with open(os.path.join(path, "index.js"), "w") as f:
			data = json.dumps(indexFileData, **options)
			f.write("var translationIndex=%s;" % data)

	def enable(self):
		self._modules = next(iter(self._mm.mods(type="modules")))

		self.active = True

	def disable(self):
		self.active = False

		del self._modules

def init(moduleManager):
	return TranslationIndexJSONWriterModule(moduleManager)
