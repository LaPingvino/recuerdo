/*
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
	function formatNote(note) {
		if (note === 10) {
			//makes sure '10,0' isn't returned, since that's not a valid
			//dutch note. (It would mean that 10.8 would be possible,
			//which isn't.)
			return "10";
		}
		return note.toFixed(1).replace(".", ",");
	}

	function calculateNumber(test) {
		var results = map(function (result) {
			return result.result === "right" ? 1 : 0;
		}, test.results);
		var total = results.length;
		var amountRight = sum(results);

		return amountRight / total * 9 + 1;
	}

	calculateNote = function (test) {
		return formatNote(calculateNumber(test));
	};

	calculateAverageNote = function (tests) {
		var notes = map(calculateNumber, tests);
		return formatNote(sum(notes) / tests.length);
	};
}());
