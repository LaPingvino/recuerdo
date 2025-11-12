var calculateAveragePercents, calculatePercents;

(function () {
	function ZeroDivisionError() {
		var err = new Error("integer division or modulo by zero");
		err.name = "ZeroDivisionError";
		return err;
	}

	calculateAveragePercents = function (tests) {
		var percents = map(calculatePercents, tests);
		var percentSum = sum(percents);

		if (tests.length === 0) {
			throw new ZeroDivisionError();
		}
		return Math.round(percentSum / tests.length);
	};

	calculatePercents = function (test) {
		var results = map(function (result) {
			return result.result === "right" ? 1 : 0;
		}, test.results);
		var total = results.length;
		var amountRight = sum(results);

		if (total === 0) {
			throw ZeroDivisionError();
		}
		return Math.round(amountRight / total * 100);
	};
}());
