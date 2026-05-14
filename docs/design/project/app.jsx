// app.jsx — Country counter, monospace edition.

const SEED_VISITED = ["United States", "Canada", "Mexico", "Iceland", "United Kingdom", "France", "Spain", "Italy", "Portugal", "Netherlands", "Greece", "Japan", "Thailand", "Vietnam", "Australia", "New Zealand"];

// Reverse of toGeoName — minimal map for click-to-add: topojson name -> canonical name (matches all_countries.json).
const GEO_TO_CANON = {
  "United States of America": "United States",
  "Czech Republic": "Czechia",
  "Cape Verde": "Cabo Verde",
  "Ivory Coast": "Cote d'Ivoire",
  "Democratic Republic of the Congo": "DR Congo",
  "Republic of the Congo": "Congo",
  "Swaziland": "Eswatini",
  "East Timor": "Timor-Leste",
  "Macedonia": "North Macedonia",
  "Vatican": "Vatican City",
  "The Bahamas": "Bahamas",
  "Guinea Bissau": "Guinea-Bissau",
  "São Tomé and Principe": "Sao Tome and Principe",
  "Republic of Serbia": "Serbia",
  "United Republic of Tanzania": "Tanzania",
  "Somaliland": null, // not a UN member; skip
  "Kosovo": null,     // disputed; not in all_countries
  "N. Cyprus": null,
};

function geoToCanon(geoName) {
  if (geoName in GEO_TO_CANON) return GEO_TO_CANON[geoName];
  return geoName;
}

// Continent mapping (canonical names). Compact — covers most countries.
const CONTINENT_OF = (() => {
  const m = {};
  const add = (cont, names) => names.forEach(n => m[n] = cont);
  add("AF", ["Algeria","Angola","Benin","Botswana","Burkina Faso","Burundi","Cabo Verde","Cameroon","Central African Republic","Chad","Comoros","Congo","DR Congo","Djibouti","Egypt","Equatorial Guinea","Eritrea","Eswatini","Ethiopia","Gabon","Gambia","Ghana","Guinea","Guinea-Bissau","Ivory Coast","Cote d'Ivoire","Kenya","Lesotho","Liberia","Libya","Madagascar","Malawi","Mali","Mauritania","Mauritius","Morocco","Mozambique","Namibia","Niger","Nigeria","Rwanda","Sao Tome and Principe","Senegal","Seychelles","Sierra Leone","Somalia","South Africa","South Sudan","Sudan","Tanzania","Togo","Tunisia","Uganda","Zambia","Zimbabwe"]);
  add("EU", ["Albania","Andorra","Austria","Belarus","Belgium","Bosnia and Herzegovina","Bulgaria","Croatia","Cyprus","Czechia","Czech Republic","Denmark","Estonia","Finland","France","Germany","Greece","Hungary","Iceland","Ireland","Italy","Kosovo","Latvia","Liechtenstein","Lithuania","Luxembourg","Malta","Moldova","Monaco","Montenegro","Netherlands","North Macedonia","Norway","Poland","Portugal","Romania","Russia","San Marino","Serbia","Slovakia","Slovenia","Spain","Sweden","Switzerland","Ukraine","United Kingdom","Vatican City"]);
  add("AS", ["Afghanistan","Armenia","Azerbaijan","Bahrain","Bangladesh","Bhutan","Brunei","Cambodia","China","Georgia","India","Indonesia","Iran","Iraq","Israel","Japan","Jordan","Kazakhstan","Kuwait","Kyrgyzstan","Laos","Lebanon","Malaysia","Maldives","Mongolia","Myanmar","Nepal","North Korea","Oman","Pakistan","Palestine","Philippines","Qatar","Saudi Arabia","Singapore","South Korea","Sri Lanka","Syria","Taiwan","Tajikistan","Thailand","Timor-Leste","Turkey","Turkmenistan","United Arab Emirates","Uzbekistan","Vietnam","Yemen"]);
  add("NA", ["Antigua and Barbuda","Bahamas","Barbados","Belize","Canada","Costa Rica","Cuba","Dominica","Dominican Republic","El Salvador","Grenada","Guatemala","Haiti","Honduras","Jamaica","Mexico","Nicaragua","Panama","Saint Kitts and Nevis","Saint Lucia","Saint Vincent and the Grenadines","Trinidad and Tobago","United States"]);
  add("SA", ["Argentina","Bolivia","Brazil","Chile","Colombia","Ecuador","Guyana","Paraguay","Peru","Suriname","Uruguay","Venezuela"]);
  add("OC", ["Australia","Fiji","Kiribati","Marshall Islands","Micronesia","Nauru","New Zealand","Palau","Papua New Guinea","Samoa","Solomon Islands","Tonga","Tuvalu","Vanuatu"]);
  return m;
})();

const CONTINENT_TOTALS = { AF:54, EU:46, AS:47, NA:23, SA:12, OC:14 };
const CONTINENT_LABELS = { NA:"N.AMER", SA:"S.AMER", EU:"EUROPE", AF:"AFRICA", AS:"ASIA", OC:"OCEAN." };

// ─────────────────────────────────────────────────────────────
function App() {
  const [t, setTweak] = useTweaks(TWEAK_DEFAULTS);
  const [allCountries, setAllCountries] = React.useState([]);
  const [visited, setVisited] = React.useState(() => {
    // recent-first order
    const seeded = SEED_VISITED.map((n, i) => ({ name: n, addedAt: Date.now() - (SEED_VISITED.length - i) * 86400000 }));
    return seeded;
  });
  const [query, setQuery] = React.useState("");
  const [sort, setSort] = React.useState("recent"); // "recent" | "alpha" | "continent"
  const [showList, setShowList] = React.useState(true);
  const [toast, setToast] = React.useState(null); // {text, err}
  const [lastAddedGeo, setLastAddedGeo] = React.useState(null);
  const [activeSuggest, setActiveSuggest] = React.useState(0);
  const [focused, setFocused] = React.useState(false);
  const inputRef = React.useRef(null);

  // Theme
  React.useEffect(() => {
    document.documentElement.setAttribute("data-theme", t.dark ? "dark" : "light");
  }, [t.dark]);

  // Load all countries
  React.useEffect(() => {
    fetch("data/all_countries.json")
      .then(r => r.json())
      .then(setAllCountries)
      .catch(() => setAllCountries([]));
  }, []);

  const visitedSet = React.useMemo(() => new Set(visited.map(v => v.name)), [visited]);

  function showToast(text, err = false) {
    setToast({ text, err });
    setTimeout(() => setToast(null), 2400);
  }

  function add(name) {
    if (!name) return;
    if (visitedSet.has(name)) {
      showToast(`Already visited: ${name}`, true);
      return;
    }
    setVisited(prev => [{ name, addedAt: Date.now() }, ...prev]);
    setLastAddedGeo(toGeoName(name));
    setQuery("");
    showToast(`Added — ${name}`);
    setTimeout(() => setLastAddedGeo(null), 2000);
  }

  function remove(name) {
    setVisited(prev => prev.filter(v => v.name !== name));
    showToast(`Removed — ${name}`);
  }

  function toggleFromMap(geoName) {
    const canon = geoToCanon(geoName);
    if (canon === null) {
      showToast(`${geoName} — not tracked`, true);
      return;
    }
    if (visitedSet.has(canon)) remove(canon);
    else add(canon);
  }

  // Suggestions
  const suggestions = React.useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    return allCountries
      .filter(c => c.toLowerCase().startsWith(q) || c.toLowerCase().includes(q))
      .slice(0, 6);
  }, [query, allCountries]);

  React.useEffect(() => setActiveSuggest(0), [query]);

  function handleAddClick() {
    if (!query.trim()) {
      inputRef.current && inputRef.current.focus();
      return;
    }
    // Use top suggestion if exact match isn't in list
    const exact = allCountries.find(c => c.toLowerCase() === query.trim().toLowerCase());
    if (exact) return add(exact);
    if (suggestions.length > 0) return add(suggestions[activeSuggest] || suggestions[0]);
    showToast(`No match for "${query}"`, true);
  }

  function handleKeyDown(e) {
    if (e.key === "ArrowDown") { e.preventDefault(); setActiveSuggest(i => Math.min(i + 1, suggestions.length - 1)); }
    else if (e.key === "ArrowUp") { e.preventDefault(); setActiveSuggest(i => Math.max(i - 1, 0)); }
    else if (e.key === "Enter") { e.preventDefault(); handleAddClick(); }
    else if (e.key === "Escape") { setQuery(""); e.target.blur(); }
  }

  function clearAll() {
    if (visited.length === 0) return;
    setVisited([]);
    showToast(`Cleared ${visited.length} countries`);
  }

  // Sorted visited list
  const sortedVisited = React.useMemo(() => {
    const arr = [...visited];
    if (sort === "alpha") arr.sort((a, b) => a.name.localeCompare(b.name));
    else if (sort === "recent") arr.sort((a, b) => b.addedAt - a.addedAt);
    else if (sort === "continent") arr.sort((a, b) => {
      const ca = CONTINENT_OF[a.name] || "ZZ", cb = CONTINENT_OF[b.name] || "ZZ";
      if (ca !== cb) return ca.localeCompare(cb);
      return a.name.localeCompare(b.name);
    });
    return arr;
  }, [visited, sort]);

  // Continent breakdown
  const contStats = React.useMemo(() => {
    const counts = { NA:0, SA:0, EU:0, AF:0, AS:0, OC:0 };
    for (const v of visited) {
      const c = CONTINENT_OF[v.name];
      if (c) counts[c]++;
    }
    return counts;
  }, [visited]);

  // 195 UN-recognized country total. all_countries.json has 195.
  const TOTAL = allCountries.length || 195;
  const pct = Math.round((visited.length / TOTAL) * 100);
  const progressCells = 26;
  const filledCells = Math.round((visited.length / TOTAL) * progressCells);

  const lastEntry = sortedVisited.length > 0 && sort === "recent" ? sortedVisited[0] : null;

  // ── Render ──
  return (
    <div className="app" data-screen-label="Country Counter">
      {/* Header */}
      <div className="app-header">
        <div className="brand">
          <span className="bracket">[</span>countrycounter<span className="bracket">]</span>
        </div>
        <div className="meta">
          <span style={{ marginRight: 8 }}>v0.4</span>
          <span>· LIVE</span>
        </div>
      </div>

      {/* Quick actions strip */}
      <div className="qa">
        <button onClick={() => inputRef.current && inputRef.current.focus()} title="Add country">
          <span className="glyph">+</span>
          <span>add</span>
        </button>
        <button onClick={() => setShowList(s => !s)} title="Toggle list">
          <span className="glyph">{showList ? "▮" : "▯"}</span>
          <span>{showList ? "hide list" : "show list"}</span>
        </button>
        <button onClick={() => setSort(s => s === "recent" ? "alpha" : s === "alpha" ? "continent" : "recent")} title="Sort">
          <span className="glyph">{sort === "recent" ? "↓" : sort === "alpha" ? "A" : "§"}</span>
          <span>sort: {sort}</span>
        </button>
        <button onClick={() => {
          const text = visited.map(v => v.name).join("\n");
          navigator.clipboard && navigator.clipboard.writeText(text);
          showToast(`Copied ${visited.length} countries`);
        }} title="Export to clipboard">
          <span className="glyph">↗</span>
          <span>share</span>
        </button>
      </div>

      {/* Hero stat */}
      <div className="hero">
        <div className="hero-marker">§ STATUS · 01</div>
        <div className="row">
          <span className="big t-num">{String(visited.length).padStart(2, "0")}</span>
          <span className="slash">/</span>
          <span className="denom t-num">{TOTAL}</span>
        </div>
        <div className="caption">
          <span className="strong">{pct}%</span> OF EARTH ·{" "}
          <span>{TOTAL - visited.length} remaining</span>
        </div>
        <div className="progress" aria-label={`${pct}% complete`}>
          {Array.from({ length: progressCells }).map((_, i) => (
            <span key={i} className={"progress-cell" + (i < filledCells ? " on" : "")} />
          ))}
        </div>
      </div>

      {/* Map */}
      <div className="map-wrap" style={{ height: 252 }}>
        <WorldMap
          visitedSet={visitedSet}
          style={t.mapStyle}
          lastAddedGeo={lastAddedGeo}
          height={252}
        />
        <div className="map-corner tl">EQUAL EARTH · 1:155M</div>
        <div className="map-corner tr">{visited.length}/{TOTAL} · {pct}%</div>
        {lastEntry && (
          <div className="last-added">
            <span className="arrow">→</span>LAST · {lastEntry.name}
          </div>
        )}
      </div>

      {/* Action row */}
      <div className="actions">
        <div className="combo">
          {focused && suggestions.length > 0 && (
            <div className="suggest">
              {suggestions.map((s, i) => (
                <div
                  key={s}
                  className={"opt" + (i === activeSuggest ? " active" : "")}
                  onMouseDown={(e) => { e.preventDefault(); add(s); }}
                  onMouseEnter={() => setActiveSuggest(i)}
                >
                  <span>{s}</span>
                  <span className="hint">{CONTINENT_OF[s] || ""}{visitedSet.has(s) ? " · IN" : ""}</span>
                </div>
              ))}
            </div>
          )}
          <span className="prefix">›</span>
          <input
            ref={inputRef}
            value={query}
            onChange={e => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            onFocus={() => setFocused(true)}
            onBlur={() => setTimeout(() => setFocused(false), 100)}
            placeholder="search country..."
            className="with-prefix"
            spellCheck="false"
            autoComplete="off"
          />
          {query && (
            <button className="clear" onClick={() => setQuery("")} aria-label="Clear">×</button>
          )}
        </div>
        <button className="btn primary" onClick={handleAddClick} disabled={!query.trim()}>
          + Add
        </button>
      </div>

      {/* Drawer */}
      {showList && (
        <div className="drawer">
          {/* Continent breakdown */}
          <div className="continents">
            {["NA","SA","EU","AF","AS","OC"].map(k => {
              const num = contStats[k];
              const den = CONTINENT_TOTALS[k];
              const full = num === den;
              return (
                <div key={k} className={"cont" + (full ? " full" : "")}>
                  <div className="k">{CONTINENT_LABELS[k]}</div>
                  <div className="v"><span className="num">{num}</span><span className="den">/{den}</span></div>
                </div>
              );
            })}
          </div>

          <div className="drawer-head">
            <div className="title">§ VISITED</div>
            <div className="count">{visited.length} ENTRIES · SORT: {sort.toUpperCase()}</div>
          </div>

          <div className="row-toolbar">
            <div className="seg">
              <button className={sort === "recent" ? "on" : ""} onClick={() => setSort("recent")}>recent</button>
              <button className={sort === "alpha" ? "on" : ""} onClick={() => setSort("alpha")}>A–Z</button>
              <button className={sort === "continent" ? "on" : ""} onClick={() => setSort("continent")}>region</button>
            </div>
            <div style={{ flex: 1 }} />
            {visited.length > 0 && (
              <button className="btn ghost" style={{ fontSize: 10, padding: "2px 8px", boxShadow: "none" }} onClick={() => {
                if (confirm("Clear all visited countries?")) clearAll();
              }}>clear</button>
            )}
          </div>

          {sortedVisited.length === 0 ? (
            <div className="empty">
              No countries yet. <span className="hint">Tap the map or search above to add your first one.</span>
            </div>
          ) : (
            sortedVisited.map((v, i) => {
              const cont = CONTINENT_OF[v.name];
              const ago = relativeTime(v.addedAt);
              return (
                <div key={v.name} className="country-row">
                  <span className="idx t-num">{String(i + 1).padStart(2, "0")}</span>
                  <span className="name">
                    {v.name}
                    {cont && <span className="meta">· {cont}</span>}
                    {sort === "recent" && <span className="meta">· {ago}</span>}
                  </span>
                  <button className="x" onClick={() => remove(v.name)} aria-label={`Remove ${v.name}`}>✗</button>
                </div>
              );
            })
          )}
        </div>
      )}

      {toast && (
        <div className={"toast" + (toast.err ? " err" : "")}>
          <span className="glyph">{toast.err ? "!" : "✓"}</span>
          <span>{toast.text}</span>
        </div>
      )}

      {/* Tweaks */}
      <TweaksPanel>
        <TweakSection label="Theme" />
        <TweakRadio label="Mode" value={t.dark ? "dark" : "light"}
                    options={["light", "dark"]}
                    onChange={(v) => setTweak('dark', v === "dark")} />
        <TweakSection label="Map" />
        <TweakRadio label="Fill" value={t.mapStyle}
                    options={["solid", "dotfill", "outline"]}
                    onChange={(v) => setTweak('mapStyle', v)} />
        <TweakSection label="Data" />
        <TweakButton onClick={() => {
          // Seed with extended set
          const more = ["Germany", "Switzerland", "Argentina", "Brazil", "South Africa", "Egypt", "Morocco", "India", "Indonesia", "South Korea"];
          const now = Date.now();
          setVisited(prev => {
            const have = new Set(prev.map(p => p.name));
            const additions = more.filter(n => !have.has(n)).map((n, i) => ({ name: n, addedAt: now - i * 1000 }));
            return [...additions, ...prev];
          });
          showToast("Seeded extra countries");
        }}>+ seed sample data</TweakButton>
        <TweakButton onClick={() => { setVisited([]); showToast("Cleared all"); }}>clear all</TweakButton>
      </TweaksPanel>
    </div>
  );
}

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

window.App = App;
