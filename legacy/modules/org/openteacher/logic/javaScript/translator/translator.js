/*
	Copyright 2012-2013, Marten de Vries
	Copyright 2012, Milan Boers

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

function translator(index, lang, callback) {
	if (index.hasOwnProperty(lang)) {
		//if a match, don't do anything
	} else if (index.hasOwnProperty(lang.split("-")[0])) {
		//if nl_NL doesn't match but nl does, lang becomes 'nl'
		lang = lang.split("-")[0];
	} else {
		//if all else fails...
		lang = "en";
	}

	if (typeof index[lang].url === "undefined") {
		//english, use a simple pass through function.
		var tr = function (str) {
			return str;
		};
		callback(tr);
	} else {
		//download the translation file
		$.getJSON(index[lang].url, function (translations) {
			//use it for translating the ui.
			var tr = function (str) {
				return translations[str] || str;
			};
			callback(tr);
		});
	}
}
