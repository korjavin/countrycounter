// app.js — monospace edition. Vanilla JS port of docs/design/project/app.jsx.
// Drives the new UI: hero stat, quick actions, search combo, drawer, toast,
// world map. Keeps the existing /api/countries backend contract and the
// Telegram WebApp init flow.

(function () {
  // ── Continent tables (from docs/design/project/app.jsx) ──
  const CONTINENT_OF = (function () {
    const m = {};
    const add = (cont, names) => names.forEach((n) => (m[n] = cont));
    add("AF", [
      "Algeria","Angola","Benin","Botswana","Burkina Faso","Burundi","Cabo Verde",
      "Cameroon","Central African Republic","Chad","Comoros","Congo","DR Congo",
      "Congo, Republic of the","Congo, Democratic Republic of the",
      "Djibouti","Egypt","Equatorial Guinea","Eritrea","Eswatini","Ethiopia","Gabon",
      "Gambia","Ghana","Guinea","Guinea-Bissau","Ivory Coast","Cote d'Ivoire","Kenya",
      "Lesotho","Liberia","Libya","Madagascar","Malawi","Mali","Mauritania",
      "Mauritius","Morocco","Mozambique","Namibia","Niger","Nigeria","Rwanda",
      "Sao Tome and Principe","Senegal","Seychelles","Sierra Leone","Somalia",
      "South Africa","South Sudan","Sudan","Tanzania","Togo","Tunisia","Uganda",
      "Zambia","Zimbabwe",
    ]);
    add("EU", [
      "Albania","Andorra","Austria","Belarus","Belgium","Bosnia and Herzegovina",
      "Bulgaria","Croatia","Cyprus","Czechia","Czech Republic","Denmark","Estonia",
      "Finland","France","Germany","Greece","Hungary","Iceland","Ireland","Italy",
      "Kosovo","Latvia","Liechtenstein","Lithuania","Luxembourg","Malta","Moldova",
      "Monaco","Montenegro","Netherlands","North Macedonia","Norway","Poland",
      "Portugal","Romania","Russia","San Marino","Serbia","Slovakia","Slovenia",
      "Spain","Sweden","Switzerland","Ukraine","United Kingdom","Vatican City",
    ]);
    add("AS", [
      "Afghanistan","Armenia","Azerbaijan","Bahrain","Bangladesh","Bhutan","Brunei",
      "Cambodia","China","Georgia","India","Indonesia","Iran","Iraq","Israel",
      "Japan","Jordan","Kazakhstan","Kuwait","Kyrgyzstan","Laos","Lebanon",
      "Malaysia","Maldives","Mongolia","Myanmar","Nepal","North Korea","Oman",
      "Pakistan","Palestine","Palestine State","Philippines","Qatar","Saudi Arabia","Singapore",
      "South Korea","Sri Lanka","Syria","Taiwan","Tajikistan","Thailand",
      "Timor-Leste","Turkey","Turkmenistan","United Arab Emirates","Uzbekistan",
      "Vietnam","Yemen",
    ]);
    add("NA", [
      "Antigua and Barbuda","Bahamas","Barbados","Belize","Canada","Costa Rica",
      "Cuba","Dominica","Dominican Republic","El Salvador","Grenada","Guatemala",
      "Haiti","Honduras","Jamaica","Mexico","Nicaragua","Panama",
      "Saint Kitts and Nevis","Saint Lucia","Saint Vincent and the Grenadines",
      "Trinidad and Tobago","United States","United States of America",
    ]);
    add("SA", [
      "Argentina","Bolivia","Brazil","Chile","Colombia","Ecuador","Guyana",
      "Paraguay","Peru","Suriname","Uruguay","Venezuela",
    ]);
    add("OC", [
      "Australia","Fiji","Kiribati","Marshall Islands","Micronesia","Nauru",
      "New Zealand","Palau","Papua New Guinea","Samoa","Solomon Islands","Tonga",
      "Tuvalu","Vanuatu",
    ]);
    return m;
  })();

  const CONTINENT_TOTALS = { AF: 54, EU: 46, AS: 48, NA: 23, SA: 12, OC: 14 };
  const CONTINENT_LABELS = {
    NA: "N.AMER", SA: "S.AMER", EU: "EUROPE",
    AF: "AFRICA", AS: "ASIA", OC: "OCEAN.",
  };
  const CONTINENT_ORDER = ["NA", "SA", "EU", "AF", "AS", "OC"];
  const PROGRESS_CELLS = 26;
  const TOAST_MS = 2400;
  const LAST_ADDED_MS = 2000;

  function relativeTime(ts) {
    const diff = Date.now() - ts;
    const days = Math.floor(diff / 86400000);
    if (days <= 0) return "today";
    if (days === 1) return "1d ago";
    if (days < 30) return `${days}d ago`;
    const mo = Math.floor(days / 30);
    if (mo < 12) return `${mo}mo ago`;
    return `${Math.floor(mo / 12)}y ago`;
  }

  function toGeoName(name) {
    if (window.WorldMap && typeof window.WorldMap.toGeoName === "function") {
      return window.WorldMap.toGeoName(name);
    }
    return name;
  }

  document.addEventListener("DOMContentLoaded", function () {
    const tg = window.Telegram && window.Telegram.WebApp ? window.Telegram.WebApp : null;
    if (tg && typeof tg.ready === "function") tg.ready();

    let userId = null;
    if (tg && tg.initDataUnsafe && tg.initDataUnsafe.user) {
      userId = tg.initDataUnsafe.user.id;
    }

    // ── State ──
    const state = {
      visited: [],
      allCountries: [],
      sort: "recent",
      showList: true,
      query: "",
      activeSuggest: 0,
      focused: false,
      lastAddedGeo: null,
      lastAddedTimer: null,
      pendingAdds: new Set(),
      ready: false,
    };

    // ── DOM refs ──
    const $ = (id) => document.getElementById(id);
    const els = {
      visitedCount: $("visited-count"),
      visitedDenom: $("visited-denom"),
      visitedPct: $("visited-pct"),
      visitedRemaining: $("visited-remaining"),
      progressCells: $("progress-cells"),
      mapCornerTr: $("map-corner-tr"),
      lastAdded: $("last-added"),
      lastAddedText: $("last-added-text"),
      addFocusBtn: $("add-focus-btn"),
      toggleListBtn: $("toggle-list-btn"),
      sortCycleBtn: $("sort-cycle-btn"),
      shareBtn: $("share-btn"),
      countryInput: $("country-input"),
      clearQueryBtn: $("clear-query-btn"),
      addCountryBtn: $("add-country-btn"),
      suggest: $("suggest"),
      drawer: $("drawer"),
      continents: $("continents"),
      drawerCount: $("drawer-count"),
      sortSeg: $("sort-seg"),
      clearAllBtn: $("clear-all-btn"),
      countriesUl: $("countries-ul"),
      toast: $("toast"),
      toastGlyph: $("toast-glyph"),
      toastText: $("toast-text"),
      worldSvg: $("world"),
    };

    // ── Theme ──
    function applyTheme() {
      let dark = false;
      if (tg && tg.colorScheme) {
        dark = tg.colorScheme === "dark";
      } else if (window.matchMedia) {
        dark = window.matchMedia("(prefers-color-scheme: dark)").matches;
      }
      document.documentElement.setAttribute("data-theme", dark ? "dark" : "light");
    }
    applyTheme();
    if (tg && typeof tg.onEvent === "function") {
      tg.onEvent("themeChanged", applyTheme);
    }

    // ── Toast ──
    let toastTimer = null;
    function showToast(text, isError) {
      els.toast.hidden = false;
      els.toast.classList.toggle("err", !!isError);
      els.toastGlyph.textContent = isError ? "!" : "✓";
      els.toastText.textContent = text;
      if (toastTimer) clearTimeout(toastTimer);
      toastTimer = setTimeout(() => {
        els.toast.hidden = true;
      }, TOAST_MS);
    }

    // ── Derived helpers ──
    function visitedSet() {
      return new Set(state.visited.map((v) => v.name));
    }

    function totalCountries() {
      return state.allCountries.length || 195;
    }

    function sortedVisited() {
      const arr = state.visited.slice();
      if (state.sort === "alpha") {
        arr.sort((a, b) => a.name.localeCompare(b.name));
      } else if (state.sort === "recent") {
        arr.sort((a, b) => b.addedAt - a.addedAt);
      } else if (state.sort === "continent") {
        arr.sort((a, b) => {
          const ca = CONTINENT_OF[a.name] || "ZZ";
          const cb = CONTINENT_OF[b.name] || "ZZ";
          if (ca !== cb) return ca.localeCompare(cb);
          return a.name.localeCompare(b.name);
        });
      }
      return arr;
    }

    function suggestions() {
      const q = state.query.trim().toLowerCase();
      if (!q) return [];
      const starts = [];
      const contains = [];
      for (const c of state.allCountries) {
        const lc = c.toLowerCase();
        if (lc.startsWith(q)) starts.push(c);
        else if (lc.includes(q)) contains.push(c);
      }
      return starts.concat(contains).slice(0, 6);
    }

    // ── Map render ──
    function renderMap() {
      if (!window.WorldMap || !els.worldSvg) return;
      window.WorldMap.render({
        svg: els.worldSvg,
        visitedSet: visitedSet(),
        style: "solid",
        lastAddedGeo: state.lastAddedGeo,
      });
    }

    // ── Hero render ──
    function renderHero() {
      const count = state.visited.length;
      const total = totalCountries();
      const pct = total ? Math.round((count / total) * 100) : 0;
      els.visitedCount.textContent = String(count).padStart(count >= 100 ? 3 : 2, "0");
      els.visitedDenom.textContent = String(total);
      els.visitedPct.textContent = pct + "%";
      els.visitedRemaining.textContent = (total - count) + " remaining";
      els.mapCornerTr.textContent = `${count}/${total} · ${pct}%`;

      const filled = total ? Math.round((count / total) * PROGRESS_CELLS) : 0;
      const cells = els.progressCells.children;
      if (cells.length !== PROGRESS_CELLS) {
        els.progressCells.innerHTML = "";
        for (let i = 0; i < PROGRESS_CELLS; i++) {
          const s = document.createElement("span");
          s.className = "progress-cell";
          els.progressCells.appendChild(s);
        }
      }
      for (let i = 0; i < PROGRESS_CELLS; i++) {
        const c = els.progressCells.children[i];
        c.classList.toggle("on", i < filled);
      }
      els.progressCells.setAttribute("aria-label", pct + "% complete");
    }

    // ── Last-added overlay ──
    function renderLastAdded() {
      const sv = sortedVisited();
      const showLast = sv.length > 0 && state.sort === "recent";
      if (showLast) {
        els.lastAdded.hidden = false;
        els.lastAddedText.textContent = "LAST · " + sv[0].name;
      } else {
        els.lastAdded.hidden = true;
      }
    }

    // ── Quick actions render (toggle list + sort labels) ──
    function renderQuickActions() {
      // Toggle list label and glyph
      const tlBtn = els.toggleListBtn;
      const tlGlyph = tlBtn.querySelector(".glyph");
      const tlLabel = tlBtn.querySelector(".qa-label");
      if (state.showList) {
        if (tlGlyph) tlGlyph.textContent = "▮";
        if (tlLabel) tlLabel.textContent = "hide list";
      } else {
        if (tlGlyph) tlGlyph.textContent = "▯";
        if (tlLabel) tlLabel.textContent = "show list";
      }
      // Sort label and glyph
      const sBtn = els.sortCycleBtn;
      const sGlyph = sBtn.querySelector(".glyph");
      const sLabel = sBtn.querySelector(".qa-label");
      if (sGlyph) sGlyph.textContent =
        state.sort === "recent" ? "↓" : state.sort === "alpha" ? "A" : "§";
      if (sLabel) sLabel.textContent = "sort: " + state.sort;
    }

    // ── Drawer render ──
    function renderContinents() {
      // Dedupe by canonical geo name so paired aliases (e.g. "Palestine" and
      // "Palestine State", "Czechia" and "Czech Republic") count once and the
      // total never exceeds CONTINENT_TOTALS.
      const seen = { NA: new Set(), SA: new Set(), EU: new Set(), AF: new Set(), AS: new Set(), OC: new Set() };
      for (const v of state.visited) {
        const c = CONTINENT_OF[v.name];
        if (c) seen[c].add(toGeoName(v.name));
      }
      els.continents.innerHTML = "";
      for (const k of CONTINENT_ORDER) {
        const num = seen[k].size;
        const den = CONTINENT_TOTALS[k];
        const full = num >= den;
        const cell = document.createElement("div");
        cell.className = "cont" + (full ? " full" : "");
        cell.innerHTML =
          '<div class="k">' + CONTINENT_LABELS[k] + '</div>' +
          '<div class="v"><span class="num">' + num +
          '</span><span class="den">/' + den + '</span></div>';
        els.continents.appendChild(cell);
      }
    }

    function renderDrawerHead() {
      els.drawerCount.textContent =
        state.visited.length + " ENTRIES · SORT: " + state.sort.toUpperCase();
    }

    function renderSortSeg() {
      const btns = els.sortSeg.querySelectorAll("button");
      btns.forEach((b) => {
        b.classList.toggle("on", b.dataset.sort === state.sort);
      });
    }

    function renderCountryList() {
      const ul = els.countriesUl;
      ul.innerHTML = "";
      const sv = sortedVisited();
      els.clearAllBtn.hidden = sv.length === 0;
      if (sv.length === 0) {
        const empty = document.createElement("div");
        empty.className = "empty";
        empty.innerHTML =
          'No countries yet. <span class="hint">Tap the search above to add your first one.</span>';
        ul.appendChild(empty);
        return;
      }
      sv.forEach((v, i) => {
        const row = document.createElement("div");
        row.className = "country-row";

        const idx = document.createElement("span");
        idx.className = "idx t-num";
        idx.textContent = String(i + 1).padStart(2, "0");

        const name = document.createElement("span");
        name.className = "name";
        name.textContent = v.name;

        const cont = CONTINENT_OF[v.name];
        if (cont) {
          const m = document.createElement("span");
          m.className = "meta";
          m.textContent = "· " + cont;
          name.appendChild(m);
        }
        if (state.sort === "recent") {
          const m = document.createElement("span");
          m.className = "meta";
          m.textContent = "· " + relativeTime(v.addedAt);
          name.appendChild(m);
        }

        const x = document.createElement("button");
        x.className = "x";
        x.type = "button";
        x.setAttribute("aria-label", "Remove " + v.name);
        x.textContent = "✗";
        x.addEventListener("click", () => removeCountry(v.name));

        row.appendChild(idx);
        row.appendChild(name);
        row.appendChild(x);
        ul.appendChild(row);
      });
    }

    function renderDrawer() {
      els.drawer.style.display = state.showList ? "" : "none";
      if (!state.showList) return;
      renderContinents();
      renderDrawerHead();
      renderSortSeg();
      renderCountryList();
    }

    // ── Suggest dropdown render ──
    function renderSuggest() {
      const sugs = suggestions();
      els.clearQueryBtn.hidden = state.query.length === 0;
      els.addCountryBtn.disabled = state.query.trim().length === 0;
      if (!state.focused || sugs.length === 0) {
        els.suggest.hidden = true;
        els.suggest.innerHTML = "";
        return;
      }
      els.suggest.hidden = false;
      els.suggest.innerHTML = "";
      const vSet = visitedSet();
      sugs.forEach((s, i) => {
        const opt = document.createElement("div");
        opt.className = "opt" + (i === state.activeSuggest ? " active" : "");
        const name = document.createElement("span");
        name.textContent = s;
        const hint = document.createElement("span");
        hint.className = "hint";
        const cont = CONTINENT_OF[s] || "";
        hint.textContent = cont + (vSet.has(s) ? (cont ? " · IN" : "IN") : "");
        opt.appendChild(name);
        opt.appendChild(hint);
        opt.addEventListener("mousedown", (e) => {
          e.preventDefault();
          addCountry(s);
        });
        // Toggle the .active class in place rather than re-rendering, so the
        // node firing mouseenter isn't destroyed mid-event.
        opt.addEventListener("mouseenter", () => {
          state.activeSuggest = i;
          const opts = els.suggest.querySelectorAll(".opt");
          opts.forEach((o, j) => o.classList.toggle("active", j === i));
        });
        els.suggest.appendChild(opt);
      });
    }

    // ── Full render ──
    function render() {
      renderHero();
      renderLastAdded();
      renderQuickActions();
      renderDrawer();
      renderSuggest();
      renderMap();
    }

    // ── Backend wiring ──
    async function loadVisited() {
      if (!userId) return;
      try {
        const res = await fetch("/api/countries?userId=" + encodeURIComponent(userId));
        if (res.status === 404) {
          state.visited = [];
          return;
        }
        if (!res.ok) throw new Error("HTTP " + res.status);
        const names = await res.json();
        const now = Date.now();
        // No backend timestamps — synthesize ordering so recent sort is stable.
        state.visited = (Array.isArray(names) ? names : []).map((n, i) => ({
          name: n,
          addedAt: now - (names.length - i) * 1000,
        }));
      } catch (e) {
        console.error("loadVisited failed", e);
        showToast("Failed to load visited", true);
      }
    }

    async function loadAllCountries() {
      try {
        const res = await fetch("all_countries.json");
        if (!res.ok) throw new Error("HTTP " + res.status);
        state.allCountries = await res.json();
      } catch (e) {
        console.error("loadAllCountries failed", e);
        state.allCountries = [];
        showToast("Country list failed to load", true);
      }
    }

    async function postCountry(name) {
      if (!userId) return true; // standalone mode: skip backend
      try {
        const res = await fetch("/api/countries", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ userId: userId, country: name }),
        });
        return res.status === 201 || res.status === 200;
      } catch (e) {
        console.error("postCountry failed", e);
        return false;
      }
    }

    async function deleteCountryAPI(name) {
      if (!userId) return true;
      try {
        const res = await fetch("/api/countries", {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ userId: userId, country: name }),
        });
        return res.status === 200;
      } catch (e) {
        console.error("deleteCountryAPI failed", e);
        return false;
      }
    }

    // ── Actions ──
    async function addCountry(name) {
      if (!name) return;
      const vSet = visitedSet();
      if (vSet.has(name) || state.pendingAdds.has(name)) {
        showToast("Already visited: " + name, true);
        return;
      }
      state.pendingAdds.add(name);
      let ok;
      try {
        ok = await postCountry(name);
      } finally {
        state.pendingAdds.delete(name);
      }
      if (!ok) {
        showToast("Failed to add " + name, true);
        return;
      }
      // Re-check after the POST returns: a concurrent add for the same name
      // (e.g. from /location in the bot) may have already updated state.
      if (visitedSet().has(name)) return;
      state.visited.unshift({ name: name, addedAt: Date.now() });
      state.lastAddedGeo = toGeoName(name);
      state.query = "";
      els.countryInput.value = "";
      state.activeSuggest = 0;
      showToast("Added — " + name);
      render();
      // Cancel any prior pending clear so a fresh add doesn't get its
      // highlight cut short by an earlier country's timer.
      if (state.lastAddedTimer) clearTimeout(state.lastAddedTimer);
      const geoForThisAdd = state.lastAddedGeo;
      state.lastAddedTimer = setTimeout(() => {
        state.lastAddedTimer = null;
        // Only clear if no newer add has replaced the highlight in the meantime.
        if (state.lastAddedGeo === geoForThisAdd) {
          state.lastAddedGeo = null;
          renderMap();
        }
      }, LAST_ADDED_MS);
    }

    async function removeCountry(name) {
      const ok = await deleteCountryAPI(name);
      if (!ok) {
        showToast("Failed to remove " + name, true);
        return;
      }
      state.visited = state.visited.filter((v) => v.name !== name);
      showToast("Removed — " + name);
      render();
    }

    async function clearAll() {
      if (state.visited.length === 0) return;
      const names = state.visited.map((v) => v.name);
      const failed = [];
      for (const nm of names) {
        const ok = await deleteCountryAPI(nm);
        if (ok) {
          state.visited = state.visited.filter((v) => v.name !== nm);
        } else {
          failed.push(nm);
        }
      }
      if (failed.length === 0) {
        showToast("Cleared " + names.length + " countries");
      } else {
        showToast(
          "Cleared " + (names.length - failed.length) + ", " + failed.length + " failed",
          true,
        );
      }
      render();
    }

    function handleAddClick() {
      const q = state.query.trim();
      if (!q) {
        els.countryInput.focus();
        return;
      }
      const exact = state.allCountries.find(
        (c) => c.toLowerCase() === q.toLowerCase()
      );
      if (exact) { addCountry(exact); return; }
      const sugs = suggestions();
      if (sugs.length > 0) {
        addCountry(sugs[state.activeSuggest] || sugs[0]);
        return;
      }
      showToast('No match for "' + q + '"', true);
    }

    function cycleSort() {
      state.sort = state.sort === "recent"
        ? "alpha"
        : state.sort === "alpha"
          ? "continent"
          : "recent";
      render();
    }

    // ── Wiring ──
    els.addFocusBtn.addEventListener("click", () => {
      els.countryInput.focus();
    });
    els.toggleListBtn.addEventListener("click", () => {
      state.showList = !state.showList;
      renderQuickActions();
      renderDrawer();
    });
    els.sortCycleBtn.addEventListener("click", cycleSort);
    els.shareBtn.addEventListener("click", async () => {
      const text = state.visited.map((v) => v.name).join("\n");
      if (!navigator.clipboard || !navigator.clipboard.writeText) {
        showToast("Clipboard unavailable", true);
        return;
      }
      try {
        await navigator.clipboard.writeText(text);
        showToast("Copied " + state.visited.length + " countries");
      } catch (_) {
        showToast("Copy failed", true);
      }
    });

    els.countryInput.addEventListener("input", (e) => {
      state.query = e.target.value;
      state.activeSuggest = 0;
      renderSuggest();
    });
    els.countryInput.addEventListener("focus", () => {
      state.focused = true;
      renderSuggest();
    });
    els.countryInput.addEventListener("blur", () => {
      setTimeout(() => {
        state.focused = false;
        renderSuggest();
      }, 120);
    });
    els.countryInput.addEventListener("keydown", (e) => {
      const sugs = suggestions();
      if (e.key === "ArrowDown") {
        e.preventDefault();
        state.activeSuggest = Math.min(state.activeSuggest + 1, Math.max(sugs.length - 1, 0));
        renderSuggest();
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        state.activeSuggest = Math.max(state.activeSuggest - 1, 0);
        renderSuggest();
      } else if (e.key === "Enter") {
        e.preventDefault();
        handleAddClick();
      } else if (e.key === "Escape") {
        state.query = "";
        els.countryInput.value = "";
        e.target.blur();
        renderSuggest();
      }
    });
    els.clearQueryBtn.addEventListener("click", () => {
      state.query = "";
      els.countryInput.value = "";
      els.countryInput.focus();
      renderSuggest();
    });
    els.addCountryBtn.addEventListener("click", handleAddClick);

    els.sortSeg.querySelectorAll("button").forEach((b) => {
      b.addEventListener("click", () => {
        state.sort = b.dataset.sort || "recent";
        render();
      });
    });
    els.clearAllBtn.addEventListener("click", () => {
      if (window.confirm("Clear all visited countries?")) clearAll();
    });

    // ── Init ──
    (async function init() {
      await loadAllCountries();
      await loadVisited();
      render();
      state.ready = true;
    })();

    // Expose for tests / debugging.
    window.__app = { state, render, addCountry, removeCountry };
  });
})();
