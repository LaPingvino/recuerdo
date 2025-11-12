/*
	Copyright 2011, Cas Widdershoven
	Copyright 2009-2013, Marten de Vries

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

var calculateNote, calculateAverageNote;

(function () {
	function convert(percents) {
		var i = bisect([30, 40, 50, 55, 60, 70], percents);
		return ["F", "FX", "E", "D", "C", "B", "A"][i];
	}

	calculateNote = function (test) {
		return convert(calculatePercents(test));
	};

	calculateAverageNote = function (tests) {
		return convert(calculateAveragePercents(tests));
	};
}());
