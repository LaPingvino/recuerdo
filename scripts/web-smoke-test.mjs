// Smoke test of Recuerdo's web version under Node: loads web/memfs.js
// (the browser's in-memory file system), Go's wasm_exec.js and the
// WebAssembly, then opens a lesson, practises it and saves it.
//
//   GOOS=js GOARCH=wasm go build -o /tmp/recuerdo.wasm ./cmd/recuerdo-web
//   node scripts/web-smoke-test.mjs /tmp/recuerdo.wasm "$(go env GOROOT)/lib/wasm/wasm_exec.js"
import fs from "node:fs";
import vm from "node:vm";

const [wasmPath, execPath] = process.argv.slice(2);
const load = (p) => vm.runInThisContext(fs.readFileSync(p, "utf8"), { filename: p });
load(new URL("../web/memfs.js", import.meta.url).pathname);
load(execPath);

const go = new globalThis.Go();
const { instance } = await WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject);
go.run(instance);
const r = globalThis.recuerdo;
const call = (name, ...args) => {
	const out = JSON.parse(r[name](...args));
	if (out && out.error) throw new Error(`${name}: ${out.error}`);
	return out;
};
const check = (ok, what) => { if (!ok) { console.error("FAIL:", what); process.exit(1); } };

// a lesson file from its bytes (an OpenTeacher .otwd is a zip)
const sample = new URL("../testdata/legacy_files/application_x-openteachingwords.openteacher3x.otwd", import.meta.url);
const lesson = call("open", "sample.otwd", new Uint8Array(fs.readFileSync(sample)));
check(lesson.items.length > 0, "open .otwd: " + JSON.stringify(lesson));

// a typed list, practised
call("openText", "Dieren", "hond = dog\nkat = cat\n");
let st = call("start", JSON.stringify({ lessonType: "All once" }));
check(st.active && st.question === "hond", "start: " + JSON.stringify(st));
check(call("answer", "dog").right, "dog is right");
check(!call("answer", "mouse").right, "mouse is wrong");
st = call("state");
check(st.done && st.right === 1 && st.answered === 2, "state: " + JSON.stringify(st));
check(call("report").length === 2, "report");

// saved with the session, and opened again
const saved = r.save("dieren.otwd");
check(saved instanceof Uint8Array && saved[0] === 0x50 && saved[1] === 0x4b, "save gives a zip");
check(call("open", "dieren.otwd", saved).sessions === 1, "the session was saved");
check(JSON.parse(r.start("not json")).error, "bad options give an error");
console.log("web version OK:", lesson.items.length, "words from the sample;", saved.length, "bytes saved");
process.exit(0);
