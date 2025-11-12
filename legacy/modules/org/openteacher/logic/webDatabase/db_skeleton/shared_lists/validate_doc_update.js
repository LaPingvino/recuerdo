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
	if (userCtx.roles.indexOf("_admin") === -1) {
		throw({"unauthorized": "need to be admin to make changes."});
	}
	if (newDoc._deleted) {
		//that's fine.
		return;
	}
	if (oldDoc && oldDoc.type === "list" && oldDoc.shares.length) {
		//You can never get a shared list
		//de-published again, but at least it's
		//removed from the view this time, and from
		//next time on changes to the list will not
		//be visible anymore (until the list is
		//shared again.)
		return;
	}
	if (newDoc.type !== "list") {
		throw({"forbidden": "should be here only if a list."});
	}
	if (!newDoc.shares.length) {
		throw({"forbidden": "should be here only if having shares."});
	}
});
