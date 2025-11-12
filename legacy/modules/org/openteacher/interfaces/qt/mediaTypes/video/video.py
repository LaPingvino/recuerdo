#! /usr/bin/env python
# -*- coding: utf-8 -*-

#    Copyright 2008-2011, Milan Boers
#    Copyright 2012-2013, 2017, Marten de Vries
#
#    This file is part of OpenTeacher.
#
#    OpenTeacher is free software: you can redistribute it and/or modify
#    it under the terms of the GNU General Public License as published by
#    the Free Software Foundation, either version 3 of the License, or
#    (at your option) any later version.
#
#    OpenTeacher is distributed in the hope that it will be useful,
#    but WITHOUT ANY WARRANTY; without even the implied warranty of
#    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#    GNU General Public License for more details.
#
#    You should have received a copy of the GNU General Public License
#    along with OpenTeacher.  If not, see <http://www.gnu.org/licenses/>.

import mimetypes

class MediaTypeModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(MediaTypeModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.uses = (
			self._mm.mods(type="settings"),
		)

		self.type = "mediaType"
		self.extensions = [".avi", ".wmv", ".flv", ".mp4", ".mpg", ".mpeg", ".mov"]
		self.priorities = {
			"default": 400,
		}

	def enable(self):
		global QtCore, QtMultimedia
		try:
			from PyQt5 import QtCore, QtMultimedia
		except ImportError:
			# 'noPhonon'
			pass
		self._modules = set(self._mm.mods(type="modules")).pop()

		try:
			self._settings = self._modules.default(type="settings")
			self._html5 = self._settings.setting("org.openteacher.lessons.media.videohtml5")["value"]
		except:
			self._html5 = False

		self.active = True

	def disable(self):
		self.active = False
		del self._modules
		if hasattr(self, "_settings"):
			del self._settings
		del self._html5

	@property
	def phononControls(self):
		return not self._html5

	def supports(self, path):
		return mimetypes.guess_type(str(path))[0].split("/")[0] == "video"

	def path(self, path, autoplay):
		return path

	def _mediaContent(self, path):
		return QtMultimedia.QMediaContent(QtCore.QUrl(u'file://' + path))

	def showMedia(self, path, mediaDisplay, autoplay):
		try:
			self._html5 = self._settings.setting("org.openteacher.lessons.media.videohtml5")["value"]
		except:
			self._html5 = False

		if self._html5 or mediaDisplay.noPhonon:
			if not mediaDisplay.noPhonon:
				# Stop any media playing
				mediaDisplay.videoPlayer.stop()
			# Set the widget to the web view
			mediaDisplay.setCurrentWidget(mediaDisplay.webviewer)
			# Set the right html
			autoplayhtml = ""
			if autoplay:
				autoplayhtml = '''autoplay="autoplay"'''
			mediaDisplay.webviewer.setHtml('''
			<html><head>
			<title>Video</title>
			<style type="text/css">
			body
			{
			margin: 0px;
			}
			</style>
			</head><body onresize="size()"><video id="player" src="file://''' + path + '''" ''' + autoplayhtml + ''' controls="controls" />
			<script>
			function size()
			{
				document.getElementById('player').style.width = window.innerWidth;
				document.getElementById('player').style.height = window.innerHeight;
			}
			size()
			</script>
			</body></html>
			''')
		else:
			# Set the widget to video player
			mediaDisplay.setCurrentWidget(mediaDisplay.videoWidget)
			# Play the video
			mediaDisplay.videoPlayer.setMedia(self._mediaContent(path))
			mediaDisplay.videoPlayer.play()
			if not autoplay:
				# Immediately pause it
				mediaDisplay.videoPlayer.pause()

def init(moduleManager):
	return MediaTypeModule(moduleManager)
