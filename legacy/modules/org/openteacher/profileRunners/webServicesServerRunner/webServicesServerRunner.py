#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2013, 2017, Marten de Vries
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
import __builtin__

class WebServicesServerRunnerModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(WebServicesServerRunnerModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "webServicesServerRunner"
		self.requires = (
			self._mm.mods(type="execute"),
			self._mm.mods(type="webServicesServer"),
		)
		self.priorities = {
			"default": -1,
			"web-services-server": 0,
		}

	def _run(self):
		app = self._modules.default("active", type="webServicesServer").app
		if os.environ.get("UWSGI"):
			__builtin__.application = app
		else:
			print("Server will soon be running at: http://0.0.0.0:5000/")
			app.run(host="0.0.0.0")

	def enable(self):
		self._modules = next(iter(self._mm.mods(type="modules")))
		self._modules.default(type="execute").startRunning.handle(self._run)

		self.active = True

	def disable(self):
		self.active = False

		del self._modules

def init(moduleManager):
	return WebServicesServerRunnerModule(moduleManager)
