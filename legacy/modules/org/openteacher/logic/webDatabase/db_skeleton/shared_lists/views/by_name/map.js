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

(function (doc) {
	if (doc.type === "list") {
		for (var i = 0; i < doc.shares.length; i += 1) {
			//sort order: by share, by title, by lastEdited
			emit([doc.shares[i], doc.title.toLowerCase(), doc.lastEdited], doc);
		}
	}
});
