#! /usr/bin/env python
# -*- coding: utf-8 -*-

#	Copyright 2011-2012, Milan Boers
#	Copyright 2012-2013, Marten de Vries
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
import uuid
import urllib2
import copy
import superjson
import contextlib

def installQtClasses():
	global AnswerChecker, PersonAdderWidget, PropertyLabel, StudentsInTestWidget, TakenTestWidget, TeacherPanel, TestActionWidget, TestInfoWidget, TestWidget, TestsWidget

	class PropertyLabel(QtWidgets.QLabel):
		def __init__(self, *args, **kwargs):
			super(PropertyLabel, self).__init__(*args, **kwargs)

			self.setAlignment(QtCore.Qt.AlignRight)

	class TestsWidget(QtWidgets.QWidget):
		"""Widget that shows all the tests"""

		testSelected = QtCore.pyqtSignal(dict)
		message = QtCore.pyqtSignal(str)
		def __init__(self, connection, upload, testSelecter, *args, **kwargs):
			super(TestsWidget, self).__init__(*args, **kwargs)

			# fixme > 3.0 maybe: create event when test is created so it can be added here

			self.connection = connection

			# Setup layout
			layout = QtWidgets.QVBoxLayout()

			self.testsLabel = QtWidgets.QLabel()
			layout.addWidget(self.testsLabel)

			testSelecter.testChosen.connect(self.testSelected.emit)
			layout.addWidget(testSelecter)

			self.addLessonButton = QtWidgets.QPushButton()
			self.addLessonButton.clicked.connect(upload)

			layout.addWidget(self.addLessonButton)

			self.setLayout(layout)

		def retranslate(self):
			self.testsLabel.setText(_("Tests"))
			self.addLessonButton.setText(_("Add lesson"))

	class PersonAdderWidget(QtWidgets.QWidget):
		"""The widget you see when you press 'Add student'"""

		back = QtCore.pyqtSignal()
		def __init__(self, connection, info, studentsView, studentsList, *args, **kwargs):
			"""Init parameters:
				- Connection object (like in the testModeConnection module)
				- The dictionary representing the test this adds persons for
				- Students view object (like in the testModeStudentsView module)
				- Students list object (like an object of the StudentsInTestWidget class here)

			"""
			super(PersonAdderWidget, self).__init__(*args, **kwargs)

			self.connection = connection
			self.info = info
			self.studentsView = studentsView
			self.studentsInTest = studentsList

			layout = QtWidgets.QVBoxLayout()

			self.label = QtWidgets.QLabel()
			layout.addWidget(self.label)

			layout.addWidget(self.studentsView)

			buttonLayout = QtWidgets.QHBoxLayout()

			self.addButton = QtWidgets.QPushButton()
			self.addButton.clicked.connect(self._addPersons)
			buttonLayout.addWidget(self.addButton)

			self.backButton = QtWidgets.QPushButton()
			self.backButton.clicked.connect(self.back.emit)
			buttonLayout.addWidget(self.backButton)

			layout.addLayout(buttonLayout)

			self.setLayout(layout)

		def retranslate(self):
			self.label.setText(_("Select a student/group:"))
			self.addButton.setText(_("Add person/group"))
			self.backButton.setText(_("Back"))

		# Adds the current person or group of the studentsView to the studentsList, but keeps unique
		def _addPersons(self):
			students = self.studentsView.getCurrentStudents()

			for student in students:
				# Add to the list (uniquely)
				if len(self.studentsInTest.findItems(student["username"], QtCore.Qt.MatchExactly)) == 0:
					# Add to the remote list
					self.connection.post(self.info["students"], {"student_id":student["id"]})

					# Add to the local list
					self.studentsInTest.update()
				else:
					# Student has already been added. Let's not say anything about it.
					pass

			self.back.emit()

	class StudentsInTestWidget(QtWidgets.QListWidget):
		"""Widget with the students in a test (second column, middle)"""

		def __init__(self, connection, testInfo, answerChecker, *args, **kwargs):
			super(StudentsInTestWidget, self).__init__(*args, **kwargs)

			self.connection = connection
			self.testInfo = testInfo
			self.answerChecker = answerChecker

			# If an answer in the answersChecker changes, I need to be updated
			self.answerChecker.answersChanged.connect(self.update)

			# Widget keeps a local buffer of info
			# list of tests/<id>/students/<id>
			self.studentInTests = []
			# list of users/<id>
			self.studentInfos = []

			self.update()

		# Initial adding of people to the list
		def update(self):
			# Clear self
			self.clear()

			# tests/<id>/students
			self.studentsInTest = self.connection.get(self.testInfo["students"])

			# tests/<id>/checked_answers
			checkedAnswers = self.connection.get(self.testInfo["checked_answers"])
			checkedAnswersIds = map(lambda x: int(os.path.basename(x)), checkedAnswers)
			# tests/<id>/answers
			answers = self.connection.get(self.testInfo["answers"])
			answersIds = map(lambda x: int(os.path.basename(x)), answers)

			for student in self.studentsInTest:
				studentInTest = self.connection.get(student)
				# Remember this, so the server can rest next time
				self.studentInTests.append(studentInTest)

				studentInfo = self.connection.get(studentInTest["student"])
				# Remember this, so the server can rest next time
				self.studentInfos.append(studentInfo)


				# Set appended text
				appender = _("(did not participate)")
				# Look if answers have already been checked
				if studentInfo["id"] in checkedAnswersIds:
					checkedAnswers = self.answerChecker.getCheckedAnswer(self.testInfo["id"], studentInfo["id"])
					appender = "(" + str(checkedAnswers["note"]) + ")"
				elif studentInfo["id"] in answersIds:
					appender = _("(Handed in)")

				self.addItem(studentInfo["username"] + " " + appender)

		def retranslate(self):
			self.update()

		def getCurrentStudentInTest(self):
			return self.studentInTests[self.currentRow()]

	class TestInfoWidget(QtWidgets.QWidget):
		# Parameter = dictionary as parsed tests/<id>/students/<id>
		takenTestSelected = QtCore.pyqtSignal(dict)
		def __init__(self, connection, info, list, answerChecker, *args, **kwargs):
			super(TestInfoWidget, self).__init__(*args, **kwargs)

			layout = QtWidgets.QVBoxLayout()
			fl = QtWidgets.QFormLayout()

			self.questionsLabel = QtWidgets.QLabel()

			fl.addRow(self.questionsLabel, PropertyLabel(str(len(list["items"]))))
			layout.addLayout(fl)

			self.studentsInTestLabel = QtWidgets.QLabel()
			self.studentsInTest = StudentsInTestWidget(connection, info, answerChecker)
			self.studentsInTest.currentItemChanged.connect(self.selectedStudentChanged)

			layout.addWidget(self.studentsInTestLabel)
			layout.addWidget(self.studentsInTest)

			self.addPersonButton = QtWidgets.QPushButton()
			layout.addWidget(self.addPersonButton)

			self.setLayout(layout)

		def selectedStudentChanged(self, current, previous):
			self.takenTestSelected.emit(self.studentsInTest.getCurrentStudentInTest())

		def retranslate(self):
			self.questionsLabel.setText(_("Questions:"))
			self.studentsInTestLabel.setText(_("People in this test:"))
			self.addPersonButton.setText(_("Add person"))

			self.studentsInTest.retranslate()

	class TestActionWidget(QtWidgets.QStackedWidget):
		# Parameter = dictionary as parsed tests/<id>/students/<id>
		takenTestSelected = QtCore.pyqtSignal(dict)
		def __init__(self, connection, studentsView, info, list, answerChecker, *args, **kwargs):
			super(TestActionWidget, self).__init__(*args, **kwargs)

			self.testInfoWidget = TestInfoWidget(connection, info, list, answerChecker)
			self.testInfoWidget.addPersonButton.clicked.connect(self._addPerson)

			self.personAdderWidget = PersonAdderWidget(connection, info, studentsView, self.testInfoWidget.studentsInTest)
			self.personAdderWidget.back.connect(self._personAdded)

			self.addWidget(self.testInfoWidget)
			self.addWidget(self.personAdderWidget)

			self.testInfoWidget.takenTestSelected.connect(self.takenTestSelected.emit)

		def _addPerson(self):
			self.setCurrentWidget(self.personAdderWidget)

		def _personAdded(self):
			self.setCurrentWidget(self.testInfoWidget)

		def retranslate(self):
			self.testInfoWidget.retranslate()
			self.personAdderWidget.retranslate()

	class TestWidget(QtWidgets.QWidget):
		"""Widget that shows the currently selected test (second column)"""

		# Parameter = dictionary as parsed tests/<id>/students/<id>
		takenTestSelected = QtCore.pyqtSignal(dict)
		message = QtCore.pyqtSignal(str)
		def __init__(self, connection, info, studentsView, answerChecker, *args, **kwargs):
			super(TestWidget, self).__init__(*args, **kwargs)

			self.connection = connection
			self.info = info
			self.list = self.info["list"]
			self.answerChecker = answerChecker

			self.testLabel = QtWidgets.QLabel()

			layout = QtWidgets.QVBoxLayout()
			layout.addWidget(self.testLabel)

			name = QtWidgets.QLabel()
			name.setStyleSheet("font-size: 18px;")
			layout.addWidget(name)

			self.testActionWidget = TestActionWidget(self.connection, studentsView, self.info, self.list, self.answerChecker)
			self.testActionWidget.takenTestSelected.connect(self.takenTestSelected.emit)

			layout.addWidget(self.testActionWidget)

			self.checkButton = QtWidgets.QPushButton()
			self.checkButton.clicked.connect(self.checkAnswers)
			layout.addWidget(self.checkButton)

			self.publishButton = QtWidgets.QPushButton()
			self.publishButton.clicked.connect(self.publishAnswers)

			# If no answers in the test are checked, the publish button should be disabled.
			# Get checked answers for this test
			checkedAnswers = self.connection.get(info["checked_answers"])
			if len(checkedAnswers) == 0:
				self.publishButton.setEnabled(False)

			layout.addWidget(self.publishButton)

			self.setLayout(layout)

			# Fill widget with contents
			name.setText(self.list["title"])

			self.retranslate()

		def retranslate(self):
			self.testLabel.setText(_("Test"))
			self.checkButton.setText(_("(Re)check answers"))
			self.publishButton.setText(_("(Re)publish answers"))

			self.testActionWidget.retranslate()

		def checkAnswers(self):
			givenAnswers = self.connection.get(self.info["answers"])
			rightAnswers = self.list
			# Check results
			self.answerChecker.checkAnswers(self.info["id"], givenAnswers, rightAnswers)
			# Enable publish button
			self.publishButton.setEnabled(True)
			# Show message
			self.message.emit(_("Student's answers have been checked!"))

		def publishAnswers(self):
			# Publish results
			self.answerChecker.publishAnswers(self.info["id"])
			# Show message
			self.message.emit(_("Student's results have been (re)published!"))

	class AnswerChecker(QtCore.QObject):
		answersChanged = QtCore.pyqtSignal()
		def __init__(self, connection, testChecker, *args, **kwargs):
			super(AnswerChecker, self).__init__(*args, **kwargs)
			self.connection = connection
			self.testChecker = testChecker

			self.results = dict()

		def checkAnswers(self, testId, givenAnswersUrls, rightAnswers):
			for givenAnswersUrl in givenAnswersUrls:
				results = []
				givenAnswers = self.connection.get(givenAnswersUrl)
				givenAnswers =  json.loads(givenAnswers["list"])

				studentId = int(os.path.basename(givenAnswersUrl))

				# Loop over all answers
				for rightAnswer in rightAnswers["items"]:
					results.append(self.testChecker((self.lookupItem(rightAnswer["id"], givenAnswers["items"]))["answer"], rightAnswer))

				# Add results to list
				rightAnswers["results"] = results

				self.update(testId, {"list": superjson.dumps(rightAnswers), "note": self.calculateNote(results), "answer_id": studentId}, False)

			self.answersChanged.emit()

		def update(self, testid, result, emit=True):
			if not testid in self.results:
				self.results[testid] = dict()
			self.results[testid][result["answer_id"]] = result

			if emit:
				self.answersChanged.emit()

		def getCheckedAnswer(self, testid, studentid):
			# Check if it's already buffered locally
			if testid in self.results and studentid in self.results[testid]:
				return self.results[testid][studentid]
			else:
				# Otherwise, get it
				checkedAnswer = self.connection.get("tests/" + str(testid) + "/checked_answers/" + str(studentid))
				if type(checkedAnswer) == urllib2.HTTPError:
					return None
				result = {"list": checkedAnswer["list"], "note": checkedAnswer["note"], "answer_id": studentid}
				self.update(testid, result, False)
				return result

		def correctAnswer(self, testid, studentid, questionid):
			list = json.loads(self.results[testid][studentid]["list"])
			for result in list["results"]:
				if result["itemId"] == questionid:
					result["result"] = "right"
					break

			self.update(testid, {"list": superjson.dumps(list), "note": self.calculateNote(list["results"]), "answer_id": studentid})

		def publishAnswers(self, testid):
			for studentResult in self.results[testid].values():
				e = self.connection.post("tests/" + str(testid) + "/checked_answers", studentResult)
				if type(e) == urllib2.HTTPError:
					studentid = studentResult["answer_id"]
					newStudentResult = copy.deepcopy(studentResult)
					del newStudentResult["answer_id"]
					self.connection.put("tests/" + str(testid) + "/checked_answers/" + str(studentid), newStudentResult)

		def calculateNote(self, results):
			results = map(lambda x: 1 if x["result"] == "right" else 0, results)
			total = len(results)
			amountRight = sum(results)

			return int(float(amountRight) / float(total) * 100)

		def lookupItem(self, id, items):
			for item in items:
				if item["id"] == id:
					return item

	class TakenTestWidget(QtWidgets.QWidget):
		"""Widget that shows the currently selected person in a test (third column)"""

		message = QtCore.pyqtSignal(str)
		def __init__(self, connection, studentInTest, compose, answerChecker, appName, *args, **kwargs):
			super(TakenTestWidget, self).__init__(*args, **kwargs)

			self.answerChecker = answerChecker
			self.appName = appName

			self.studentInTest = studentInTest
			self.student = connection.get(studentInTest["student"])
			self.test = connection.get(studentInTest["test"])

			self.personLabel = QtWidgets.QLabel()

			layout = QtWidgets.QVBoxLayout()
			layout.addWidget(self.personLabel)

			name = QtWidgets.QLabel(self.student["username"])
			name.setStyleSheet("font-size: 18px;")
			layout.addWidget(name)

			# Get the checked answers
			self.checkedAnswers = self.answerChecker.getCheckedAnswer(self.test["id"], self.student["id"])

			if self.checkedAnswers == None:
				self.l = QtWidgets.QLabel()
				self.l.setWordWrap(True)
				self.l.setSizePolicy(QtWidgets.QSizePolicy.Preferred, QtWidgets.QSizePolicy.Expanding)
				layout.addWidget(self.l)
			else:
				list = json.loads(self.checkedAnswers["list"])
				# Add checked answer to the answer checker
				answerChecker.update(self.test["id"], {"list": self.checkedAnswers["list"], "note": self.checkedAnswers["note"], "answer_id": self.student["id"]})

				fl = QtWidgets.QFormLayout()
				self.answersRightLabelLabel = QtWidgets.QLabel()
				self.answersRightLabel = PropertyLabel()
				fl.addRow(self.answersRightLabelLabel, self.answersRightLabel)
				self.answersWrongLabelLabel = QtWidgets.QLabel()
				self.answersWrongLabel = PropertyLabel()
				fl.addRow(self.answersWrongLabelLabel, self.answersWrongLabel)
				self.markLabelLabel = QtWidgets.QLabel()
				self.markLabel = PropertyLabel()
				fl.addRow(self.markLabelLabel, self.markLabel)
				layout.addLayout(fl)

				self.table = QtWidgets.QTableWidget(3,3)
				# Fill table
				self.questionIds = []
				for item in list["items"]:
					# Find result
					for result in list["results"]:
						if result["itemId"] == item["id"]:
							itemResult = result
							break

					resultWidget = QtWidgets.QTableWidgetItem()
					if itemResult["result"] == "right":
						resultWidget.setCheckState(QtCore.Qt.Checked)
					else:
						resultWidget.setCheckState(QtCore.Qt.Unchecked)
					resultWidget.setFlags(QtCore.Qt.ItemIsSelectable | QtCore.Qt.ItemIsEnabled)

					self.table.setItem(len(self.questionIds), 0, resultWidget)
					qItem = QtWidgets.QTableWidgetItem(compose(item["questions"]))
					qItem.setFlags(QtCore.Qt.ItemIsSelectable | QtCore.Qt.ItemIsEnabled)
					self.table.setItem(len(self.questionIds), 1, qItem)
					aItem = QtWidgets.QTableWidgetItem(itemResult["givenAnswer"])
					aItem.setFlags(QtCore.Qt.ItemIsSelectable | QtCore.Qt.ItemIsEnabled)
					self.table.setItem(len(self.questionIds), 2, aItem)

					self.questionIds.append(item["id"])

				self.table.verticalHeader().setVisible(False)
				self.table.horizontalHeader().setResizeMode(QtWidgets.QHeaderView.Stretch)
				self.table.horizontalHeader().setResizeMode(0, QtWidgets.QHeaderView.ResizeToContents)
				self.table.resizeRowsToContents()

				self.table.setSelectionBehavior(QtWidgets.QAbstractItemView.SelectRows)
				self.table.setSelectionMode(QtWidgets.QAbstractItemView.SingleSelection)

				#other headers are set in retranslate()
				self.table.setHorizontalHeaderItem(0, QtWidgets.QTableWidgetItem(""))
				self.table.cellClicked.connect(self.questionSelected)

				layout.addWidget(self.table)

				self.correctButton = QtWidgets.QPushButton()
				self.correctButton.setEnabled(False)
				self.correctButton.clicked.connect(self.correctAnswer)

				layout.addWidget(self.correctButton)

				self.finalMarkLabelLabel = QtWidgets.QLabel()
				self.finalMarkLabel = PropertyLabel()
				self.finalMarkLabel.setStyleSheet("font-size: 18px;")

				fm = QtWidgets.QFormLayout()
				fm.addRow(self.finalMarkLabelLabel, self.finalMarkLabel)
				layout.addLayout(fm)

				self.fillLabels()

			self.setLayout(layout)

			self.retranslate()

		def retranslate(self):
			self.personLabel.setText(_("Person"))
			if hasattr(self, "l"):
				self.l.setText(_("Answers of this student have not been checked yet. Click the 'Check answers' button in the second column."))
			if hasattr(self, "answersRightLabelLabel"):
				self.answersRightLabelLabel.setText(_("Answers right:"))
			if hasattr(self, "answersWrongLabelLabel"):
				self.answersWrongLabelLabel.setText(_("Answers wrong:"))
			if hasattr(self, "markLabelLabel"):
				self.markLabelLabel.setText(_("{appName} mark:").format(appName=self.appName))
			if hasattr(self, "finalMarkLabelLabel"):
				self.finalMarkLabelLabel.setText(_("Final mark:"))
			if hasattr(self, "table"):
				self.table.setHorizontalHeaderItem(1, QtWidgets.QTableWidgetItem(_("Question")))
				self.table.setHorizontalHeaderItem(2, QtWidgets.QTableWidgetItem(_("Given answer")))
			if hasattr(self, "correctButton"):
				#TRANSLATORS: The name of a button clicked if an answer is correct.
				self.correctButton.setText(_("Correct"))

		def fillLabels(self):
			checkedAnswer = self.answerChecker.getCheckedAnswer(self.test["id"], self.student["id"])
			list = json.loads(checkedAnswer["list"])
			results = map(lambda x: 1 if x["result"] == "right" else 0, list["results"])

			answersRight = sum(results)
			answersWrong = len(list["results"]) - answersRight
			note = checkedAnswer["note"]

			self.answersRightLabel.setText(str(answersRight))
			self.answersWrongLabel.setText(str(answersWrong))
			self.markLabel.setText(str(note))
			self.finalMarkLabel.setText(str(note))

		def questionSelected(self, row, column):
			if self.table.item(row, 0).text() == "X":
				self.correctButton.setEnabled(True)
			else:
				self.correctButton.setEnabled(False)

		def correctAnswer(self):
			resultWidget = QtWidgets.QTableWidgetItem("O")
			resultWidget.setFlags(QtCore.Qt.ItemIsSelectable | QtCore.Qt.ItemIsEnabled)

			self.answerChecker.correctAnswer(self.test["id"], self.student["id"], self.questionIds[self.table.currentRow()])

			self.table.setItem(self.table.currentRow(), 0, resultWidget)

			# Update labels
			self.fillLabels()

	class TeacherPanel(QtWidgets.QSplitter):
		message = QtCore.pyqtSignal(str)
		def __init__(self, connection, studentsView, testSelecter, uploaderModule, answerChecker, compose, appName, *args, **kwargs):
			super(TeacherPanel, self).__init__(*args, **kwargs)

			self.connection = connection
			self.testSelecter = testSelecter
			self.compose = compose
			self.answerChecker = answerChecker
			self.appName = appName

			# Add tests layoutumn
			self.testsWidget = TestsWidget(connection, uploaderModule, testSelecter)
			self.testsWidget.testSelected.connect(lambda testInfo: self.addTestlayoutumn(testInfo, studentsView, self.answerChecker))
			self.testsWidget.message.connect(self.message.emit)

			self.addWidget(self.testsWidget)
			self.addWidget(QtWidgets.QWidget())

		def addTestlayoutumn(self, testInfo, studentsView, answerChecker):
			testWidget = TestWidget(self.connection, testInfo, studentsView, answerChecker)
			testWidget.takenTestSelected.connect(self.addTakenTestlayoutumn)
			testWidget.message.connect(self.message.emit)

			with contextlib.suppress(AttributeError):
				self.widget(1).setParent(None)
			self.insertWidget(1, testWidget)

		def addTakenTestlayoutumn(self, studentInTest):
			takenTestWidget = TakenTestWidget(self.connection, studentInTest, self.compose, self.answerChecker, self.appName)
			takenTestWidget.message.connect(self.message.emit)

			with contextlib.suppress(AttributeError):
				self.widget(2).setParent(None)
			self.insertWidget(2, takenTestWidget)

		def retranslate(self):
			for i in range(self.count()):
				widget = self.widget(i)
				if widget.__class__ == QtWidgets.QWidget:
					#splitter handle
					continue
				widget.retranslate()

class TestModeTeacherPanelModule(object):
	def __init__(self, moduleManager, *args, **kwargs):
		super(TestModeTeacherPanelModule, self).__init__(*args, **kwargs)
		self._mm = moduleManager

		self.type = "testModeTeacherPanel"
		x = 492
		self.priorities = {
			"all": x,
			"teacher": x,
			"code-documentation": x,
			"test-suite": x,
			"default": -1,
		}

		self.uses = (
			self._mm.mods(type="translator"),
		)
		self.requires = (
			self._mm.mods(type="metadata"),
			self._mm.mods(type="event"),
			self._mm.mods(type="ui"),
			self._mm.mods(type="testMenu"),
			self._mm.mods(type="testModeUploader"),
			self._mm.mods(type="testModeStudentsView"),
			self._mm.mods(type="testModeConnection"),
			self._mm.mods(type="wordsStringComposer"),
			self._mm.mods(type="dialogShower"),
		)
		self.filesWithTranslations = ("teacherPanel.py",)

	def enable(self):
		global QtCore, QtWidgets
		try:
			from PyQt5 import QtCore, QtWidgets
		except ImportError:
			return

		self._modules = set(self._mm.mods(type="modules")).pop()

		self._testMenu = self._modules.default("active", type="testMenu").menu

		self._action = self._testMenu.addAction(self.priorities["default"])
		self._action.triggered.handle(self.showPanel)

		self.dialogShower = self._modules.default("active", type="dialogShower")

		try:
			translator = self._modules.default("active", type="translator")
		except IndexError:
			pass
		else:
			translator.languageChanged.handle(self._retranslate)
		self._retranslate()

		self.active = True

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

		self._action.text = _("Teacher panel")
		if hasattr(self, "tab"):
			self.tab.title = _("Teacher Panel")
		if hasattr(self, "teacherPanel"):
			self.teacherPanel.retranslate()

	def disable(self):
		self.active = False

		self._action.remove()

		del self._modules
		del self._testMenu
		del self._action
		del self.dialogShower

	def showPanel(self):
		# First, login
		self.connection = self._modules.default("active", type="testModeConnection").getConnection()
		self.connection.loggedIn.handle(self.showPanel_)
		self.loginid = uuid.uuid4()
		self.connection.login(self.loginid)

	def showPanel_(self, loginid):
		# Check if this is indeed from the request I sent out
		if loginid == self.loginid:
			uiModule = self._modules.default("active", type="ui")

			studentsView = self._modules.default("active", type="testModeStudentsView").getStudentsView()
			upload = self._modules.default("active", type="testModeUploader").upload
			testSelecter = self._modules.default("active", type="testModeTestSelecter").getTestSelecter()
			testChecker = self._modules.default("active", type="wordsStringChecker").check
			compose = self._modules.default("active", type="wordsStringComposer").compose
			appName = self._modules.default("active", type="metadata").metadata["name"]
			# Create an answer checker
			answerChecker = AnswerChecker(self.connection, testChecker)

			self.teacherPanel = TeacherPanel(self.connection, studentsView, testSelecter, upload, answerChecker, compose, appName)

			self.tab = uiModule.addCustomTab(self.teacherPanel)
			self.tab.closeRequested.handle(self.tab.close)
			#set tab title by retranslating
			self._retranslate()

			self.teacherPanel.message.connect(self.showMessage)

	def showMessage(self, text):
		self.dialogShower.showMessage.send(self.tab, text)

def init(moduleManager):
	return TestModeTeacherPanelModule(moduleManager)
