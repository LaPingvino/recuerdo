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

var sum;

(function () {
	"use strict";

	function reduce(func, data) {
		var result = data.pop();
		while (data.length) {
			result = func(result, data.pop());
		}
		return result;
	}

	sum = function (data) {
		return reduce(function (a, b) {
			return a + b;
		}, data);
	};
}());

