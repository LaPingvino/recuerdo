// The touch typing course in the web version (Go's internal/typing does
// the course; the profiles are kept in this browser).
"use strict";
const TYPING_KEY = "recuerdo.typing.profiles";
const FINGER_COLORS = { 1: "#f4a3a3", 2: "#f6c38f", 3: "#f3e08a", 4: "#c8e6a0", 7: "#a6dcd8", 8: "#a9c8f0", 9: "#c3b5ef", 10: "#e7b3e0", 5: "#d7d7d7" };
const KEY_TEXT = { Backspace: "⌫", Tab: "Tab", Enter: "Enter", "Caps Lock": "Caps", Shift: "Shift", Space: "" };
let typingProfiles = [], typingCurrent = null, typingChoicesCache = null, typingRun = null;

function typingLoad() {
	try { typingProfiles = JSON.parse(localStorage.getItem(TYPING_KEY) || "[]"); } catch (e) { typingProfiles = []; }
}
function typingSave() {
	try { localStorage.setItem(TYPING_KEY, JSON.stringify(typingProfiles)); } catch (e) { /* private window: kept until the page closes */ }
}
function typingMessage(msg, kind) { const m = $("typingMessage"); m.className = "feedback " + (kind || ""); text(m, msg || ""); }
function typingChoices() { return typingChoicesCache || (typingChoicesCache = call("typingChoices")); }
function typingShowSection() {
	$("start").hidden = true; $("lesson").hidden = true; $("testmode").hidden = true; $("typing").hidden = false;
}

// keyboard draws a layout: every key in its finger's colour, the next key
// black, a wrong key red
function keyboard(layout, current, wrong) {
	const box = el("div", { className: "keyboard" });
	layout.rows.forEach((row, r) => {
		for (const k of row) {
			const key = el("div", { className: "key" }, k.label in KEY_TEXT ? KEY_TEXT[k.label] : k.label);
			key.style.left = (k.x / 15 * 100) + "%"; key.style.top = (r / 5 * 100) + "%";
			key.style.width = (k.width / 15 * 100) + "%"; key.style.height = "20%";
			key.style.background = FINGER_COLORS[k.finger] || "#ddd";
			const is = (c) => c && (c === " " ? k.label === "Space" : k.label === c.toLowerCase());
			if (is(wrong)) key.classList.add("wrong"); else if (is(current)) key.classList.add("current");
			box.append(key);
		}
	});
	return box;
}

function openTyping() {
	typingShowSection(); typingMessage(""); typingLoad();
	const body = $("typingBody"); body.replaceChildren();
	const choices = typingChoices();
	if (typingProfiles.length) {
		body.append(el("h3", {}, t("Who is practising?")));
		const list = el("div", { className: "row" });
		for (const p of typingProfiles) {
			const st = call("typingShow", JSON.stringify(p));
			list.append(button(`${p.name} (${t("level %d of %d", st.level, st.levels)})`, () => { typingCurrent = p; typingInstructions(); }, true));
		}
		body.append(list);
	}
	body.append(el("h3", {}, t("New profile")));
	const name = el("input", { placeholder: t("Name"), autocomplete: "off" });
	const layout = select(choices.layouts.map((l) => [l.id, l.name]), t("Keyboard")); layout.value = "qwerty";
	const lang = select(choices.languages, t("Words in"));
	const ui = (currentLanguage || "").split("_")[0];
	lang.value = choices.languages.some(([c]) => c === ui) ? ui : "en";
	const preview = el("div", {});
	const draw = () => preview.replaceChildren(keyboard(choices.layouts.find((l) => l.id === layout.value) || choices.layouts[0]));
	layout.addEventListener("change", draw); draw();
	body.append(el("div", { className: "row" }, name, layout, lang, button(t("Start"), () => {
		try {
			const st = call("typingNew", name.value, layout.value, lang.value);
			if (typingProfiles.some((p) => p.name.toLowerCase() === st.profile.name.toLowerCase())) {
				typingMessage(t("That name is taken: choose it in the list above."), "wrong"); return;
			}
			typingProfiles.push(st.profile); typingSave(); typingCurrent = st.profile; typingInstructions();
		} catch (e) { typingMessage(t("A name is needed."), "wrong"); }
	}, true)), preview);
	name.focus();
}

function typingInstructions() {
	typingRun = null; typingShowSection();
	const st = call("typingShow", JSON.stringify(typingCurrent));
	const body = $("typingBody"); body.replaceChildren();
	body.append(el("div", { className: "row" }, el("h3", {}, typingCurrent.name), el("span", { className: "spacer" }), button(t("Other profile"), openTyping)));
	for (const para of st.instruction.split("\n\n")) body.append(el("p", {}, para));
	const facts = [[t("Level"), `${st.level} / ${st.levels}`]];
	if (st.done) facts.push([t("Speed"), t("%d words per minute (needed: %d)", st.speed, st.target)], [t("Mistakes"), String(st.mistakes)]);
	else facts.push([t("Speed needed"), t("%d words per minute", st.target)]);
	body.append(el("dl", { className: "typing-facts" }, ...facts.flatMap(([k, v]) => [el("dt", {}, k), el("dd", {}, v)])));
	const start = button(t("Start the exercise"), typingExercise, true);
	body.append(el("div", { className: "row" }, start));
	start.focus();
}

function typingExercise() {
	const layout = typingChoices().layouts.find((l) => l.id === typingCurrent.layout) || typingChoices().layouts[0];
	typingRun = { text: [...typingCurrent.exercise], pos: 0, mistakes: 0, started: 0, layout };
	typingMessage("");
	const body = $("typingBody"); body.replaceChildren();
	body.append(el("p", { className: "hint" }, t("Type the text below. The clock starts at your first key; a wrong key is shown in red and must be typed again.")),
		el("div", { className: "typing-text", id: "typingText" }), el("div", { id: "typingKeys" }),
		el("div", { className: "row" }, button(t("Stop"), typingInstructions)));
	typingRender("");
	document.activeElement && document.activeElement.blur();
}

function typingRender(wrong) {
	const r = typingRun;
	const box = $("typingText"); box.replaceChildren(
		el("span", { className: "typed" }, r.text.slice(0, r.pos).join("")),
		el("span", { className: "next" }, r.text[r.pos] === " " ? " " : r.text[r.pos]),
		el("span", {}, r.text.slice(r.pos + 1).join("")));
	$("typingKeys").replaceChildren(keyboard(r.layout, r.text[r.pos], wrong));
	if (wrong) typingMessage(t("That's a mistake (mistakes: %d).", r.mistakes), "wrong"); else typingMessage("");
}

// typingType takes a typed character
function typingType(c) {
	const r = typingRun;
	if (!r || r.pos >= r.text.length) return;
	if (!r.started) r.started = performance.now();
	if (c !== r.text[r.pos]) { r.mistakes++; typingRender(c); return; }
	r.pos++;
	if (r.pos < r.text.length) { typingRender(""); return; }
	const seconds = (performance.now() - r.started) / 1000;
	const st = call("typingFinish", JSON.stringify(typingCurrent), seconds, r.mistakes);
	const i = typingProfiles.findIndex((p) => p.name === typingCurrent.name);
	typingCurrent = st.profile; if (i >= 0) typingProfiles[i] = st.profile; typingSave();
	typingInstructions();
}

document.addEventListener("keydown", (e) => {
	if (!typingRun || $("typing").hidden || e.ctrlKey || e.metaKey || e.altKey || e.key.length !== 1) return;
	e.preventDefault(); // the space bar does not scroll
	typingType(e.key);
});
$("openTyping").addEventListener("click", openTyping);
$("typingBack").addEventListener("click", () => {
	if (typingRun) { typingInstructions(); return; }
	$("typing").hidden = true; $("start").hidden = false;
});
