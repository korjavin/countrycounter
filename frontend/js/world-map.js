// world-map.js — SVG world map using d3-geo + topojson, vanilla JS port of
// docs/design/project/world-map.jsx. Exposes window.WorldMap.render(opts).

(function () {
  const SVG_NS = "http://www.w3.org/2000/svg";
  const WORLD_ATLAS_URL =
    "https://cdn.jsdelivr.net/npm/world-atlas@2.0.2/countries-110m.json";

  // Values match country names in world-atlas@2.0.2 countries-110m.json
  // (Natural Earth display names). Keys are names that may show up in user
  // data — from the autocomplete in all_countries.json or legacy entries.
  const NAME_ALIASES = {
    "United States": "United States of America",
    "United States of America": "United States of America",
    "Czech Republic": "Czechia",
    "Cote d'Ivoire": "Côte d'Ivoire",
    "Ivory Coast": "Côte d'Ivoire",
    "Congo, Democratic Republic of the": "Dem. Rep. Congo",
    "DR Congo": "Dem. Rep. Congo",
    "Congo, Republic of the": "Congo",
    "Eswatini": "eSwatini",
    "North Macedonia": "Macedonia",
    "Bosnia and Herzegovina": "Bosnia and Herz.",
    "Central African Republic": "Central African Rep.",
    "Dominican Republic": "Dominican Rep.",
    "Equatorial Guinea": "Eq. Guinea",
    "Solomon Islands": "Solomon Is.",
    "South Sudan": "S. Sudan",
    "Palestine State": "Palestine",
  };

  function toGeoName(name) {
    return NAME_ALIASES[name] || name;
  }

  let pathsCache = null;
  let loadPromise = null;

  function librariesReady() {
    return (
      typeof window.topojson !== "undefined" &&
      typeof window.d3 !== "undefined" &&
      typeof window.d3.geoEqualEarth === "function"
    );
  }

  async function waitForLibraries() {
    let tries = 0;
    while (!librariesReady() && tries < 100) {
      await new Promise((r) => setTimeout(r, 50));
      tries++;
    }
    if (!librariesReady()) {
      throw new Error("d3-geo / topojson-client failed to load");
    }
  }

  async function loadPaths() {
    if (pathsCache) return pathsCache;
    if (loadPromise) return loadPromise;
    loadPromise = (async () => {
      try {
        await waitForLibraries();
        const world = await fetch(WORLD_ATLAS_URL).then((r) => r.json());
        const fc = window.topojson.feature(world, world.objects.countries);
        const w = 800;
        const h = 460;
        const proj = window.d3.geoEqualEarth().fitSize([w - 12, h - 12], fc);
        const path = window.d3.geoPath(proj);
        const countries = fc.features
          .map((f) => ({ name: f.properties.name, d: path(f) }))
          .filter((p) => p.d);
        const graticule = path(window.d3.geoGraticule10());
        const sphere = path({ type: "Sphere" });
        pathsCache = { countries, graticule, sphere, w, h };
        return pathsCache;
      } catch (e) {
        // Don't poison the cache — let future calls retry.
        loadPromise = null;
        throw e;
      }
    })();
    return loadPromise;
  }

  function makeEl(tag, attrs) {
    const el = document.createElementNS(SVG_NS, tag);
    if (attrs) {
      for (const k in attrs) {
        if (attrs[k] != null) el.setAttribute(k, attrs[k]);
      }
    }
    return el;
  }

  function ensureDefs(svg) {
    let defs = svg.querySelector(":scope > defs");
    if (defs) return defs;
    defs = makeEl("defs");
    const pattern = makeEl("pattern", {
      id: "dotfill",
      patternUnits: "userSpaceOnUse",
      width: "4",
      height: "4",
    });
    pattern.appendChild(
      makeEl("rect", { width: "4", height: "4", fill: "var(--bg-0)" }),
    );
    pattern.appendChild(
      makeEl("circle", { cx: "2", cy: "2", r: "0.9", fill: "var(--fg-0)" }),
    );
    defs.appendChild(pattern);
    svg.appendChild(defs);
    return defs;
  }

  function clearChildrenExceptDefs(svg) {
    const children = Array.from(svg.childNodes);
    for (const c of children) {
      if (c.nodeType === 1 && c.tagName.toLowerCase() === "defs") continue;
      svg.removeChild(c);
    }
  }

  async function render(opts) {
    const svg = opts && opts.svg;
    if (!svg) throw new Error("WorldMap.render: opts.svg is required");
    const visitedSet = (opts && opts.visitedSet) || new Set();
    const style = (opts && opts.style) || "solid";
    const lastAddedGeo = (opts && opts.lastAddedGeo) || null;

    svg.setAttribute("class", `world style-${style}`);
    svg.setAttribute("preserveAspectRatio", "xMidYMid meet");

    ensureDefs(svg);

    let paths;
    try {
      paths = await loadPaths();
    } catch (e) {
      console.error("Map load failed", e);
      return;
    }

    svg.setAttribute("viewBox", `0 0 ${paths.w} ${paths.h}`);

    const visitedGeoNames = new Set();
    visitedSet.forEach((n) => visitedGeoNames.add(toGeoName(n)));

    clearChildrenExceptDefs(svg);

    svg.appendChild(makeEl("path", { class: "sphere", d: paths.sphere }));
    svg.appendChild(
      makeEl("path", { class: "graticule", d: paths.graticule }),
    );

    for (const c of paths.countries) {
      const visited = visitedGeoNames.has(c.name);
      const isLast = lastAddedGeo === c.name;
      let cls = "country";
      if (visited) cls += " visited";
      if (isLast) cls += " just-added";
      const p = makeEl("path", { class: cls, d: c.d });
      const title = document.createElementNS(SVG_NS, "title");
      title.textContent = visited ? `${c.name} — visited` : c.name;
      p.appendChild(title);
      svg.appendChild(p);
    }

    if (lastAddedGeo) {
      const c = paths.countries.find((x) => x.name === lastAddedGeo);
      if (c) {
        const ring = makeEl("path", {
          d: c.d,
          fill: "none",
          stroke: "var(--fg-0)",
          "stroke-width": "1.5",
          "vector-effect": "non-scaling-stroke",
          "pointer-events": "none",
        });
        const anim = makeEl("animate", {
          attributeName: "stroke-opacity",
          values: "1;0.2;1",
          dur: "1.2s",
          repeatCount: "3",
        });
        ring.appendChild(anim);
        svg.appendChild(ring);
      }
    }
  }

  window.WorldMap = { render, toGeoName, NAME_ALIASES };
  window.toGeoName = toGeoName;
})();
