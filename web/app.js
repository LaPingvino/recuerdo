// Recuerdo's web page: the lesson logic is Go (recuerdo.wasm, offered as
// globalThis.recuerdo by cmd/recuerdo-web); this file only shows it.
"use strict";
const $ = (id) => document.getElementById(id);
const EXAMPLE = "hond = dog\nkat = cat\nhuis = house, home\nboom = tree\nfiets = bicycle\nkaas = cheese\n";

let api = null;
function call(name, ...args) {
	const out = JSON.parse(api[name](...args));
	if (out && out.error) throw new Error(out.error);
	return out;
}
function text(el, s) { el.textContent = s; }
function showError(e) { const el = $("startError"); text(el, e.message || String(e)); el.hidden = false; }

// ---- start: open a file, type a list or try the example ----
async function openFile(file) {
	try {
		const bytes = new Uint8Array(await file.arrayBuffer());
		showLesson(call("open", file.name, bytes));
	} catch (e) { showError(e); }
}
$("file").addEventListener("change", (e) => e.target.files[0] && openFile(e.target.files[0]));
const drop = $("drop");
drop.addEventListener("dragover", (e) => { e.preventDefault(); drop.classList.add("over"); });
drop.addEventListener("dragleave", () => drop.classList.remove("over"));
drop.addEventListener("drop", (e) => {
	e.preventDefault(); drop.classList.remove("over");
	if (e.dataTransfer.files[0]) openFile(e.dataTransfer.files[0]);
});
$("useTyped").addEventListener("click", () => {
	try { showLesson(call("openText", "Typed list", $("typed").value)); } catch (e) { showError(e); }
});
$("example").addEventListener("click", () => { $("typed").value = EXAMPLE; showLesson(call("openText", "Example", EXAMPLE)); });

// ---- the lesson: words, practise, results ----
let lesson = null;
function showLesson(l) {
	lesson = l;
	$("start").hidden = true; $("lesson").hidden = false; $("startError").hidden = true;
	text($("lessonTitle"), l.title);
	const rows = $("wordRows"); rows.replaceChildren();
	for (const it of l.items) {
		const tr = document.createElement("tr");
		for (const s of [it.question, it.answer]) { const td = document.createElement("td"); text(td, s); tr.append(td); }
		rows.append(tr);
	}
	$("practice").hidden = true; $("done").hidden = true; $("options").hidden = false;
	showResults();
	showTab("practise");
}
function showTab(name) {
	for (const b of document.querySelectorAll(".tab")) b.classList.toggle("active", b.dataset.tab === name);
	for (const p of document.querySelectorAll(".panel")) p.hidden = p.dataset.panel !== name;
	if (name === "practise" && !$("practice").hidden) $("answer").focus();
	if (name === "results" && api) showResults();
}
for (const b of document.querySelectorAll(".tab")) b.addEventListener("click", () => showTab(b.dataset.tab));
$("close").addEventListener("click", () => { $("lesson").hidden = true; $("start").hidden = false; text($("lessonTitle"), ""); });

function fillChoices() {
	const c = call("choices");
	for (const [id, list] of [["lessonType", c.lessonTypes], ["order", c.orders], ["words", c.words]]) {
		const sel = $(id); sel.replaceChildren();
		for (const [value, label] of list) { const o = document.createElement("option"); o.value = value; o.textContent = label; sel.append(o); }
	}
}

function startPractice() {
	try {
		const st = call("start", JSON.stringify({
			lessonType: $("lessonType").value, order: $("order").value, words: $("words").value,
			askAnswers: $("askAnswers").checked,
		}));
		$("options").hidden = true; $("done").hidden = true; $("practice").hidden = false;
		text($("feedback"), ""); $("feedback").className = "feedback";
		showState(st);
	} catch (e) { text($("feedback"), e.message); }
}
$("startButton").addEventListener("click", startPractice);
$("again").addEventListener("click", () => { $("done").hidden = true; $("options").hidden = false; });

function showState(st) {
	if (st.done || !st.active) { finish(st); return; }
	text($("question"), st.question);
	text($("counter"), `${st.asked + 1} / ${st.total}`);
	$("progressBar").style.width = `${st.total ? (100 * st.asked) / st.total : 0}%`;
	$("answer").value = ""; $("answer").focus();
}
$("answerForm").addEventListener("submit", (e) => {
	e.preventDefault();
	const given = $("answer").value;
	if (!given.trim()) return;
	try {
		const r = call("answer", given);
		const fb = $("feedback");
		fb.className = "feedback " + (r.right ? "right" : "wrong");
		text(fb, r.right ? `Right: ${r.correct}` : `Wrong: ${given} → ${r.correct}`);
		showState(call("state"));
		showResults();
	} catch (err) { text($("feedback"), err.message); }
});
$("stopButton").addEventListener("click", () => finish(call("stop")));

function finish(st) {
	$("practice").hidden = true; $("done").hidden = false;
	const pct = st.answered ? Math.round((100 * st.right) / st.answered) : 0;
	text($("doneText"), `${st.right} of ${st.answered} right (${pct}%)`);
	lesson = call("lesson");
	showResults();
}

function showResults() {
	const rows = call("report") || [];
	text($("resultsSummary"), rows.length
		? `Last session: ${rows.filter((r) => r.right).length} of ${rows.length} right. Sessions in this lesson: ${lesson.sessions}.`
		: "Practise to see your answers here.");
	const body = $("resultRows"); body.replaceChildren();
	for (const r of rows) {
		const tr = document.createElement("tr");
		for (const s of [r.question, r.answer, r.given]) { const td = document.createElement("td"); text(td, s); tr.append(td); }
		const mark = document.createElement("td"); mark.className = r.right ? "ok" : "no"; text(mark, r.right ? "right" : "wrong"); tr.append(mark);
		body.append(tr);
	}
}

// ---- download the lesson (with its results) ----
$("download").addEventListener("click", () => $("downloadDialog").showModal());
$("downloadDialog").addEventListener("close", () => {
	if ($("downloadDialog").returnValue !== "ok") return;
	const ext = $("format").value;
	const name = (lesson.title || "lesson").replace(/[\\/:*?"<>|]+/g, "_") + "." + ext;
	const bytes = api.save(name);
	if (typeof bytes === "string") { alert(JSON.parse(bytes).error); return; }
	const a = document.createElement("a");
	a.href = URL.createObjectURL(new Blob([bytes])); a.download = name; a.click();
	setTimeout(() => URL.revokeObjectURL(a.href), 1000);
});

// ---- start the WebAssembly ----
(async () => {
	const go = new Go();
	const wasm = await WebAssembly.instantiateStreaming(fetch("recuerdo.wasm"), go.importObject)
		.catch(async () => WebAssembly.instantiate(await (await fetch("recuerdo.wasm")).arrayBuffer(), go.importObject));
	go.run(wasm.instance);
	api = globalThis.recuerdo;
	fillChoices();
	$("loading").hidden = true; $("start").hidden = false;
	// ?example opens the example lesson at once (for screenshots and demos)
	if (new URLSearchParams(location.search).has("example")) {
		$("example").click();
		const tab = new URLSearchParams(location.search).get("tab");
		if (tab) showTab(tab);
		if (new URLSearchParams(location.search).has("practise")) {
			startPractice();
			for (const a of ["dog", "kat"]) { $("answer").value = a; $("answerForm").requestSubmit(); }
		}
	}
})();
