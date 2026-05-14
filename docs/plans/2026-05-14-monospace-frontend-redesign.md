# Monospace Frontend Redesign

## Overview
Reimplement `frontend/index.html` (and its companion CSS/JS) to match the
hi-fi prototype delivered in the Claude Design handoff bundle
(`countrycountrer/project/index.html` from
`https://api.anthropic.com/v1/design/h/NLFBTX80T4dYFOn0kpkwlw`).

The prototype is a **monospace, two-color (paper / oilcloth) Telegram
WebApp** built with React + d3-geo + topojson. It introduces:

- a header with `[countrycounter]` brand and version meta,
- a "quick actions" strip (`+ add`, hide list, sort, share),
- a big hero stat (`NN / 195`) with character-grid progress cells,
- an SVG world map (Equal-Earth projection, monochrome),
- a search row with autocomplete suggestions,
- a drawer with continent breakdown, segmented sort controls, visited
  list with row index + continent meta + delete, and a clear button,
- a toast for feedback.

The production app currently uses Leaflet tiles + a generic sans-serif
layout. We will rebuild the frontend to match the prototype's visual
output **pixel-faithfully**, while keeping our backend API
(`GET/POST/DELETE /api/countries?userId=…`) and Telegram WebApp
integration unchanged.

The README's instructions are unambiguous: match the visual output, but
do not have to copy the prototype's React/Babel internal structure. We
will translate the design into vanilla JS (matching the rest of the
existing frontend stack) and drop Leaflet in favor of an inline d3-geo +
topojson SVG map, since Leaflet tiles directly conflict with the
prototype's stark monochrome aesthetic.

The scaffolding parts of the prototype (the iOS/device frame, left and
right "plate" marginalia, tweaks panel) are designer-only chrome and
will not ship — the app already runs inside Telegram on a real device.

## Context (from discovery)
- Design bundle: `/tmp/design-fetch/countrycountrer/`
  - `project/index.html` — primary file
  - `project/styles.css` — app-level component styles
  - `project/lib/monospace.css` — design tokens (colors, font, spacing)
  - `project/app.jsx` — React app with full interaction logic
  - `project/world-map.jsx` — d3-geo + topojson SVG renderer
  - `project/lib/dotgrid-bg.svg` — repeating dot grid texture
  - `project/fonts/JetBrainsMono-400*.woff2` — variable JetBrains Mono
- Files affected:
  - `frontend/index.html` — full rewrite
  - `frontend/css/style.css` — full rewrite, replaced by the
    monospace token system + component styles
  - `frontend/js/app.js` — full rewrite (vanilla JS, no React)
  - new: `frontend/css/monospace.css` (design tokens)
  - new: `frontend/css/dotgrid-bg.svg`
  - new: `frontend/fonts/JetBrainsMono-400.woff2`,
    `frontend/fonts/JetBrainsMono-400italic.woff2`
  - `e2e/tests/app.spec.js` — selectors change
- Backend: `backend/main.go:243-244` serves `./frontend` as a static
  file server. New asset paths (`/css/monospace.css`, `/fonts/...`)
  fall under that root with no server change required.
- The `frontend/all_countries.json` file already exists with the same
  shape the prototype expects.

## Development Approach
- **Testing approach**: Regular (code first, then tests). Frontend is
  primarily visual; the existing test surface is Playwright e2e. We
  update e2e in lockstep with each DOM change.
- Complete each task fully before moving to the next.
- Make small, focused changes.
- **CRITICAL: every task that touches code MUST end with passing tests
  before starting the next.** Tests are not optional.
  - For pure asset/styling tasks where no behavior changes, run a
    smoke check: load the page in the dev server and confirm no console
    errors and the layout renders.
  - For behavior tasks (search, add, delete, sort, toast), update or
    add e2e specs in the same task.
- **CRITICAL: update this plan file when scope changes.**
- Maintain backward compatibility on the backend API
  (`GET/POST/DELETE /api/countries`).

## Testing Strategy
- **Unit tests**: backend Go tests are unaffected — re-run them once
  at the end as a regression sanity check.
- **E2E tests** (Playwright):
  - Update `e2e/tests/app.spec.js` selectors as the DOM changes.
  - Add coverage for: hero count update, progress cells filling, sort
    segment switch, autocomplete keyboard navigation, toast text,
    delete via the `✗` button.
  - Each behavior task updates e2e in the same task.

## Progress Tracking
- Mark completed items with `[x]` immediately.
- Add newly discovered tasks with `➕`.
- Document blockers with `⚠️`.

## What Goes Where
- **Implementation Steps** (`[ ]`): code, asset, and test changes.
- **Post-Completion** (no checkboxes): manual verification on a real
  Telegram client, screenshot refresh in README, deployment.

## Implementation Steps

### Task 1: Bring design assets into the frontend
- [ ] copy `JetBrainsMono-400.woff2` and `JetBrainsMono-400italic.woff2`
      from the design bundle into `frontend/fonts/`
- [ ] copy `dotgrid-bg.svg` from the design bundle into
      `frontend/css/dotgrid-bg.svg`
- [ ] create `frontend/css/monospace.css` with the design-token
      `:root` block (colors, type scale, spacing, motion) and dark
      `:root[data-theme="dark"]` overrides — copied verbatim from the
      bundle's `lib/monospace.css`, minus the typography utilities we
      do not use
- [ ] verify in a browser that fonts load (no 404 in devtools network
      panel) — the page still works with the old layout
- [ ] no behavior change in this task; smoke check only

### Task 2: Rewrite `frontend/index.html` to the design DOM
- [ ] replace the body with the new structure: `app-header` (brand +
      meta), `qa` quick-actions strip, `hero` (stat row + caption +
      progress cells), `map-wrap` (with `<svg id="world">` and corner
      labels), `actions` (combo input + suggest dropdown + primary
      add button), `drawer` (continents, drawer-head, row-toolbar,
      visited list)
- [ ] keep the IDs the e2e tests will rely on
      (`#visited-count`, `#countries-ul`, `#country-input`,
      `#add-country-btn`, `#toggle-list-btn`)
- [ ] swap the Leaflet `<link>` and `<script>` tags for d3 v7 +
      topojson-client UMD bundles (matches the prototype, served from
      jsdelivr/unpkg)
- [ ] include `css/monospace.css` before `css/style.css`
- [ ] keep `telegram-web-app.js` and the `js/app.js` script tags
- [ ] confirm DOM renders without error in the browser (layout will
      be unstyled until Task 3); no e2e change yet

### Task 3: Replace `frontend/css/style.css` with the design's component CSS
- [ ] write a new `frontend/css/style.css` that imports
      `monospace.css` and contains the component styles from the
      bundle's `styles.css` for: `.app`, `.app-header`, `.brand`,
      `.qa`, `.hero` + `.progress`, `.map-wrap` + `.world` paths,
      `.map-corner`, `.last-added`, `.actions`, `.combo`, `.btn`,
      `.drawer`, `.continents`, `.country-row`, `.empty`, `.toast`,
      `.suggest`
- [ ] adapt `body`/`.app` to fill the viewport (drop the prototype's
      `.page` device frame, marginalia, and dotgrid background — the
      app runs inside Telegram, not on a designer's canvas)
- [ ] keep all CSS variables from `monospace.css` for parity
- [ ] dark-mode rules from the prototype are preserved (so the
      Telegram theme integration in Task 5 works)
- [ ] visual smoke check: load the page, confirm fonts, colors,
      borders, and shadow-offset buttons match the prototype

### Task 4: Build the d3-geo + topojson SVG world map
- [ ] add a new `frontend/js/world-map.js` (vanilla JS, no JSX) that
      exposes `window.WorldMap.render({ svg, visitedSet, style,
      lastAddedGeo })` — fetches `countries-110m.json` once, caches
      paths, redraws country fills when `visitedSet` or `style`
      changes, draws sphere/graticule/countries with the same CSS
      classes as the prototype (`world`, `country`, `visited`,
      `graticule`, `sphere`)
- [ ] include the `<defs><pattern id="dotfill">…</pattern></defs>` so
      the `style-dotfill` variant works
- [ ] include the `NAME_ALIASES` table from `world-map.jsx` so
      canonical names from `all_countries.json` map to topojson names
      correctly (United States ↔ United States of America, etc.)
- [ ] load it via a `<script src="js/world-map.js">` tag in index.html
      (insert before `js/app.js` so `app.js` can call it)
- [ ] manual check: page loads, world silhouette appears, marking
      `["United States","France"]` visited shades the right shapes
- [ ] add a Playwright assertion that `svg.world path.country.visited`
      count equals the number of seeded visited countries after page
      load

### Task 5: Rewrite `frontend/js/app.js` to drive the new UI
- [ ] keep the same backend API calls (GET/POST/DELETE
      `/api/countries?userId=…`) and the Telegram WebApp init flow
- [ ] introduce a small in-module state object: `{ visited: [{name,
      addedAt}], allCountries, sort: 'recent'|'alpha'|'continent',
      showList, query, activeSuggest }`
- [ ] hero render: `padStart(2,'0')` count, `/195` denominator, `NN%`
      caption with "remaining" tail, 26 progress cells with `.on`
      class for filled cells
- [ ] quick-actions strip: focus input on `+ add`, toggle drawer on
      hide/show list, cycle sort on sort, copy list to clipboard on
      share — each action triggers a toast
- [ ] search combo: live `suggest` dropdown filtering
      `allCountries`, arrow-key navigation, Enter to add, Escape to
      clear, click suggestion to add, `×` clear button
- [ ] drawer: continent breakdown (6 cells, `.full` when num===den),
      `drawer-head` count + sort label, segmented sort control, list
      of `.country-row` (idx, name, continent meta, recent-only `Xd
      ago` meta, `✗` delete), empty state when zero
- [ ] toast helper: `showToast(text, isError)` that toggles `.err`
      and auto-clears in 2.4s
- [ ] dark mode: read `tg.colorScheme` from `Telegram.WebApp` and
      set `document.documentElement.dataset.theme` accordingly; fall
      back to `prefers-color-scheme`
- [ ] map integration: instantiate `WorldMap.render` once, then
      re-call it whenever `visited` changes (passing the new
      `visitedSet`); briefly highlight `lastAddedGeo` for 2s after
      add
- [ ] preserve standalone-mode behavior (no userId → in-memory state)
- [ ] update `e2e/tests/app.spec.js` to drive the new combo input
      (`#country-input`), the new add button (`#add-country-btn`),
      assert `#visited-count`, click `#toggle-list-btn`, and use the
      new `.country-row .x` selector for delete
- [ ] add e2e cases: hero progress cell count after add, sort
      segment switch toggles `.on`, toast appears with success text
- [ ] run Playwright e2e — must pass before next task

### Task 6: Server static-file sanity + final polish
- [ ] confirm `backend/main.go` still serves the new asset paths
      (`/css/monospace.css`, `/fonts/JetBrainsMono-400.woff2`,
      `/js/world-map.js`) — no code change expected, just a curl
      check on each
- [ ] regression: run Go unit tests (`go test ./...`) — must pass
- [ ] regression: re-run full Playwright e2e — must pass

### Task 7: Verify acceptance criteria
- [ ] visual diff vs. prototype: header brand, hero stat, progress
      cells, map silhouette + visited shading, quick actions strip,
      drawer with continents and visited rows, autocomplete dropdown,
      toast — each must look like the bundle's `index.html`
- [ ] keyboard interactions: input focus on `+ add`, ↑/↓ in
      suggestions, Enter adds top match, Escape clears
- [ ] dark mode toggle from Telegram WebApp (`tg.colorScheme`)
      flips theme correctly
- [ ] all e2e specs pass; all Go unit tests pass

### Task 8: [Final] Update documentation
- [ ] update `readme.md` "Frontend" tech-stack bullet to mention
      d3-geo + topojson and JetBrains Mono (was: Leaflet + sans-serif)
- [ ] note in `readme.md` that the design is now a monospace,
      two-color treatment matching the design bundle
- [ ] (optional) refresh `logo.png` reference / screenshot if the
      old screenshot is now stale

## Technical Details

### Design tokens (from `monospace.css`)
- Light: `--bg-0:#f4f1ea`, `--bg-1:#ece8de`, `--bg-elev:#fbf9f3`,
  `--fg-0:#16140f`, `--fg-1:#3a362d`, `--fg-2:#6b6657`,
  `--line-1:#16140f`, `--line-2:#c8c1ad`.
- Dark: `--bg-0:#0e0e0c`, `--bg-1:#161614`, `--bg-elev:#1a1a17`,
  `--fg-0:#e8e4d6`, `--line-1:#e8e4d6`, `--line-2:#2e2d29`.
- Font: `JetBrains Mono` variable, fallback to `SF Mono / Menlo /
  Consolas / monospace`.
- Hard `2px 2px 0 0 var(--line-1)` shadow on buttons; click translates
  to `2px,2px` and clears the shadow. Square corners (`border-radius:
  0`) everywhere.

### Map rendering
- Uses `d3.geoEqualEarth().fitSize([w-12, h-12], featureCollection)`
  on an `800×460` viewport, drawn into a single SVG with
  `preserveAspectRatio="xMidYMid meet"`.
- World atlas: `https://cdn.jsdelivr.net/npm/world-atlas@2.0.2/countries-110m.json`.
- Visited countries are matched against topojson `properties.name`
  via the `NAME_ALIASES` table from `world-map.jsx`.
- Map is **display-only** — per the chat transcript the user
  explicitly removed tap-to-toggle; we keep it that way.

### Continent breakdown
- Static name-to-continent table (`CONTINENT_OF`) lifted from
  `app.jsx`; totals `{ AF:54, EU:46, AS:47, NA:23, SA:12, OC:14 }`;
  labels `{ NA:"N.AMER", SA:"S.AMER", EU:"EUROPE", AF:"AFRICA",
  AS:"ASIA", OC:"OCEAN." }`.

### Backend API contract (unchanged)
- `GET /api/countries?userId=<id>` → `["United States", "France", …]`
- `POST /api/countries` body `{userId, country}` → 201
- `DELETE /api/countries` body `{userId, country}` → 200

## Post-Completion

**Manual verification**:
- Load the app inside the actual Telegram bot WebApp on iOS and
  Android; confirm fonts load over the network and dark mode
  follows the Telegram theme.
- Confirm the safe-area / status-bar interaction looks right on a
  notched device.

**External system updates**:
- None required. The Docker image build and the Go server are
  unaffected.
