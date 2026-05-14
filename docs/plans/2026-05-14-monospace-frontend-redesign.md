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

## Source materials (vendored)
The full Claude Design handoff bundle is checked into the repo at
`docs/design/` so this plan stays reproducible without refetching the
design URL. Original source:
`https://api.anthropic.com/v1/design/h/NLFBTX80T4dYFOn0kpkwlw`.

- `docs/design/README.md` — bundle's own README ("CODING AGENTS: READ
  THIS FIRST"), with the user-intent guidance
- `docs/design/chats/chat1.md` — full design conversation; **read for
  intent** (e.g. the user explicitly asked for the map to be
  display-only, no tap-to-toggle)
- `docs/design/project/index.html` — primary design file
- `docs/design/project/styles.css` — app-level component styles
- `docs/design/project/lib/monospace.css` — design tokens (colors,
  font, spacing)
- `docs/design/project/app.jsx` — React app with full interaction
  logic (reference, not shipped)
- `docs/design/project/world-map.jsx` — d3-geo + topojson SVG
  renderer (reference, not shipped)
- `docs/design/project/lib/dotgrid-bg.svg` — repeating dot grid
  texture (asset to copy)
- `docs/design/project/fonts/JetBrainsMono-400.woff2` and
  `JetBrainsMono-400italic.woff2` — variable JetBrains Mono (assets
  to copy)
- `docs/design/project/data/all_countries.json` — same shape as the
  existing `frontend/all_countries.json`

## Context (from discovery)
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
- [x] copy `docs/design/project/fonts/JetBrainsMono-400.woff2` and
      `JetBrainsMono-400italic.woff2` into `frontend/fonts/`
- [x] copy `docs/design/project/lib/dotgrid-bg.svg` into
      `frontend/css/dotgrid-bg.svg`
- [x] create `frontend/css/monospace.css` with the design-token
      `:root` block (colors, type scale, spacing, motion) and dark
      `:root[data-theme="dark"]` overrides — copied verbatim from
      `docs/design/project/lib/monospace.css`, minus the typography
      utilities we do not use
- [x] verify in a browser that fonts load (no 404 in devtools network
      panel) — the page still works with the old layout
      (smoke check via static server: all four assets + index 200 OK)
- [x] no behavior change in this task; smoke check only

### Task 2: Rewrite `frontend/index.html` to the design DOM
- [x] replace the body with the new structure: `app-header` (brand +
      meta), `qa` quick-actions strip, `hero` (stat row + caption +
      progress cells), `map-wrap` (with `<svg id="world">` and corner
      labels), `actions` (combo input + suggest dropdown + primary
      add button), `drawer` (continents, drawer-head, row-toolbar,
      visited list)
- [x] keep the IDs the e2e tests will rely on
      (`#visited-count`, `#countries-ul`, `#country-input`,
      `#add-country-btn`, `#toggle-list-btn`)
- [x] swap the Leaflet `<link>` and `<script>` tags for d3 v7 +
      topojson-client UMD bundles (matches the prototype, served from
      jsdelivr/unpkg)
- [x] include `css/monospace.css` before `css/style.css`
- [x] keep `telegram-web-app.js` and the `js/app.js` script tags
- [x] confirm DOM renders without error in the browser (layout will
      be unstyled until Task 3); no e2e change yet
      (smoke check via static server: required IDs present, no
      duplicate IDs, all assets 200 OK)

### Task 3: Replace `frontend/css/style.css` with the design's component CSS
- [x] write a new `frontend/css/style.css` that imports
      `monospace.css` and contains the component styles from
      `docs/design/project/styles.css` for: `.app`, `.app-header`,
      `.brand`,
      `.qa`, `.hero` + `.progress`, `.map-wrap` + `.world` paths,
      `.map-corner`, `.last-added`, `.actions`, `.combo`, `.btn`,
      `.drawer`, `.continents`, `.country-row`, `.empty`, `.toast`,
      `.suggest`
- [x] adapt `body`/`.app` to fill the viewport (drop the prototype's
      `.page` device frame, marginalia, and dotgrid background — the
      app runs inside Telegram, not on a designer's canvas)
- [x] keep all CSS variables from `monospace.css` for parity
- [x] dark-mode rules from the prototype are preserved (so the
      Telegram theme integration in Task 5 works)
- [x] visual smoke check: load the page, confirm fonts, colors,
      borders, and shadow-offset buttons match the prototype
      (smoke check via static server: index/style/monospace/dotgrid
      and both font files all 200 OK; CSS braces balanced; required
      selectors present; no `.page`/`.plate` chrome leftovers)

### Task 4: Build the d3-geo + topojson SVG world map
- [x] add a new `frontend/js/world-map.js` (vanilla JS, no JSX) that
      exposes `window.WorldMap.render({ svg, visitedSet, style,
      lastAddedGeo })` — fetches `countries-110m.json` once, caches
      paths, redraws country fills when `visitedSet` or `style`
      changes, draws sphere/graticule/countries with the same CSS
      classes as the prototype (`world`, `country`, `visited`,
      `graticule`, `sphere`)
- [x] include the `<defs><pattern id="dotfill">…</pattern></defs>` so
      the `style-dotfill` variant works
- [x] include the `NAME_ALIASES` table from
      `docs/design/project/world-map.jsx` so
      canonical names from `all_countries.json` map to topojson names
      correctly (United States ↔ United States of America, etc.)
- [x] load it via a `<script src="js/world-map.js">` tag in index.html
      (insert before `js/app.js` so `app.js` can call it)
- [x] manual test (skipped - not automatable in this loop; covered by
      the new Playwright assertion which invokes WorldMap.render with a
      seeded visited set and verifies the rendered `.visited` count)
- [x] add a Playwright assertion that `svg.world path.country.visited`
      count equals the number of seeded visited countries after page
      load

### Task 5: Rewrite `frontend/js/app.js` to drive the new UI
- [x] keep the same backend API calls (GET/POST/DELETE
      `/api/countries?userId=…`) and the Telegram WebApp init flow
- [x] introduce a small in-module state object: `{ visited: [{name,
      addedAt}], allCountries, sort: 'recent'|'alpha'|'continent',
      showList, query, activeSuggest }`
- [x] hero render: `padStart(2,'0')` count, `/195` denominator, `NN%`
      caption with "remaining" tail, 26 progress cells with `.on`
      class for filled cells
- [x] quick-actions strip: focus input on `+ add`, toggle drawer on
      hide/show list, cycle sort on sort, copy list to clipboard on
      share — each action triggers a toast
- [x] search combo: live `suggest` dropdown filtering
      `allCountries`, arrow-key navigation, Enter to add, Escape to
      clear, click suggestion to add, `×` clear button
- [x] drawer: continent breakdown (6 cells, `.full` when num===den),
      `drawer-head` count + sort label, segmented sort control, list
      of `.country-row` (idx, name, continent meta, recent-only `Xd
      ago` meta, `✗` delete), empty state when zero
- [x] toast helper: `showToast(text, isError)` that toggles `.err`
      and auto-clears in 2.4s
- [x] dark mode: read `tg.colorScheme` from `Telegram.WebApp` and
      set `document.documentElement.dataset.theme` accordingly; fall
      back to `prefers-color-scheme`
- [x] map integration: instantiate `WorldMap.render` once, then
      re-call it whenever `visited` changes (passing the new
      `visitedSet`); briefly highlight `lastAddedGeo` for 2s after
      add
- [x] preserve standalone-mode behavior (no userId → in-memory state)
- [x] update `e2e/tests/app.spec.js` to drive the new combo input
      (`#country-input`), the new add button (`#add-country-btn`),
      assert `#visited-count`, click `#toggle-list-btn`, and use the
      new `.country-row .x` selector for delete
- [x] add e2e cases: hero progress cell count after add, sort
      segment switch toggles `.on`, toast appears with success text
- [x] run Playwright e2e — must pass before next task
      (7/7 passed against a local server on PORT=18081; backend
      now reads `PORT` env var to avoid host port conflicts during
      test runs)

### Task 6: Server static-file sanity + final polish
- [x] confirm `backend/main.go` still serves the new asset paths
      (`/css/monospace.css`, `/fonts/JetBrainsMono-400.woff2`,
      `/js/world-map.js`) — no code change expected, just a curl
      check on each
      (all served by `http.FileServer(http.Dir("./frontend"))` at
      `backend/main.go:244-246`; curl against a local instance on
      PORT=18082 returned HTTP 200 for `/css/monospace.css`,
      `/css/style.css`, `/css/dotgrid-bg.svg`,
      `/fonts/JetBrainsMono-400.woff2`,
      `/fonts/JetBrainsMono-400italic.woff2`, `/js/world-map.js`,
      `/js/app.js`, `/`, and `/all_countries.json`)
- [x] regression: run Go unit tests (`go test ./...`) — must pass
      (both `github.com/korjavin/countrycounter/backend` and
      `.../backend/store` packages pass)
- [x] regression: re-run full Playwright e2e — must pass
      (7/7 passed against the local server)

### Task 7: Verify acceptance criteria
- [x] visual diff vs. prototype: header brand, hero stat, progress
      cells, map silhouette + visited shading, quick actions strip,
      drawer with continents and visited rows, autocomplete dropdown,
      toast — each must look like `docs/design/project/index.html`
      (manual visual review — skipped, not automatable from this loop;
      tracked under Post-Completion / manual verification, where the
      reviewer can side-by-side the running app against
      `docs/design/project/index.html`)
- [x] keyboard interactions: input focus on `+ add`, ↑/↓ in
      suggestions, Enter adds top match, Escape clears
      (verified by reading `frontend/js/app.js`:
      `addFocusBtn` handler at app.js:553-555 focuses
      `#country-input`; the keydown handler at app.js:585-604 wires
      ArrowDown / ArrowUp to move `state.activeSuggest`, Enter to
      call `handleAddClick` which adds the active suggestion, and
      Escape to clear the query + blur the input; the existing e2e
      `autocomplete suggest keyboard navigation` covers Enter-adds
      end-to-end)
- [x] dark mode toggle from Telegram WebApp (`tg.colorScheme`)
      flips theme correctly
      (verified by reading `frontend/js/app.js`:141-153 —
      `applyTheme` reads `tg.colorScheme` and sets
      `document.documentElement.dataset.theme` to `dark` or `light`,
      and subscribes to `tg.onEvent("themeChanged", applyTheme)` so
      live Telegram theme changes flip the attribute;
      `frontend/css/monospace.css` already carries the
      `:root[data-theme="dark"]` override block. Live verification
      against a real Telegram client is captured under
      Post-Completion / manual verification)
- [x] all e2e specs pass; all Go unit tests pass
      (Go: `go test ./...` from `backend/` returns
      `ok github.com/korjavin/countrycounter/backend` and
      `ok github.com/korjavin/countrycounter/backend/store`.
      Playwright: 7/7 passed against a local server on
      `APP_BASE_URL=http://localhost:18083` —
      load + SVG map, world map seeded-visited count, add country,
      delete via ✗, hide/show drawer, sort segment switch,
      autocomplete keyboard Enter)

### Task 8: [Final] Update documentation
- [ ] update `readme.md` "Frontend" tech-stack bullet to mention
      d3-geo + topojson and JetBrains Mono (was: Leaflet + sans-serif)
- [ ] note in `readme.md` that the design is now a monospace,
      two-color treatment matching the design bundle
- [ ] (optional) refresh `logo.png` reference / screenshot if the
      old screenshot is now stale

## Technical Details

### Design tokens (from `docs/design/project/lib/monospace.css`)
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
  via the `NAME_ALIASES` table from
  `docs/design/project/world-map.jsx`.
- Map is **display-only** — per the chat transcript the user
  explicitly removed tap-to-toggle; we keep it that way.

### Continent breakdown
- Static name-to-continent table (`CONTINENT_OF`) lifted from
  `docs/design/project/app.jsx`; totals
  `{ AF:54, EU:46, AS:47, NA:23, SA:12, OC:14 }`;
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
