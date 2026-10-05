// Test mode in Recuerdo's web page: when it is served by a test server
// (`recuerdo testserver`), students log in, take the tests their teacher
// gave them and see their results. The server keeps the answers; the
// questions come with their markup made safe (Go's internal/richtext).
"use strict";
let testUser = null;

// api calls the test server; it throws the server's error message.
async function testApi(method, path, body) {
	const opt = { method, credentials: "same-origin", headers: {} };
	if (body !== undefined) { opt.headers["Content-Type"] = "application/json"; opt.body = JSON.stringify(body); }
	const resp = await fetch("api/" + path, opt);
	const data = await resp.json().catch(() => ({}));
	if (!resp.ok) { const e = new Error(data.error || resp.statusText); e.status = resp.status; throw e; }
	return data;
}

function testShow(view) {
	$("start").hidden = true; $("lesson").hidden = true; $("testmode").hidden = false;
	for (const id of ["testLogin", "testList", "testTake", "testResult"]) $(id).hidden = id !== view;
	$("testLogout").hidden = !testUser;
	text($("testWho"), testUser ? testUser.name : "");
}
function testMessage(msg, kind) { const m = $("testMessage"); m.className = "feedback " + (kind || ""); text(m, msg || ""); }

async function openTestMode() {
	$("start").hidden = true; $("lesson").hidden = true; $("testmode").hidden = false;
	testMessage("");
	try { testUser = await testApi("GET", "me"); showTestList(); } catch (e) { testUser = null; testShow("testLogin"); $("testName").focus(); }
}

$("openTestMode").addEventListener("click", openTestMode);
$("testBack").addEventListener("click", () => {
	if (!$("testList").hidden || !$("testLogin").hidden) { $("testmode").hidden = true; $("start").hidden = false; return; }
	testMessage(""); showTestList();
});
$("testLogin").addEventListener("submit", async (e) => {
	e.preventDefault();
	try {
		const r = await testApi("POST", "login", { name: $("testName").value, password: $("testPassword").value });
		testUser = r.user; $("testPassword").value = ""; testMessage(""); showTestList();
	} catch (err) { testMessage(err.status === 401 ? t("Wrong name or password.") : err.message, "wrong"); }
});
$("testLogout").addEventListener("click", async () => {
	await testApi("POST", "logout").catch(() => {});
	testUser = null; testMessage(""); testShow("testLogin");
});

async function showTestList() {
	const box = $("testList"); box.replaceChildren(); testShow("testList");
	if (testUser.role !== "student") {
		const p = document.createElement("p"); text(p, t("This page is for students. Teachers and admins: use the teacher's page.")); box.append(p);
		return;
	}
	let tests = [];
	try { tests = await testApi("GET", "tests"); } catch (e) { testMessage(e.message, "wrong"); return; }
	if (!tests.length) { const p = document.createElement("p"); text(p, t("No tests for you yet.")); box.append(p); return; }
	const table = document.createElement("table"); table.className = "words tests";
	const head = document.createElement("tr");
	for (const h of [t("Test"), t("Words"), t("Status"), ""]) { const th = document.createElement("th"); text(th, h); head.append(th); }
	table.append(head);
	for (const test of tests) {
		const tr = document.createElement("tr");
		for (const v of [test.title, String(test.words)]) { const td = document.createElement("td"); text(td, v); tr.append(td); }
		const status = document.createElement("td"), action = document.createElement("td");
		const button = (label, fn, primary) => {
			const b = document.createElement("button"); b.className = "button" + (primary ? " primary" : ""); text(b, label);
			b.addEventListener("click", fn); action.append(b);
		};
		if (test.published) { text(status, t("Result ready")); button(t("See the result"), () => showTestResult(test.id), true); }
		else if (test.handedIn) text(status, t("Handed in"));
		else if (!test.open) text(status, t("Closed"));
		else { text(status, t("To do")); button(t("Take the test"), () => takeTest(test.id), true); }
		tr.append(status, action); table.append(tr);
	}
	box.append(table);
}

async function takeTest(id) {
	let test;
	try { test = await testApi("GET", "tests/" + id); } catch (e) { testMessage(e.message, "wrong"); return; }
	const form = $("testTake"); form.replaceChildren(); testShow("testTake"); testMessage("");
	const h = document.createElement("h3"); text(h, test.title); form.append(h);
	const inputs = new Map();
	test.items.forEach((it, n) => {
		const row = document.createElement("div"); row.className = "test-item";
		const label = document.createElement("label"); label.className = "test-question";
		const num = document.createElement("span"); num.className = "test-num"; text(num, (n + 1) + ".");
		const q = document.createElement("span"); html(q, it.questionHtml);
		label.append(num, q);
		const input = document.createElement("input"); input.dir = "auto"; input.autocomplete = "off"; input.spellcheck = false;
		input.placeholder = it.math ? t("Formula, e.g. x^2 + 1") : t("Your answer");
		label.htmlFor = input.id = "testAnswer" + it.id;
		const field = document.createElement("div"); field.className = "word-field"; field.append(input);
		row.append(label, field);
		if (it.comment) { const c = document.createElement("p"); c.className = "hint"; text(c, it.comment); row.append(c); }
		if (it.math) row.append(formulaBuilder(input, false));
		form.append(row); inputs.set(it.id, input);
	});
	const actions = document.createElement("div"); actions.className = "row";
	const hand = document.createElement("button"); hand.className = "button primary"; hand.type = "submit"; text(hand, t("Hand in"));
	actions.append(hand); form.append(actions);
	let sure = false;
	form.onsubmit = async (e) => {
		e.preventDefault();
		const empty = [...inputs.values()].filter((i) => !i.value.trim()).length;
		if (!sure) {
			// a second press hands in: answers cannot be changed afterwards
			sure = true; text(hand, t("Yes, hand in"));
			testMessage(empty ? t("%d questions have no answer. Hand in anyway? You cannot change your answers afterwards.", empty)
				: t("Hand in? You cannot change your answers afterwards."), "");
			return;
		}
		const answers = {};
		for (const [itemId, input] of inputs) answers[itemId] = input.value;
		try {
			await testApi("POST", "tests/" + id + "/answers", { answers });
			testMessage(t("Handed in. Your teacher will check your answers."), "right"); showTestList();
		} catch (err) { sure = false; text(hand, t("Hand in")); testMessage(err.message, "wrong"); }
	};
	const first = inputs.values().next().value; if (first) first.focus();
}

async function showTestResult(id) {
	let test, result;
	try {
		[test, result] = await Promise.all([testApi("GET", "tests/" + id), testApi("GET", "tests/" + id + "/results/" + testUser.id)]);
	} catch (e) { testMessage(e.message, "wrong"); return; }
	const box = $("testResult"); box.replaceChildren(); testShow("testResult"); testMessage("");
	const h = document.createElement("h3"); text(h, test.title); box.append(h);
	const right = result.items.filter((i) => i.right).length;
	const p = document.createElement("p"); p.className = "question";
	text(p, t("%d of %d right (%d%%)", right, result.items.length, result.note)); box.append(p);
	const questions = new Map(test.items.map((it) => [it.id, it]));
	const table = document.createElement("table"); table.className = "words";
	const head = document.createElement("tr");
	for (const hd of [t("Question"), t("Your answer"), ""]) { const th = document.createElement("th"); text(th, hd); head.append(th); }
	table.append(head);
	for (const it of result.items) {
		const tr = document.createElement("tr");
		const q = document.createElement("td"); html(q, (questions.get(it.itemId) || {}).questionHtml || "");
		const g = document.createElement("td"); g.dir = "auto"; text(g, it.given || "—");
		const m = document.createElement("td"); m.className = it.right ? "ok" : "no";
		text(m, it.right ? (it.overridden ? t("Right (your teacher's decision)") : t("Right")) : t("Wrong"));
		tr.append(q, g, m); table.append(tr);
	}
	box.append(table);
}

// offered when the page comes from a test server
fetch("api/info").then((r) => r.ok ? r.json() : null).then((info) => {
	if (info && info.name === "Recuerdo test server") $("testEntry").hidden = false;
}).catch(() => {});
