// world-map.jsx — SVG world map using d3-geo + topojson, on-brand monospace.

const NAME_ALIASES = {
  // map canonical (all_countries.json) -> topojson properties.name
  "United States": "United States of America",
  "Czechia": "Czech Republic",
  "Czech Republic": "Czech Republic",
  "Cabo Verde": "Cape Verde",
  "Cote d'Ivoire": "Ivory Coast",
  "Côte d'Ivoire": "Ivory Coast",
  "Congo, Democratic Republic of the": "Democratic Republic of the Congo",
  "DR Congo": "Democratic Republic of the Congo",
  "Congo": "Republic of the Congo",
  "Eswatini": "Swaziland",
  "Myanmar": "Myanmar",
  "Timor-Leste": "East Timor",
  "Macedonia": "Macedonia",
  "North Macedonia": "Macedonia",
  "Bosnia and Herzegovina": "Bosnia and Herzegovina",
  "Vatican City": "Vatican",
  "Holy See": "Vatican",
  "Bahamas": "The Bahamas",
  "Gambia": "Gambia",
  "Guinea-Bissau": "Guinea Bissau",
  "Sao Tome and Principe": "São Tomé and Principe",
  "Serbia": "Republic of Serbia",
  "Tanzania": "United Republic of Tanzania",
};

// Inverse lookup helper — returns topojson name for a canonical name
function toGeoName(n) {
  return NAME_ALIASES[n] || n;
}

function WorldMap({
  visitedSet, // Set<string> of canonical country names
  onToggle,   // (canonicalName) => void
  style = "solid", // "solid" | "dotfill" | "outline"
  lastAddedGeo = null, // topojson properties.name to highlight briefly
  height = 280,
}) {
  const [geo, setGeo] = React.useState(null);
  const [hovered, setHovered] = React.useState(null);
  const svgRef = React.useRef(null);

  // Wait for UMD globals (topojson, d3) — set up by index.html script tags
  const [paths, setPaths] = React.useState(null);
  React.useEffect(() => {
    let cancelled = false;
    function ready() {
      return typeof window.topojson !== "undefined"
        && typeof window.d3 !== "undefined"
        && typeof window.d3.geoEqualEarth === "function";
    }
    async function go() {
      // Wait for UMD libs to attach
      let tries = 0;
      while (!ready() && tries < 100) {
        await new Promise(r => setTimeout(r, 50));
        tries++;
      }
      if (cancelled || !ready()) {
        if (!ready()) console.error("d3-geo / topojson-client failed to load");
        return;
      }
      try {
        const world = await fetch("https://cdn.jsdelivr.net/npm/world-atlas@2.0.2/countries-110m.json").then(r => r.json());
        if (cancelled) return;
        const fc = window.topojson.feature(world, world.objects.countries);
        setGeo(fc);
        const w = 800, h = 460;
        const proj = window.d3.geoEqualEarth().fitSize([w - 12, h - 12], fc);
        const path = window.d3.geoPath(proj);
        const out = fc.features.map(f => ({
          name: f.properties.name,
          d: path(f),
        })).filter(p => p.d);
        const graticule = window.d3.geoGraticule10();
        setPaths({ countries: out, graticule: path(graticule), sphere: path({type: "Sphere"}), w, h });
      } catch (e) {
        console.error("Map load failed", e);
      }
    }
    go();
    return () => { cancelled = true; };
  }, []);

  // Build a quick lookup: topojson name -> visited?
  const visitedGeoNames = React.useMemo(() => {
    const s = new Set();
    if (visitedSet) for (const n of visitedSet) s.add(toGeoName(n));
    return s;
  }, [visitedSet]);

  // Reverse map for click: topojson name -> canonical (best effort)
  // We just pass the topojson name back; the parent handles matching.
  function handleClick(geoName) {
    onToggle && onToggle(geoName);
  }

  return (
    <svg
      ref={svgRef}
      className={`world style-${style}`}
      viewBox={paths ? `0 0 ${paths.w} ${paths.h}` : "0 0 800 460"}
      preserveAspectRatio="xMidYMid meet"
      style={{ height }}
    >
      <defs>
        <pattern id="dotfill" patternUnits="userSpaceOnUse" width="4" height="4">
          <rect width="4" height="4" fill="var(--bg-0)" />
          <circle cx="2" cy="2" r="0.9" fill="var(--fg-0)" />
        </pattern>
      </defs>
      {paths && <path className="sphere" d={paths.sphere} />}
      {paths && <path className="graticule" d={paths.graticule} />}
      {paths && paths.countries.map((c, i) => {
        const visited = visitedGeoNames.has(c.name);
        const isLast = lastAddedGeo === c.name;
        return (
          <path
            key={c.name + i}
            className={`country${visited ? " visited" : ""}${isLast ? " just-added" : ""}`}
            d={c.d}
          >
            <title>{c.name}{visited ? " — visited" : ""}</title>
          </path>
        );
      })}
      {/* Highlight ring on last added */}
      {paths && lastAddedGeo && (() => {
        const c = paths.countries.find(x => x.name === lastAddedGeo);
        if (!c) return null;
        return (
          <path
            d={c.d}
            fill="none"
            stroke="var(--fg-0)"
            strokeWidth="1.5"
            vectorEffect="non-scaling-stroke"
            style={{ pointerEvents: "none" }}
          >
            <animate attributeName="stroke-opacity" values="1;0.2;1" dur="1.2s" repeatCount="3" />
          </path>
        );
      })()}
    </svg>
  );
}

window.WorldMap = WorldMap;
window.toGeoName = toGeoName;
