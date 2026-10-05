// Test mode for teachers and admins (see testmode.js): teachers make
// tests of lessons, give them to students and groups, check and publish
// the results; admins manage the accounts and groups.
"use strict";
let teacherDetail = null; // the test whose page is open

// el makes an element: el("button", {className: "button"}, "text", child...)
function el(tag, props, ...children) {
	const e = document.createElement(tag);
	Object.assign(e, props || {});
	for (const c of children) e.append(c);
	return e;
}
function button(label, onclick, primary) {
	const b = el("button", { type: "button", className: "button" + (primary ? " primary" : "") }, label);
	b.addEventListener("click", onclick);
	return b;
}
// confirmButton asks for a second press before it does something final
function confirmButton(label, sure, onclick) {
	let armed = false;
	const b = button(label, () => {
		if (!armed) { armed = true; b.textContent = sure; return; }
		onclick();
	});
	return b;
}
function table(headers, rows) {
	const tb = el("table", { className: "words" });
	tb.append(el("tr", {}, ...headers.map((h) => el("th", {}, h))));
	for (const r of rows) tb.append(el("tr", {}, ...r.map((c) => c instanceof Node ? el("td", {}, c) : el("td", {}, String(c)))));
	return tb;
}
function select(options, placeholder) {
	const s = el("select", {});
	s.append(el("option", { value: "" }, placeholder));
	for (const [value, label] of options) s.append(el("option", { value: String(value) }, label));
	return s;
}
async function guarded(fn) { try { await fn(); } catch (e) { testMessage(e.message, "wrong"); } }
// shown shows words (a word list's raw words) with their markup, made
// safe by Go's internal/richtext, and formulas typeset
function shown(list) {
	const span = el("span", {});
	html(span, api && api.sanitize ? api.sanitize((list || []).join(", ")) : "");
	return span;
}

// ---- teachers ----

async function showTeacherHome() {
	teacherDetail = null;
	const box = $("testTeacher"); box.replaceChildren(); testShow("testTeacher");
	const tests = await testApi("GET", "tests").catch((e) => { testMessage(e.message, "wrong"); return []; });
	box.append(el("h3", {}, t("Your tests")));
	box.append(tests.length ? table([t("Test"), t("Words"), t("Status"), ""], tests.map((x) => [
		x.title, x.words, x.open ? t("Open") : t("Closed"), button(t("Open the test"), () => showTeacherTest(x.id), true)]))
		: el("p", {}, t("No tests yet.")));

	// a new test: of a lesson file, or of the lesson that is open
	const file = el("input", { type: "file", hidden: true });
	file.addEventListener("change", () => guarded(async () => {
		const f = file.files[0]; if (!f) return;
		call("open", f.name, new Uint8Array(await f.arrayBuffer()));
		await makeTest();
	}));
	const row = el("div", { className: "row" },
		el("label", { className: "button primary" }, t("New test of a lesson file…"), file));
	try { if (api && call("lesson").items.length) row.append(button(t("New test of the open lesson"), () => guarded(makeTest))); } catch (e) { /* no lesson open */ }
	box.append(el("h3", {}, t("New test")), row,
		el("p", { className: "hint" }, t("The test gets the lesson's words; students see the questions, never the answers.")));
}

async function makeTest() {
	const list = call("wordList");
	const test = await testApi("POST", "tests", { list });
	testMessage(t("Test made: %s. Now give it to students or groups.", test.title), "right");
	showTeacherTest(test.id);
}

async function showTeacherTest(id) {
	teacherDetail = id;
	const [test, results, students, groups] = await Promise.all([testApi("GET", "tests/" + id), testApi("GET", "tests/" + id + "/results"),
		testApi("GET", "users"), testApi("GET", "groups")]).catch((e) => { testMessage(e.message, "wrong"); return []; });
	if (!test) return;
	const box = $("testTeacher"); box.replaceChildren(); testShow("testTeacher");
	const reload = () => guarded(() => showTeacherTest(id));
	box.append(el("h3", {}, test.title + " (" + t("%d words", test.list.items.length) + ")"));
	box.append(el("div", { className: "row" },
		button(test.open ? t("Close for handing in") : t("Open for handing in"),
			() => guarded(async () => { await testApi("PATCH", "tests/" + id, { open: !test.open }); reload(); })),
		confirmButton(t("Delete the test"), t("Yes, delete it with its results"),
			() => guarded(async () => { await testApi("DELETE", "tests/" + id); testMessage(""); showTeacherHome(); }))));

	// who takes it
	box.append(el("h4", {}, t("Given to")));
	const assigned = el("ul", { className: "assigned" });
	for (const s of test.students) assigned.append(el("li", {}, s.name, removeButton(() => testApi("DELETE", `tests/${id}/students/${s.id}`), reload)));
	for (const g of test.groups) assigned.append(el("li", {}, t("Group %s", g.name), removeButton(() => testApi("DELETE", `tests/${id}/groups/${g.id}`), reload)));
	if (!assigned.children.length) assigned.append(el("li", { className: "hint" }, t("No one yet.")));
	const pickStudent = select(students.map((s) => [s.id, s.name]), t("A student…"));
	const pickGroup = select(groups.map((g) => [g.id, g.name]), t("A group…"));
	box.append(assigned, el("div", { className: "row" },
		pickStudent, button(t("Add"), () => pickStudent.value && guarded(async () => { await testApi("POST", `tests/${id}/students`, { userId: +pickStudent.value }); reload(); })),
		pickGroup, button(t("Add"), () => pickGroup.value && guarded(async () => { await testApi("POST", `tests/${id}/groups`, { groupId: +pickGroup.value }); reload(); }))));

	// hand-ins
	box.append(el("h4", {}, t("Handed in")));
	const handedIn = new Set(results.map((r) => r.student.id));
	const expected = new Map(test.students.map((s) => [s.id, s.name]));
	for (const g of test.groups) for (const m of (groups.find((x) => x.id === g.id) || {}).members || []) expected.set(m.id, m.name);
	const missing = [...expected].filter(([sid]) => !handedIn.has(sid)).map(([, name]) => name).sort();
	box.append(results.length ? table([t("Student"), t("Score"), t("Published"), ""], results.map((r) => [
		r.student.name, `${r.items.filter((i) => i.right).length}/${r.items.length} (${r.note}%)`, r.published ? "✓" : "–",
		button(t("Check"), () => showTeacherResult(test, r.student.id))])) : el("p", {}, t("No one has handed in yet.")));
	if (missing.length) box.append(el("p", { className: "hint" }, t("Not handed in yet: %s", missing.join(", "))));
	if (results.length) box.append(el("div", { className: "row" },
		button(t("Publish all results"), () => guarded(async () => {
			await testApi("POST", `tests/${id}/publish`, { published: true });
			testMessage(t("Published: the students can see their results."), "right"); reload();
		}), true)));
}

function removeButton(fn, then) {
	const b = el("button", { type: "button", className: "remove", title: t("Remove") }, "×");
	b.addEventListener("click", () => guarded(async () => { await fn(); then(); }));
	return b;
}

async function showTeacherResult(test, studentId) {
	const r = await testApi("GET", `tests/${test.id}/results/${studentId}`).catch((e) => { testMessage(e.message, "wrong"); });
	if (!r) return;
	const box = $("testTeacher"); box.replaceChildren(); testShow("testTeacher");
	const items = new Map(test.list.items.map((it) => [it.id, it]));
	box.append(el("h3", {}, `${test.title}: ${r.student.name}`),
		el("p", { className: "question" }, t("%d of %d right (%d%%)", r.items.filter((i) => i.right).length, r.items.length, r.note)));
	box.append(table([t("Question"), t("Right answers"), t("Given"), "", ""], r.items.map((it) => {
		const word = items.get(it.itemId) || {};
		const mark = el("span", { className: it.right ? "ok" : "no" }, it.right ? t("Right") : t("Wrong"));
		const flip = button(it.right ? t("Count as wrong") : t("Count as right"), () => guarded(async () => {
			await testApi("PATCH", `tests/${test.id}/results/${studentId}`, { itemId: it.itemId, right: !it.right });
			showTeacherResult(test, studentId);
		}));
		const given = el("span", { dir: "auto" }, it.given || "—");
		return [shown(word.questions), shown(word.answers), given, mark, flip];
	})));
	box.append(el("div", { className: "row" },
		button(t("Back to the test"), () => showTeacherTest(test.id)),
		button(r.published ? t("Published") : t("Publish this result"), () => guarded(async () => {
			await testApi("POST", `tests/${test.id}/publish`, { studentId, published: true }); showTeacherResult(test, studentId);
		}), !r.published)));
}

// ---- admins ----

async function showAdminHome() {
	const [users, groups] = await Promise.all([testApi("GET", "users"), testApi("GET", "groups")]).catch((e) => { testMessage(e.message, "wrong"); return []; });
	if (!users) return;
	const box = $("testAdmin"); box.replaceChildren(); testShow("testAdmin");
	const reload = () => guarded(showAdminHome);
	const roles = [["student", t("Student")], ["teacher", t("Teacher")], ["admin", t("Admin")]];
	const roleName = Object.fromEntries(roles);

	box.append(el("h3", {}, t("Accounts")), table([t("Name"), t("Role"), ""], users.map((u) => [u.name, roleName[u.role] || u.role,
		u.id === testUser.id ? "" : confirmButton("×", t("Remove %s", u.name), () => guarded(async () => { await testApi("DELETE", "users/" + u.id); reload(); }))])));
	const name = el("input", { placeholder: t("Name"), autocomplete: "off" });
	const password = el("input", { placeholder: t("Password"), autocomplete: "new-password" });
	const role = select(roles, t("Role…")); role.value = "student";
	box.append(el("div", { className: "row" }, name, password, role, button(t("Add"), () => guarded(async () => {
		await testApi("POST", "users", { name: name.value, password: password.value, role: role.value }); reload();
	}), true)));

	// a whole class at once: name,password[,role] per line
	const csv = el("textarea", { rows: 4, placeholder: "anna,secret\nbram,secret2\njansen,secret3,teacher" });
	box.append(el("h4", {}, t("Add many")), el("p", { className: "hint" }, t("One account per line: name, password and (if not a student) the role, separated by commas.")), csv,
		el("div", { className: "row" }, button(t("Add these accounts"), () => guarded(async () => {
			let made = 0; const failed = [];
			for (const line of csv.value.split("\n").map((l) => l.trim()).filter(Boolean)) {
				const [n, p, r] = line.split(",").map((x) => x.trim());
				try { await testApi("POST", "users", { name: n, password: p || "", role: r || "student" }); made++; } catch (e) { failed.push(`${n}: ${e.message}`); }
			}
			await showAdminHome();
			testMessage(t("%d accounts added.", made) + (failed.length ? " " + t("Not added: %s", failed.join("; ")) : ""), failed.length ? "wrong" : "right");
		}))));

	box.append(el("h3", {}, t("Groups")));
	const students = users.filter((u) => u.role === "student");
	for (const g of groups) {
		const members = el("ul", { className: "assigned" }, ...g.members.map((m) => el("li", {}, m.name,
			removeButton(() => testApi("DELETE", `groups/${g.id}/members/${m.id}`), reload))));
		const pick = select(students.filter((s) => !g.members.some((m) => m.id === s.id)).map((s) => [s.id, s.name]), t("A student…"));
		box.append(el("div", { className: "group" }, el("h4", {}, g.name), members, el("div", { className: "row" },
			pick, button(t("Add"), () => pick.value && guarded(async () => { await testApi("POST", `groups/${g.id}/members`, { userId: +pick.value }); reload(); })),
			confirmButton(t("Delete the group"), t("Yes, delete %s", g.name), () => guarded(async () => { await testApi("DELETE", "groups/" + g.id); reload(); })))));
	}
	const gname = el("input", { placeholder: t("Group name"), autocomplete: "off" });
	box.append(el("div", { className: "row" }, gname, button(t("New group"), () => guarded(async () => { await testApi("POST", "groups", { name: gname.value }); reload(); }), true)));
}
