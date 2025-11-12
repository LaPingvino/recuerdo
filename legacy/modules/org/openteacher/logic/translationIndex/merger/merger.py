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

class TranslationIndexesMergerModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TranslationIndexesMergerModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "translationIndexesMerger"

	def enable(self):
		self.active = True

	def mergeIndexes(self, *indexes):
		"""Merges translation ``*indexes`` into one master index.
		
		   Useful when multiple files that have translations need to
		   share one translation function which uses an index. (The
		   default OT translator doesn't, the JavaScript translator
		   does.)

		"""
		self._masterIndex = {}
		for index in indexes:
			for langCode, translations in index.iteritems():
				if not langCode in self._masterIndex:
					self._masterIndex[langCode] = {}
				self._masterIndex[langCode].update(translations)
		return self._masterIndex

	def disable(self):
		self.active = False

def init(moduleManager):
	return TranslationIndexesMergerModule(moduleManager)
