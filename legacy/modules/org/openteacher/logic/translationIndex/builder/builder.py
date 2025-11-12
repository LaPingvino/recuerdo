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

import os
import collections

class LazyDict(collections.Mapping):
	def __init__(self, iterable, *args, **kwargs):
		super(LazyDict, self).__init__(*args, **kwargs)

		self._iterable = iterable
		self._actualDict = None

	@property
	def _dict(self):
		if not self._actualDict:
			self._actualDict = dict(self._iterable)
		return self._actualDict

	def __iter__(self):
		return iter(self._dict)

	def __len__(self):
		return len(self._dict)

	def __getitem__(self, key):
		return self._dict[key]

	def __repr__(self):
		return "<%s %s>" % (self.__class__.__name__, repr(self._dict))

class TranslationIndexBuilderModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TranslationIndexBuilderModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "translationIndexBuilder"

	def enable(self):
		global polib
		try:
			import polib
		except ImportError:
			pass
		self.active = True

	def _translationsIn(self, poPath):
		po = polib.pofile(poPath)
		for entry in po.translated_entries():
			yield entry.msgid, entry.msgstr

	def _translationIndex(self, path):
		for poFile in os.listdir(path):
			if not poFile.endswith(".po"):
				continue
			poPath = os.path.join(path, poFile)

			langCode = poFile[:-len(".po")]
			translations = dict(self._translationsIn(poPath))
			yield langCode, translations

	def buildTranslationIndex(self, path):
		"""Builds an overview of all translations that are in the .po
		   files in ``path``. This overview is a dict with as key the
		   language code (e.g. pt_BR), and as value a dict that maps
		   English translations to their localized equivalents.

		"""
		#reading & parsing all .po files is relatively heavy (and
		#necessary only for dev tools), so building the index is delayed
		#until it's actually accessed.
		return LazyDict(self._translationIndex(path))

	def disable(self):
		self.active = False

def init(moduleManager):
	return TranslationIndexBuilderModule(moduleManager)
