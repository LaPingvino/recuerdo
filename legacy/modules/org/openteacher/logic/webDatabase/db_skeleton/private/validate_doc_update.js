/*
	Copyright 2013-2014, Marten de Vries

	This file is part of OpenTeacher.

	OpenTeacher is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	OpenTeacher is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with OpenTeacher.  If not, see <http://www.gnu.org/licenses/>.
*/

/*jshint expr:true*/

(function (newDoc, oldDoc, userCtx) {
	if (newDoc._deleted) {
		//that's fine.
		return;
	}
	var lib = require("validation_lib");
	lib.requireAttr(newDoc, "type");

	if (newDoc.type === "list") {
		lib.requireAttr(newDoc, "title");
		lib.assertSafeHtml(newDoc.title);

		lib.requireAttr(newDoc, "shares");
		for (var i = 0; i < newDoc.shares.length; i += 1) {
			lib.assertSafeHtml(newDoc.shares[i]);
		}

		lib.assertSafeHtml(newDoc.questionLanguage || "");
		lib.assertSafeHtml(newDoc.answerLanguage || "");

		lib.requireAttr(newDoc, "lastEdited");
		lib.assertValidDate(newDoc.lastEdited);

		var items = newDoc.items || [];
		for (var j = 0; j < items.length; j += 1) {
			var item = items[j];
			lib.assertNumeric(item.id);
			lib.assertSafeHtml(item.comment || "");
			lib.assertSafeHtml(item.commentAfterAnswering || "");
			//close enough...
			lib.assertSafeHtml((item.questions || []).toString());
			lib.assertSafeHtml((item.answers || []).toString());
		}
	} else if (newDoc.type === "test") {
		lib.requireAttr(newDoc, "listId");
		var results = newDoc.results || [];
		for (var k = 0; k < results.length; k += 1) {
			lib.assertSafeHtml(results[k].givenAnswer || "");
		}
	} else if (newDoc.type === "setting") {
		lib.requireAttr(newDoc, "value");
	}
});
