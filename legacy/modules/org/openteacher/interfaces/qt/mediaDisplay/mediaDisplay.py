#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011, Milan Boers
#	Copyright 2011-2012, 2017, Marten de Vries
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

import platform
import contextlib

MIN_VOLUME = 0
BASE_VOLUME = 50
MAX_VOLUME = 100

def installQtClasses():
	global MediaControlDisplay, MediaDisplay

	class MediaControlDisplay(QtWidgets.QWidget):
		"""The video player and web viewer combination widget with controls

		"""
		def __init__(self, noPhonon, autoplay=True, *args, **kwargs):
			super(MediaControlDisplay, self).__init__(*args, **kwargs)

			self.autoplay = autoplay
			self.activeModule = None

			self.noPhonon = noPhonon

			self.mediaDisplay = MediaDisplay(self.autoplay, self.noPhonon)
			# Do not add the Phonon widget if it is not necessary
			if not self.noPhonon:
				self.mediaDisplay.videoPlayer.stateChanged.connect(self._playPauseButtonUpdate)

			layout = QtWidgets.QVBoxLayout()

			# Do not add the controls if there is not going to be any Phonon
			if not self.noPhonon:
				buttonsLayout = QtWidgets.QHBoxLayout()

				self.pauseButton = QtWidgets.QPushButton()
				self.pauseButton.setIcon(QtGui.QIcon.fromTheme("media-playback-pause",QtGui.QIcon(base._mm.resourcePath("icons/player_pause.png"))))
				self.pauseButton.clicked.connect(self.playPause)
				buttonsLayout.addWidget(self.pauseButton)

				self.seekSlider = QtWidgets.QSlider()
				self.seekSlider.setOrientation(QtCore.Qt.Horizontal)
				self.seekSlider.setRange(0, 0)
				@self.seekSlider.sliderMoved.connect
				def onSliderChange(position):
					self.mediaDisplay.videoPlayer.setPosition(position)
				@self.mediaDisplay.videoPlayer.positionChanged.connect
				def onPosChange(position):
					self.seekSlider.setValue(position)
				@self.mediaDisplay.videoPlayer.durationChanged.connect
				def onDurationChange(duration):
					self.seekSlider.setRange(0, duration)

				self.muteButton = QtWidgets.QToolButton()
				@self.muteButton.clicked.connect
				def muteClicked():
					if self.mediaDisplay.videoPlayer.volume():
						self._setVolume(MIN_VOLUME)
					else:
						self._setVolume(BASE_VOLUME)
				self.volumeSlider = QtWidgets.QSlider()
				self.volumeSlider.setOrientation(QtCore.Qt.Horizontal)
				self.volumeSlider.setRange(MIN_VOLUME, MAX_VOLUME)
				self.volumeSlider.setMaximumWidth(MAX_VOLUME)
				self.volumeSlider.sliderMoved.connect(self._setVolume)

				self._setVolume(BASE_VOLUME)

				buttonsLayout.addWidget(self.seekSlider)
				buttonsLayout.addWidget(self.muteButton)
				buttonsLayout.addWidget(self.volumeSlider)

			# Add the stacked widget
			layout.addWidget(self.mediaDisplay)

			if not self.noPhonon:
				layout.addLayout(buttonsLayout)

			self.setLayout(layout)

			# Disable the controls
			self.setControls()

		def _setVolume(self, volume):
			self.mediaDisplay.videoPlayer.setVolume(volume)
			self.volumeSlider.setValue(volume)
			if volume:
				self.muteButton.setIcon(self.style().standardIcon(QtWidgets.QStyle.SP_MediaVolume))
			else:
				self.muteButton.setIcon(self.style().standardIcon(QtWidgets.QStyle.SP_MediaVolumeMuted))

		def showMedia(self, path, remote, autoplay):
			modules = base._modules.sort("active", type="mediaType")
			for module in modules:
				if module.supports(path):
					chosenModule = module
					break

			chosenModule.showMedia(chosenModule.path(path, self.autoplay), self.mediaDisplay, autoplay)
			self.activeModule = chosenModule

			self.setControls()

		def setControls(self):
			# Only if there are controls
			if not self.noPhonon:
				needsControls = bool(self.activeModule and self.activeModule.phononControls)
				self._setControlsEnabled(enabled=needsControls)

		def playPause(self, event):
			if not self.noPhonon:
				if self.mediaDisplay.videoPlayer.state() == QtMultimedia.QMediaPlayer.PausedState:
					self.mediaDisplay.videoPlayer.play()
				else:
					self.mediaDisplay.videoPlayer.pause()

		def stop(self):
			if not self.noPhonon:
				self.mediaDisplay.videoPlayer.stop()

		def clear(self):
			self.mediaDisplay.clear()
			# Set the active type
			self.activeModule = None

			self.setControls()

		def _playPauseButtonUpdate(self, state):
			if state == QtMultimedia.QMediaPlayer.PausedState:
				self.pauseButton.setIcon(QtGui.QIcon.fromTheme("media-playback-play",QtGui.QIcon(base._mm.resourcePath("icons/player_play.png"))))
			else:
				self.pauseButton.setIcon(QtGui.QIcon.fromTheme("media-playback-pause",QtGui.QIcon(base._mm.resourcePath("icons/player_pause.png"))))

		def _setControlsEnabled(self, enabled):
			self.pauseButton.setEnabled(enabled)
			self.seekSlider.setEnabled(enabled)
			self.muteButton.setEnabled(enabled)
			self.volumeSlider.setEnabled(enabled)

	class MediaDisplay(QtWidgets.QStackedWidget):
		"""The video player and web viewer combination widget"""

		def __init__(self,autoplay,noPhonon,*args, **kwargs):
			super(MediaDisplay, self).__init__(*args, **kwargs)

			#self.activeType = None
			self.autoplay = autoplay

			self.noPhonon = noPhonon

			self.webviewer = QtWebKitWidgets.QWebView()
			self.webviewer.settings().setAttribute(QtWebKit.QWebSettings.PluginsEnabled, True)

			self.addWidget(self.webviewer)

			if not self.noPhonon:
				self.videoPlayer = QtMultimedia.QMediaPlayer(self)
				self.videoWidget = QtMultimediaWidgets.QVideoWidget()
				self.videoPlayer.setVideoOutput(self.videoWidget)
				self.addWidget(self.videoWidget)

		def clear(self):
			self.webviewer.setHtml('''
			<html><head><title>Nothing</title></head><body></body></html>
			''')
			if not self.noPhonon:
				self.videoPlayer.stop()
			self.setCurrentWidget(self.webviewer)

class MediaDisplayModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(MediaDisplayModule, self).__init__(*args, **kwargs)

		global base
		base = self

		self._mm = moduleManager

		self.type = "mediaDisplay"
		self.priorities = {
			"default": 495,
		}

		self.requires = (
			self._mm.mods(type="ui"),
		)
		self.uses = (
			self._mm.mods(type="translator"),
			self._mm.mods(type="settings"),
		)

		self.filesWithTranslations = ("mediaDisplay.py",)

	def enable(self):
		global QtCore, QtWidgets, QtGui, QtWebKit, QtWebKitWidgets, QtMultimedia, QtMultimediaWidgets
		try:
			from PyQt5 import QtCore, QtWidgets, QtGui, QtWebKit, QtWebKitWidgets
		except ImportError:
			return
		self._noPhonon = False
		try:
			from PyQt5 import QtMultimedia, QtMultimediaWidgets
		except ImportError:
			self._noPhonon = True
		installQtClasses()

		self._modules = set(self._mm.mods(type="modules")).pop()
		self.active = True

		# Add settings
		try:
			self._settings = self._modules.default(type="settings")
		except IndexError:
			self._html5VideoSetting = {"value": False}
			self._html5AudioSetting = {"value": False}
		else:
			# Settings (used in mediaTypes)
			self._html5VideoSetting = self._settings.registerSetting(**{
				"internal_name": "org.openteacher.lessons.media.videohtml5",
				"type": "boolean",
				"defaultValue": False,
				"advanced": True,
			})

			self._html5AudioSetting = self._settings.registerSetting(**{
				"internal_name": "org.openteacher.lessons.media.audiohtml5",
				"type": "boolean",
				"defaultValue": False,
				"advanced": True,
			})

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

	def _retranslate(self):
		#setup translation
		global _
		global ngettext

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			_, ngettext = unicode, lambda a, b, n: a if n == 1 else b
		else:
			_, ngettext = translator.gettextFunctions(
				self._mm.resourcePath("translations")
			)

		self._html5VideoSetting.update({
			"name": _("Use HTML5 for video"),
			"category": _("Lesson"),
			"subcategory": _("Media"),
		})
		self._html5AudioSetting.update({
			"name": _("Use HTML5 for audio"),
			"category": _("Lesson"),
			"subcategory": _("Media"),
		})

	def disable(self):
		self.active = False

		del self._modules
		del self._settings
		del self._html5VideoSetting
		del self._html5AudioSetting
		del self._noPhonon

	def createDisplay(self, autoplay):
		return MediaControlDisplay(self._noPhonon, autoplay)

def init(moduleManager):
	return MediaDisplayModule(moduleManager)
