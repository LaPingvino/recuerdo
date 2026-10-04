// Recuerdo's web page: the lesson logic is Go (recuerdo.wasm, offered as
// globalThis.recuerdo by cmd/recuerdo-web); this file only shows it.
"use strict";
const $ = (id) => document.getElementById(id);
const EXAMPLE = "hond = dog\nkat = cat\nhuis = house, home\nboom = tree\nfiets = bicycle\nkaas = cheese\n";
// words with markup: furigana (ruby) and formulas; answers are typed plainly
const RICH_EXAMPLE = [
	"<ruby>水<rt>みず</rt></ruby> = water",
	"<ruby>山<rt>やま</rt></ruby> = mountain",
	"<ruby>日本語<rt>にほんご</rt></ruby> = Japanese",
	"water = H<sub>2</sub>O",
	"carbon dioxide = CO<sub>2</sub>",
	"the area of a circle = $\\pi r^2$",
	"the roots of ax² + bx + c = $x = \\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}$",
].join("\n");

let api = null;
function call(name, ...args) {
	const out = JSON.parse(api[name](...args));
	if (out && out.error) throw new Error(out.error);
	return out;
}
function text(el, s) { el.textContent = s; }
// html shows a word's safe HTML (from Go's internal/richtext) in el, with
// its formulas ($...$, $$...$$, \(...\), \[...\]) typeset by KaTeX.
function html(el, safe) {
	el.innerHTML = safe || "";
	if (globalThis.renderMathInElement) {
		renderMathInElement(el, {
			delimiters: [
				{ left: "$$", right: "$$", display: true }, { left: "\\[", right: "\\]", display: true },
				{ left: "\\(", right: "\\)", display: false }, { left: "$", right: "$", display: false },
			],
			throwOnError: false,
		});
	}
}

// ---- the formula builder: buttons that insert TeX (Go's richtext.Palette) ----
// isMath reports whether the cursor in s is inside a $...$ formula.
function insideMath(s, pos) {
	let n = 0;
	for (let i = 0; i < pos; i++) { if (s[i] === "\\") i++; else if (s[i] === "$") n++; }
	return n % 2 === 1;
}
// formulaBuilder makes a palette for input, with a live preview. wrap:
// the field holds words with formulas in $...$ (the Words tab), so a
// button outside a formula starts one; otherwise the whole field is TeX
// (a formula answer).
function formulaBuilder(input, wrap) {
	const box = document.createElement("div"); box.className = "formula-builder";
	const preview = document.createElement("div"); preview.className = "formula-preview";
	const update = () => {
		const v = input.value;
		if (!v.trim()) { preview.replaceChildren(); return; }
		if (wrap) { text(preview, v); html(preview, preview.innerHTML); return; }
		if (globalThis.katex) katex.render(v, preview, { throwOnError: false });
		else text(preview, v);
	};
	for (const g of call("palette")) {
		const row = document.createElement("div"); row.className = "fb-group";
		const name = document.createElement("span"); name.className = "fb-name"; text(name, t(g.name)); row.append(name);
		for (const it of g.items) {
			const b = document.createElement("button"); b.type = "button"; b.className = "fb-button";
			b.textContent = it.label; b.title = it.tip ? t(it.tip) : it.template.replace("§", "").trim();
			// keep the focus (and the selection) in the field
			b.addEventListener("mousedown", (e) => e.preventDefault());
			b.addEventListener("click", () => {
				const s = input.selectionStart ?? input.value.length, e = input.selectionEnd ?? s;
				let { text: ins, cursor } = call("expand", it.id, input.value.slice(s, e));
				if (wrap && !insideMath(input.value, s)) { ins = "$" + ins + "$"; cursor++; }
				input.setRangeText(ins, s, e, "end");
				input.setSelectionRange(s + cursor, s + cursor);
				input.focus(); update();
			});
			row.append(b);
		}
		box.append(row);
	}
	input.addEventListener("input", update);
	box.append(preview); update();
	return box;
}

// ---- translations (the desktop's, through Go's internal/i18n) ----
// t translates msgid and fills in %s and %d (%% is a percent sign).
function t(msgid, ...args) {
	let s = api && api.t ? api.t(msgid) : msgid;
	let i = 0;
	return s.replace(/%(\[\d+\])?[sd%]/g, (m, n) => {
		if (m === "%%") return "%";
		const k = n ? Number(n.slice(1, -1)) - 1 : i++;
		return String(args[k]);
	});
}
// applyTranslations translates the page's marked texts (data-i18n: the
// element's English text, kept in data-msgid; data-i18n-placeholder).
function applyTranslations() {
	for (const el of document.querySelectorAll("[data-i18n]")) {
		if (!el.dataset.msgid) el.dataset.msgid = el.textContent.trim();
		el.textContent = t(el.dataset.msgid);
	}
	for (const el of document.querySelectorAll("[data-i18n-placeholder]")) {
		el.placeholder = t(el.dataset.i18nPlaceholder);
	}
	document.documentElement.lang = (currentLanguage || "en").replace("_", "-");
}
let currentLanguage = "";
async function useLanguage(lang) {
	currentLanguage = "";
	if (lang && lang !== "en") {
		// the language's translations (OpenTeacher's and Recuerdo's own)
		// go into the memory file system, where Go reads them
		for (const name of [lang + ".po", "recuerdo-" + lang + ".po"]) {
			const r = await fetch("translations/" + name);
			if (r.ok) globalThis.fs.writeFile("/translations/" + name, new Uint8Array(await r.arrayBuffer()));
		}
		const out = JSON.parse(api.setLanguage("/translations", lang));
		if (!(out && out.error)) currentLanguage = lang;
	} else {
		api.setLanguage("/translations", "");
	}
	applyTranslations();
	fillChoices();
	if (lesson) { showWords(); showResults(); }
}
// the language: chosen before (kept in this browser), or the browser's
async function startLanguages() {
	let langs = [];
	try { langs = await (await fetch("languages.json")).json(); } catch (e) { /* no menu */ }
	const menu = $("language");
	const add = (code, label) => { const o = document.createElement("option"); o.value = code; o.textContent = label; menu.append(o); };
	add("en", "English");
	for (const code of langs.sort((a, b) => api.languageName(a).localeCompare(api.languageName(b)))) add(code, api.languageName(code));
	const pick = (wanted) => {
		for (const w of wanted) {
			const n = (w || "").replace("-", "_");
			if (n === "en" || n.startsWith("en_") && !langs.includes(n)) return "en";
			if (langs.includes(n)) return n;
			const base = n.split("_")[0];
			if (langs.includes(base)) return base;
			const regional = langs.find((l) => l.startsWith(base + "_"));
			if (regional) return regional;
		}
		return "en";
	};
	let stored = null;
	try { stored = localStorage.getItem("recuerdo.language"); } catch (e) { /* private window */ }
	const param = new URLSearchParams(location.search).get("lang");
	const lang = pick([param, stored, ...(navigator.languages || [navigator.language])].filter(Boolean));
	menu.value = lang;
	await useLanguage(lang);
	menu.addEventListener("change", async () => {
		try { localStorage.setItem("recuerdo.language", menu.value); } catch (e) { /* not kept */ }
		await useLanguage(menu.value);
	});
}
function escapeText(x) { const d = document.createElement("span"); d.textContent = x; return d.innerHTML; }
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
	try { showLesson(call("openText", t("Typed list"), $("typed").value)); } catch (e) { showError(e); }
});
$("example").addEventListener("click", () => { $("typed").value = EXAMPLE; showLesson(call("openText", t("Example"), EXAMPLE)); });
$("richExample").addEventListener("click", () => {
	$("typed").value = RICH_EXAMPLE; showLesson(call("openText", t("Furigana and formulas"), RICH_EXAMPLE));
});

// ---- the lesson: words, practise, results ----
let lesson = null;
function showLesson(l) {
	lesson = l;
	$("start").hidden = true; $("lesson").hidden = false; $("startError").hidden = true;
	text($("lessonTitle"), l.title);
	$("titleEdit").value = l.title;
	showWords();
	$("practice").hidden = true; $("done").hidden = true; $("options").hidden = false;
	showResults();
	showTab("practise");
}
// the Words tab: every word editable, an empty row to add one
function wordRow(it) {
	const tr = document.createElement("tr");
	if (!it) tr.className = "new";
	const q = document.createElement("input"), a = document.createElement("input");
	q.value = it ? it.question : ""; a.value = it ? it.answer : "";
	q.placeholder = it ? "" : t("New question"); a.placeholder = it ? "" : t("Answer");
	const save = () => {
		try {
			if (it) {
				call("updateItem", it.id, q.value, a.value);
			} else if (q.value.trim() && a.value.trim()) {
				call("addItem", q.value, a.value);
				lesson = call("lesson"); showWords();
				const rows = $("wordRows").querySelectorAll("tr.new input"); if (rows[0]) rows[0].focus();
			}
		} catch (e) { alert(e.message); }
	};
	const plainHtml = (x) => { const d = document.createElement("span"); d.textContent = x; return d.innerHTML; };
	for (const [input, raw, safe] of [[q, it && it.question, it && it.questionHtml], [a, it && it.answer, it && it.answerHtml]]) {
		input.addEventListener("change", save);
		input.addEventListener("keydown", (e) => { if (e.key === "Enter") { input.blur(); save(); } });
		const td = document.createElement("td"); td.className = "word-cell";
		const field = document.createElement("div"); field.className = "word-field"; field.append(input);
		const sum = document.createElement("button"); sum.type = "button"; sum.className = "fb-toggle";
		sum.textContent = "∑"; sum.title = t("Formula builder");
		sum.addEventListener("click", () => {
			const open = td.querySelector(".formula-builder");
			if (open) { open.remove(); sum.classList.remove("on"); td.classList.remove("building"); return; }
			// the builder's live preview replaces the word's own
			td.append(formulaBuilder(input, true)); sum.classList.add("on"); td.classList.add("building"); input.focus();
		});
		field.append(sum); td.append(field);
		// words with markup also show how they look
		if (safe && (safe !== plainHtml(raw) || /\$|\\[(\[]/.test(raw))) {
			const p = document.createElement("div"); p.className = "preview"; html(p, safe); td.append(p);
		}
		tr.append(td);
	}
	const td = document.createElement("td");
	if (it) {
		const del = document.createElement("button"); del.className = "remove"; del.title = t("Remove"); del.textContent = "×";
		del.addEventListener("click", () => { call("removeItem", it.id); lesson = call("lesson"); showWords(); });
		td.append(del);
	}
	tr.append(td);
	return tr;
}
function showWords() {
	const rows = $("wordRows"); rows.replaceChildren();
	for (const it of lesson.items) rows.append(wordRow(it));
	rows.append(wordRow(null));
}
$("titleEdit").addEventListener("change", () => {
	call("setTitle", $("titleEdit").value); lesson = call("lesson"); text($("lessonTitle"), lesson.title);
});

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
		$("correctAnyway").hidden = true;
		text($("feedback"), ""); $("feedback").className = "feedback";
		showState(st);
	} catch (e) { text($("feedback"), e.message); }
}
$("startButton").addEventListener("click", startPractice);
$("again").addEventListener("click", () => { $("done").hidden = true; $("options").hidden = false; });

let repeatTimer = null, currentState = {};
function showState(st) {
	currentState = st;
	if (st.done || !st.active) { finish(st); return; }
	const mode = $("mode").value;
	// the words' safe HTML (furigana, H<sub>2</sub>O, pictures), made by
	// Go's internal/richtext
	html($("question"), st.questionHtml);
	text($("counter"), `${st.asked + 1} / ${st.total}`);
	$("progressBar").style.width = `${st.total ? (100 * st.asked) / st.total : 0}%`;
	const hint = $("modeHint"); hint.className = "mode-hint"; text(hint, "");
	clearTimeout(repeatTimer);
	$("inmind").hidden = mode !== "inmind"; $("judgeRow").hidden = true; $("viewAnswer").hidden = false;
	$("answerForm").hidden = mode === "inmind";
	$("answer").value = ""; $("answer").disabled = false;
	// a formula answer: typed as TeX (no $), with the formula builder
	const builder = $("answerBuilder");
	builder.hidden = !st.answerIsMath || mode === "inmind";
	if (!builder.hidden && !builder.firstChild) builder.append(formulaBuilder($("answer"), false));
	if (!builder.hidden) builder.querySelector(".formula-preview").replaceChildren();
	$("answer").placeholder = st.answerIsMath ? t("Formula, e.g. x^2 + 1") : t("Your answer");
	if (mode === "shuffle") text(hint, st.shuffle);
	if (mode === "repeat") {
		// the answer is shown first, then typed from memory
		hint.classList.add("answer-shown"); html(hint, st.answerHtml); $("answer").disabled = true;
		repeatTimer = setTimeout(() => { text(hint, ""); $("answer").disabled = false; $("answer").focus(); }, 2500);
		return;
	}
	if (mode !== "inmind") $("answer").focus();
}
$("viewAnswer").addEventListener("click", () => {
	call("viewAnswer");
	const hint = $("modeHint"); hint.classList.add("answer-shown"); html(hint, currentState.answerHtml);
	$("viewAnswer").hidden = true; $("judgeRow").hidden = false;
});
for (const [id, right] of [["judgeRight", true], ["judgeWrong", false]]) {
	$(id).addEventListener("click", () => {
		call("judge", right);
		$("correctAnyway").hidden = true;
		text($("feedback"), ""); showState(call("state")); showResults();
	});
}
$("skipButton").addEventListener("click", () => {
	call("skip"); text($("feedback"), ""); $("correctAnyway").hidden = true; showState(call("state"));
});
$("correctAnyway").addEventListener("click", () => {
	call("correctLast"); $("correctAnyway").hidden = true;
	const fb = $("feedback"); fb.className = "feedback right"; text(fb, t("Counted as right"));
	showState(call("state")); showResults();
});
$("answerForm").addEventListener("submit", (e) => {
	e.preventDefault();
	const given = $("answer").value;
	if (!given.trim()) return;
	try {
		const r = call("answer", given);
		const fb = $("feedback");
		fb.className = "feedback " + (r.right ? "right" : "wrong");
		$("correctAnyway").hidden = r.right;
		// the right answer with its markup; what was typed as plain text
		const esc = (x) => { const d = document.createElement("span"); d.textContent = x; return d.innerHTML; };
		html(fb, r.right ? t("Right: %s", r.correctHtml) : t("Wrong: %s → %s", esc(given), r.correctHtml));
		showState(call("state"));
		showResults();
	} catch (err) { text($("feedback"), err.message); }
});
$("stopButton").addEventListener("click", () => finish(call("stop")));

function finish(st) {
	$("practice").hidden = true; $("done").hidden = false;
	const pct = st.answered ? Math.round((100 * st.right) / st.answered) : 0;
	text($("doneText"), t("%d of %d right (%d%%)", st.right, st.answered, pct));
	lesson = call("lesson");
	showResults();
}

function showResults() {
	const rows = call("report") || [];
	text($("resultsSummary"), rows.length
		? t("Last session: %d of %d right. Sessions in this lesson: %d.", rows.filter((r) => r.right).length, rows.length, lesson.sessions)
		: t("Practise to see your answers here."));
	const body = $("resultRows"); body.replaceChildren();
	for (const r of rows) {
		const tr = document.createElement("tr");
		for (const safe of [r.questionHtml, r.answerHtml]) { const td = document.createElement("td"); html(td, safe); tr.append(td); }
		{ const td = document.createElement("td"); text(td, r.given); tr.append(td); }
		const mark = document.createElement("td"); mark.className = r.right ? "ok" : "no"; text(mark, r.right ? t("Right") : t("Wrong")); tr.append(mark);
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
	await startLanguages();
	$("loading").hidden = true; $("start").hidden = false;
	// ?example opens the example lesson at once (for screenshots and demos)
	if (new URLSearchParams(location.search).has("example")) {
		$(new URLSearchParams(location.search).get("example") === "rich" ? "richExample" : "example").click();
		const tab = new URLSearchParams(location.search).get("tab");
		if (tab) showTab(tab);
		const mode = new URLSearchParams(location.search).get("mode");
		if (mode) $("mode").value = mode;
		if (new URLSearchParams(location.search).has("practise")) {
			startPractice();
			if (!mode || mode === "typing" || mode === "shuffle") {
				for (const a of ["dog", "kat"]) { $("answer").value = a; $("answerForm").requestSubmit(); }
			}
		}
	}
})();
