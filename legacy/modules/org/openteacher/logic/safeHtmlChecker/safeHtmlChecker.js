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

var isSafeHtml = (function () {
	var whitelist = [
		"address", "b", "big", "blockquote", "br", "center", "cite",
		"code", "dd", "dfn", "div", "dl", "dt", "em", "font", "h1",
		"h2", "h3", "h4", "h5", "h6", "hr", "i", "li", "nobr", "ol",
		"p", "pre", "s", "samp", "small", "span", "strong", "sub",
		"sup", "table", "tbody", "td", "tfoot", "th", "thead", "tr",
		"tt", "u", "ul", "var"
	];

	var tagRe = new RegExp("<\\s*([a-zA-Z]+)( [^>]*)?/?>");

	return function (html) {
		if (typeof html !== "string") {
			return false;
		}
		var remainingHtml = html;
		while (true) {
			var match = remainingHtml.match(tagRe);
			if (match) {
				var tag = match[1].toLowerCase();
				if (whitelist.indexOf(tag) === -1) {
					return false; //unsafe tag
				}
				attrs = (match[2] || "");
				if (attrs.indexOf("=") !== -1) {
					//having at least one attribute: unsafe. Some
					//attributes are harmless, but that would require
					//another whitelist (and in the case of style='',
					//the most useful one, CSS parsing). -> maybe later.
					return false;
				}
				//+1 so the current tag is not matched again.
				var pos = remainingHtml.search(tagRe) + 1;
				remainingHtml = remainingHtml.substr(pos);
			} else {
				return true; //safe html
			}
		}
	};
}());
