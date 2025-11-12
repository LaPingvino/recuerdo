/*
	Copyright 2013, Marten de Vries

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

exports.assertSafeHtml = function (text) {
	if (!isSafeHtml(text)) {
		throw({forbidden: "'" + text + "' isn't safe HTML."});
	}
};

exports.requireAttr = function (obj, attr) {
	if (typeof obj[attr] === "undefined") {
		throw({forbidden: "'" + attr + "' should not be undefined."});
	}
	if (!obj[attr]) {
		throw({forbidden: "'" + attr + "' should not evaluate to false."});
	}
};

exports.assertValidDate = function (data) {
	if (isNaN(new Date(data).valueOf())) {
		//if not a valid date
		throw({forbidden: "'" + data + "' should be a valid JSON representation of a JS Date object (as serialized in JS)."});
	}
};

exports.assertNumeric = function (number) {
	if (typeof number !== "number") {
		throw({forbidden: "'" + number + "' should be a number."});
	}
};
