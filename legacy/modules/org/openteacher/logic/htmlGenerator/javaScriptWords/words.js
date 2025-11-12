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

function generateWordsHtml(list, options) {
	"use strict";

	if (typeof options === "undefined") {
		options = {};
	}

	if (typeof options.margin !== "string") {
		options.margin = "0";
	}
	if (typeof options.coloredRows !== "boolean") {
		options.coloredRows = true;
	}

	return tmpl(WORDS_HTML_GENERATOR_TEMPLATE, {
		"list": list,
		"margin": options.margin,
		"coloredRows": options.coloredRows,
		"compose": compose
	});
}
