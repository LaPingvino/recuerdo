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

/*jshint expr:true*/

(function (head, req) {
	var lib = require("presentation_lib");

	provides("atom", function () {
		var share_owner = req.info.db_name.substring("shared_lists_".length);
		var host = "http://" + req.headers.Host;
		var url = host + req.raw_path;
		var row = getRow();
		var share_name = row.key[0];

		send("<?xml version='1.0' encoding='UTF-8' ?>");
		send("<feed xmlns='http://www.w3.org/2005/Atom'>");
			send("<id>");
				send(lib.xmlEscape(url));
			send("</id>");
			send("<title>");
				send(lib.xmlEscape("Share '" + share_name + "' by '" + share_owner + "'"));
			send("</title>");
		
		do {
			send("<entry>");
				send("<content type='html'>");
					send(lib.xmlEscape(lib.generateWordsHtml(row.value)));
				send("</content>");
				send("<id>");
					send(row.id);
				send("</id>");
				send("<title>");
					send(lib.xmlEscape(row.value.title));
				send("</title>");
				send("<link href='");
					send(host + "/shared_lists_" + share_owner + "/_design/lists/_show/print/" + row.id);
				send("' />");
				send("<updated>");
					send(row.value.lastEdited);
				send("</updated>");
			send("</entry>");
			row = getRow();
		} while (row);
		send("</feed>");
	});
});
