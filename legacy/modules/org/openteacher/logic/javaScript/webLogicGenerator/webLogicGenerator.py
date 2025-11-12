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

class WebLogicGeneratorModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WebLogicGeneratorModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "webLogicGenerator"
		self._logicModTypes = [
			#wordsString
			"wordsStringParser",
			"wordsStringComposer",

			#wordListString
			"wordListStringParser",
			"wordListStringComposer",

			#lesson
			"jsInputTypingLogic",
			"javaScriptLessonType",

			#else
			"javaScriptEvent",
			"noteCalculator",
			"jsTranslator",
		]
		self.requires = [
			self._mm.mods("javaScriptImplementation", type=type)
			for type in self._logicModTypes
		]
		self.requires = tuple(self.requires + [
			self._mm.mods(type="translationIndexesMerger")
		])

	def _logicMods(self):
		for type in self._logicModTypes:
			yield self._modules.default("active", "javaScriptImplementation", type=type)

	def writeLogicCode(self, path):
		#generate logic javascript
		logic = u""
		for mod in self._logicMods():
			#add to logic code var with an additional tab before every
			#line
			logic += "\n\n\n\t" + "\n".join("\t" + s for s in mod.code.split("\n")).strip()
		logic = logic.strip()
		template = pyratemp.Template(filename=self._mm.resourcePath("logic.templ.js"))
		with open(path, "w") as f:
			f.write(template(code=logic).encode("UTF-8"))

	@property
	def translationIndex(self):
		translationIndexes = (getattr(m, "translationIndex", {}) for m in self._logicMods())
		mergeIndexes = self._modules.default("active", type="translationIndexesMerger").mergeIndexes

		return mergeIndexes(*translationIndexes)

	def enable(self):
		global pyratemp
		try:
			import pyratemp
		except ImportError:
			return
		self._modules = next(iter(self._mm.mods(type="modules")))

		self.active = True

	def disable(self):
		self.active = False

		del self._modules

def init(moduleManager):
	return WebLogicGeneratorModule(moduleManager)
