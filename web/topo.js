// Map lessons (topography) on the web page: the lesson's picture as a map
// Leaflet zooms and pans (loaded from unpkg when a map lesson opens), its
// places as dots. The Words tab edits them: click the map to put a place,
// drag one to move it, click one to rename or remove it. Practising, the
// asked place is marked to name ("Name the marked place"), or its name
// shown to click on the map ("Find the named place"). The places are in
// the picture's pixels, as the desktop has them (Go: internal/webapi/topo.go).
"use strict";

const LEAFLET = "https://unpkg.com/leaflet@1.9.4/dist/";
let leaflet = null; // the promise of Leaflet loaded

function loadLeaflet() {
	if (!leaflet) {
		leaflet = new Promise((ok, fail) => {
			const css = document.createElement("link");
			css.rel = "stylesheet"; css.href = LEAFLET + "leaflet.css";
			document.head.append(css);
			const js = document.createElement("script");
			js.src = LEAFLET + "leaflet.js";
			js.onload = () => ok(globalThis.L);
			js.onerror = () => fail(new Error(t("The map could not load: it needs the internet the first time.")));
			document.head.append(js);
		});
	}
	return leaflet;
}

// A map in el of the lesson's picture: Leaflet's flat coordinates, the
// picture's pixels with y downwards (Leaflet's lat is -y).
class LessonMap {
	constructor(L, el, url, size) {
		this.L = L; this.size = size;
		this.map = L.map(el, { crs: L.CRS.Simple, minZoom: -4, maxZoom: 4, attributionControl: false });
		const bounds = [[-size.h, 0], [0, size.w]];
		L.imageOverlay(url, bounds).addTo(this.map);
		this.map.fitBounds(bounds);
		this.layer = L.layerGroup().addTo(this.map);
	}
	latLng(x, y) { return this.L.latLng(-y, x); }
	pixel(ll) { return { x: Math.round(ll.lng), y: Math.round(-ll.lat) }; }
	dot(x, y, cls, draggable) {
		const icon = this.L.divIcon({ className: "place " + cls, iconSize: [12, 12] });
		return this.L.marker(this.latLng(x, y), { icon, draggable, keyboard: true }).addTo(this.layer);
	}
	// radius is how near a click counts as a place: 15 screen pixels
	radius() { return 15 / Math.pow(2, this.map.getZoom()); }
}

// pictureSize is the picture's size in pixels.
function pictureSize(url) {
	return new Promise((ok) => {
		const img = new Image();
		img.onload = () => ok({ w: img.naturalWidth, h: img.naturalHeight });
		img.src = url;
	});
}

const topo = {
	editor: null, practice: null, size: null,

	// show shows a lesson: a map lesson's map in the Words tab, its
	// practice choice; a words lesson as before.
	async show(l) {
		const isMap = l.kind === "topo";
		$("mapEditor").hidden = $("mapHint").hidden = $("topoModeLabel").hidden = !isMap;
		$("practiceMap").hidden = true;
		for (const o of $("format").options) o.hidden = (o.value === "ottp") !== isMap; // a map's file keeps the map
		$("format").value = isMap ? "ottp" : "otwd";
		if (this.practice) { this.practice.map.remove(); this.practice = null; } // another lesson's picture
		document.querySelector("table.words").hidden = isMap;
		if (!isMap) return;
		try {
			const L = await loadLeaflet();
			this.size = await pictureSize(l.map);
			if (this.editor) this.editor.map.remove();
			this.editor = new LessonMap(L, $("mapEditor"), l.map, this.size);
			this.editor.map.on("click", (e) => this.add(e.latlng));
			this.places(l);
		} catch (e) { showError(e); }
	},

	// places puts the lesson's places on the editor's map.
	places(l) {
		const m = this.editor;
		m.layer.clearLayers();
		for (const it of l.items) {
			if (it.x == null) continue;
			const dot = m.dot(it.x, it.y, "", true);
			dot.bindTooltip(it.question, { permanent: true, direction: "right", offset: [6, 0] });
			dot.on("dragend", () => {
				const p = m.pixel(dot.getLatLng());
				call("setPlace", it.id, it.question, p.x, p.y);
				lesson = call("lesson");
			});
			dot.on("click", () => this.edit(it, dot));
		}
	},

	// add asks for a new place's name where the map was clicked.
	add(ll) {
		const p = this.editor.pixel(ll);
		if (p.x < 0 || p.y < 0 || p.x > this.size.w || p.y > this.size.h) return;
		const name = prompt(t("The place's name"));
		if (!name || !name.trim()) return;
		call("addPlace", name.trim(), p.x, p.y);
		lesson = call("lesson");
		this.places(lesson);
	},

	// edit renames or removes a place.
	edit(it, dot) {
		const name = prompt(t("The place's name (empty: remove it)"), it.question);
		if (name === null) return;
		if (!name.trim()) call("removeItem", it.id);
		else {
			const p = this.editor.pixel(dot.getLatLng());
			call("setPlace", it.id, name.trim(), p.x, p.y);
		}
		lesson = call("lesson");
		this.places(lesson);
	},

	// state shows a practice question on the map; it reports whether it did
	// (it does for map lessons).
	state(st) {
		if (!this.size || !globalThis.L) return false;
		const find = $("topoMode").value === "find";
		text($("counter"), `${st.asked + 1} / ${st.total}`);
		$("progressBar").style.width = `${st.total ? (100 * st.asked) / st.total : 0}%`;
		$("inmind").hidden = true; $("answerBuilder").hidden = true;
		const hint = $("modeHint"); hint.className = "mode-hint"; text(hint, "");
		$("practiceMap").hidden = false;
		if (!this.practice) {
			this.practice = new LessonMap(L, $("practiceMap"), lesson.map, this.size);
			this.practice.map.on("click", (e) => this.clicked(e.latlng));
		}
		const m = this.practice;
		m.map.invalidateSize();
		m.layer.clearLayers();
		for (const it of lesson.items) {
			if (it.x == null) continue;
			const asked = !find && it.id === st.itemId;
			const dot = m.dot(it.x, it.y, asked ? "asked" : "", false);
			if (find) dot.on("click", () => answerWith(it.question));
		}
		if (find) {
			text($("question"), t("Where is %s?", st.question));
			$("answerForm").hidden = true;
		} else {
			text($("question"), t("Which place is marked?"));
			$("answerForm").hidden = false;
			$("answer").value = ""; $("answer").disabled = false; $("answer").focus();
		}
		return true;
	},

	// clicked answers with the place nearest a click on the map (finding).
	clicked(ll) {
		if ($("topoMode").value !== "find") return;
		const p = this.practice.pixel(ll);
		answerWith(api.nearestPlace(p.x, p.y, this.practice.radius()) || "—");
	},
};
globalThis.topo = topo;

// refit fits the maps in sight: one made in a hidden tab has no size
// (app.js's showTab calls it).
topo.refit = function () {
	for (const m of [this.editor, this.practice]) {
		if (m && m.map.getContainer().offsetWidth) {
			m.map.invalidateSize();
			if (!m.fitted) { m.map.fitBounds([[-this.size.h, 0], [0, this.size.w]]); m.fitted = true; }
		}
	}
};

// a new map lesson from a picture of a map
$("newMap").addEventListener("change", async () => {
	const file = $("newMap").files[0];
	if (!file) return;
	try {
		const bytes = new Uint8Array(await file.arrayBuffer());
		const title = file.name.replace(/\.[^.]*$/, "");
		showLesson(call("newTopo", title, bytes));
		showTab("words");
	} catch (e) { showError(e); }
	$("newMap").value = "";
});
