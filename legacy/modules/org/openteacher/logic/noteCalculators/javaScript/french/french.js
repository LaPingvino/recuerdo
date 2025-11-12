var calculateNote, calculateAverageNote;

(function () {
	function calculate(test) {
		var results = map(function (result) {
			return result.result === "right" ? 1 : 0;
		}, test.results);
		var total = results.length;
		var amountRight = sum(results);

		return Math.round(amountRight / total * 20);
	}

	calculateNote = function (test) {
		return calculate(test).toString();
	};

	calculateAverageNote = function (tests) {
		allNotes = map(calculate, tests);

		return Math.round(sum(allNotes) / tests.length).toString();
	};
}());
