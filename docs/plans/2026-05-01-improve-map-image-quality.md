# Improve Map Image Quality

## Overview

Upgrade the map rendering in generateMapImage to produce a more polished, visually appealing world map. Three areas: replace the flat equirectangular projection with Mercator, add country border outlines, and improve the color palette with an ocean background.

## Context

- Files involved: `backend/main.go` (generateMapImage at line 47, drawPolygon at line 143)
- Related patterns: uses `github.com/fogleman/gg` for 2D drawing, `github.com/paulmach/go.geojson` for GeoJSON parsing
- Dependencies: no new dependencies needed; gg already supports FillPreserve, Stroke, and arbitrary coordinate transforms

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- **CRITICAL: every task MUST include new/updated tests**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Improve base styling and color palette

**Files:**
- Modify: `backend/main.go`

- [ ] Change the background clear color from light gray (0.9, 0.9, 0.9) to an ocean blue (e.g. RGB 0.65, 0.81, 0.89)
- [ ] Update visited country color from gold RGBA(212,172,13) to a richer tone (e.g. RGBA 46,204,113 green or RGBA 231,76,60 warm red — pick a palette that contrasts well with ocean blue)
- [ ] Update unvisited country color from RGBA(200,200,200) to a lighter neutral that sits well on the blue ocean (e.g. RGBA 230,230,220 warm white)
- [ ] Write a test that calls generateMapImage with an empty list and a non-empty list, verifying it returns valid PNG bytes and expected dimensions
- [ ] Run project test suite — must pass before task 2

### Task 2: Add country border outlines

**Files:**
- Modify: `backend/main.go`

- [ ] In the polygon drawing loop, switch from dc.Fill() to dc.FillPreserve() so the path is retained after fill
- [ ] After FillPreserve, set a border color (e.g. RGBA 100,100,100 with alpha 180) and line width (e.g. 0.5px)
- [ ] Call dc.Stroke() to draw the border outline
- [ ] Ensure this works for both Polygon and MultiPolygon geometry types
- [ ] Write a test that generates a map image and checks the output is valid PNG with non-zero size
- [ ] Run project test suite — must pass before task 3

### Task 3: Implement Mercator projection

**Files:**
- Modify: `backend/main.go`

- [ ] Replace the linear Y-axis mapping in drawPolygon with the Mercator formula: y = ln(tan(pi/4 + lat_rad/2)), where lat_rad = latitude * pi / 180
- [ ] Clamp latitude input to [-85, 85] degrees to avoid infinity near the poles
- [ ] Update the bounding box calculation to also use the Mercator Y transform so scale factors are consistent
- [ ] Fix the bounding box loop to also iterate MultiPolygon coordinates (currently only handles Polygon, which misses countries like Russia and USA)
- [ ] Write a test that verifies the Mercator transform produces expected Y values for known latitudes (e.g. equator=0, 45N, 60N)
- [ ] Run project test suite — must pass before task 4

### Task 4: Verify acceptance criteria

- [ ] Run full test suite: `cd backend && go test ./...`
- [ ] Run linter: `cd backend && go vet ./...`
- [ ] Visually verify the generated map by running the bot or adding a test that writes the PNG to a temp file for inspection

### Task 5: Update documentation

- [ ] Update README.md if user-facing changes
- [ ] Update CLAUDE.md if internal patterns changed
- [ ] Move this plan to `docs/plans/completed/`
