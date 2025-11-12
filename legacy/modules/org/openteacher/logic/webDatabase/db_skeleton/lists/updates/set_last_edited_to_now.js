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

(function (doc, req) {
	doc = JSON.parse(req.body);
	if (doc.type !== "list") {
		return [null, "Please use this update function on list documents only."];
	}
	doc._id = req.uuid;
	doc.lastEdited = new Date();
	return [doc, toJSON(doc)];
});
