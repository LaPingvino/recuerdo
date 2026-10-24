// Media lessons on the web page: pictures, sounds, videos, texts and
// websites, each with a question and an answer. The Words tab lists them
// (add files or web addresses, name them, give them a question and an
// answer); practising shows the item and asks for its answer. Embedded
// files come from Go (internal/webapi/media.go) as blobs; online videos
// play in the site's own player.
"use strict";

const blobs = new Map(); // an embedded file's name → its blob URL

function mediaURL(it) {
	if (!it.file) return it.media;
	if (!blobs.has(it.file)) {
		const f = api.mediaFile(it.file);
		blobs.set(it.file, f ? URL.createObjectURL(new Blob([f.bytes], { type: f.type })) : "");
	}
	return blobs.get(it.file);
}

// embedURL is an online video's player address ("" if not one).
function embedURL(kind, address) {
	let m;
	switch (kind) {
	case "YouTube video":
		m = address.match(/(?:[?&]v=|youtu\.be\/)([\w-]+)/);
		return m ? `https://www.youtube-nocookie.com/embed/${m[1]}` : "";
	case "Vimeo video":
		m = address.match(/vimeo\.com\/(?:video\/)?(\d+)/);
		return m ? `https://player.vimeo.com/video/${m[1]}` : "";
	case "Dailymotion video":
		m = address.match(/dailymotion\.com\/video\/([\w]+)/);
		return m ? `https://www.dailymotion.com/embed/video/${m[1]}` : "";
	}
	return "";
}

// mediaView is the element that shows an item: small in the list, large
// when practising.
function mediaView(it) {
	const url = mediaURL(it);
	const box = document.createElement("div"); box.className = "media-view";
	if (!url) { text(box, t("The %s %s is missing from the lesson.", t(it.mediaKind), it.file)); return box; }
	let el;
	switch (it.mediaKind) {
	case "picture":
		el = document.createElement("img"); el.src = url; el.alt = it.name; break;
	case "sound":
		el = document.createElement("audio"); el.src = url; el.controls = true; break;
	case "video":
		el = document.createElement("video"); el.src = url; el.controls = true; break;
	case "text":
		el = document.createElement("pre");
		fetch(url).then((r) => r.text()).then((s) => text(el, s));
		break;
	default: {
		const player = embedURL(it.mediaKind, url);
		if (player) {
			el = document.createElement("iframe"); el.src = player; el.allowFullscreen = true;
			el.allow = "autoplay; encrypted-media; picture-in-picture";
		} else {
			// websites mostly refuse to be shown in another page: a link
			el = document.createElement("a"); el.href = url; el.target = "_blank"; el.rel = "noopener";
			if (it.file) el.download = it.file.replace(/^resources\//, "");
			el.className = "button"; text(el, it.file ? t("Download %s", el.download) : t("Open %s", url));
		}
	}
	}
	box.append(el);
	return box;
}

const media = {
	// show shows a lesson: a media lesson's items in the Words tab.
	show(l) {
		const isMedia = l.kind === "media";
		for (const u of blobs.values()) if (u) URL.revokeObjectURL(u);
		blobs.clear();
		$("mediaEditor").hidden = !isMedia;
		$("practiceMedia").hidden = true;
		if (!isMedia) return;
		document.querySelector("table.words").hidden = true;
		this.items(l);
	},

	// items lists the lesson's items to edit.
	items(l) {
		const list = $("mediaItems"); list.replaceChildren();
		for (const it of l.items) {
			const row = document.createElement("div"); row.className = "media-item";
			row.append(mediaView(it));
			const fields = document.createElement("div"); fields.className = "media-fields";
			const inputs = [["name", t("Name")], ["question", t("Question (may be empty)")], ["answer", t("Answer")]].map(([key, label]) => {
				const input = document.createElement("input");
				input.value = it[key] || ""; input.placeholder = label; input.title = label; input.dir = "auto";
				fields.append(input);
				return input;
			});
			const save = () => {
				try {
					call("setMedia", it.id, ...inputs.map((i) => i.value));
					lesson = call("lesson");
				} catch (e) { alert(e.message); }
			};
			for (const i of inputs) i.addEventListener("change", save);
			const del = document.createElement("button"); del.className = "remove"; del.title = t("Remove"); del.textContent = "×";
			del.addEventListener("click", () => { call("removeItem", it.id); lesson = call("lesson"); this.items(lesson); });
			row.append(fields, del);
			list.append(row);
		}
		if (!l.items.length) { const p = document.createElement("p"); p.className = "hint"; text(p, t("Add pictures, sounds, videos or web addresses.")); list.append(p); }
	},

	// add embeds files, each a new item.
	async add(files) {
		try {
			for (const f of files) call("addMediaFile", f.name, new Uint8Array(await f.arrayBuffer()));
		} catch (e) { alert(e.message); }
		lesson = call("lesson");
		this.items(lesson);
	},

	// state shows the asked item above its question.
	state(st) {
		const box = $("practiceMedia");
		const it = lesson.items.find((i) => i.id === st.itemId);
		box.hidden = !it;
		box.replaceChildren();
		if (it) box.append(mediaView(it));
	},
};
globalThis.media = media;

$("newMedia").addEventListener("change", async () => {
	const files = [...$("newMedia").files];
	if (!files.length) return;
	try {
		showLesson(call("newMedia", t("Media")));
		await media.add(files);
		showTab("words");
	} catch (e) { showError(e); }
	$("newMedia").value = "";
});
$("addMedia").addEventListener("change", async () => {
	await media.add([...$("addMedia").files]);
	$("addMedia").value = "";
});
$("mediaAddressForm").addEventListener("submit", (e) => {
	e.preventDefault();
	try {
		call("addMediaAddress", $("mediaAddress").value);
		$("mediaAddress").value = "";
		lesson = call("lesson");
		media.items(lesson);
	} catch (err) { alert(err.message); }
});
