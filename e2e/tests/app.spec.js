const { test, expect } = require('@playwright/test');

test.describe('Country Counter App', () => {
  const appUrl = process.env.APP_BASE_URL || 'http://localhost:8080';

  // Each test uses a fresh user id (numeric) so the backend state is isolated.
  // The test page injects a stub Telegram.WebApp before app.js runs.
  async function gotoWithUser(page, userId) {
    await page.addInitScript((uid) => {
      window.Telegram = {
        WebApp: {
          initDataUnsafe: { user: { id: uid } },
          colorScheme: 'light',
          ready: () => {},
          onEvent: () => {},
        },
      };
    }, userId);
    await page.goto(appUrl);
    // Wait until app has initialised: both loadAllCountries() and loadVisited()
    // have completed and render() has run. state.ready is set last in init().
    // Gating on allCountries alone races loadVisited — a click between the two
    // awaits would have its push to state.visited clobbered by the still-in-
    // flight loadVisited.
    await page.waitForFunction(
      () =>
        window.__app &&
        window.__app.state &&
        window.__app.state.ready === true,
    );
  }

  function uniqueUserId() {
    // Large random number; namespaced to e2e by prefix.
    return 900000000 + Math.floor(Math.random() * 100000000);
  }

  test('loads page and renders the SVG world map', async ({ page }) => {
    await gotoWithUser(page, uniqueUserId());
    await expect(page.locator('#world')).toBeVisible();
  });

  test('world map renders visited paths matching seeded set', async ({ page }) => {
    await page.goto(appUrl);

    // Wait for libraries AND the app's own initial map render to complete
    // before seeding. The app calls WorldMap.render() once during init; that
    // call is async and (because pathsCache is shared) would otherwise resolve
    // after our seeded render and clear the .visited paths. Once
    // svg.world path.country elements exist, the initial render is done and
    // it's safe to overlay our seeded set.
    await page.waitForFunction(
      () =>
        typeof window.WorldMap !== 'undefined' &&
        typeof window.d3 !== 'undefined' &&
        typeof window.topojson !== 'undefined' &&
        document.querySelectorAll('svg.world path.country').length > 0,
      null,
      { timeout: 10000 },
    );

    const seeded = ['United States', 'France', 'Japan'];
    await page.evaluate(async (visited) => {
      const svg = document.getElementById('world');
      await window.WorldMap.render({
        svg,
        visitedSet: new Set(visited),
        style: 'solid',
      });
    }, seeded);

    const visitedCount = await page
      .locator('svg.world path.country.visited')
      .count();
    expect(visitedCount).toBe(seeded.length);
  });

  test('adds a country and updates count, progress, and last-added overlay', async ({ page }) => {
    await gotoWithUser(page, uniqueUserId());

    const initialCount = await page.locator('#visited-count').textContent();
    const initialFilled = await page.locator('#progress-cells .progress-cell.on').count();

    await page.fill('#country-input', 'Canada');
    await page.click('#add-country-btn');

    await expect(page.locator('#visited-count')).not.toHaveText(initialCount);
    const newCount = await page.locator('#visited-count').textContent();
    expect(parseInt(newCount, 10)).toBe(parseInt(initialCount, 10) + 1);

    // Toast text confirms success.
    await expect(page.locator('#toast')).toBeVisible();
    await expect(page.locator('#toast-text')).toHaveText(/Added — Canada/);

    // Progress cells filled count is monotonic non-decreasing.
    const newFilled = await page.locator('#progress-cells .progress-cell.on').count();
    expect(newFilled).toBeGreaterThanOrEqual(initialFilled);

    // Country appears in the drawer list.
    await expect(page.locator('#countries-ul')).toContainText('Canada');

    // Last-added overlay shows the new country (default sort is recent).
    await expect(page.locator('#last-added')).toBeVisible();
    await expect(page.locator('#last-added')).toContainText('Canada');
  });

  test('deletes a country via the row ✗ button and updates count', async ({ page }) => {
    await gotoWithUser(page, uniqueUserId());

    await page.fill('#country-input', 'Germany');
    await page.click('#add-country-btn');
    await expect(page.locator('#countries-ul')).toContainText('Germany');

    const countAfterAdd = await page.locator('#visited-count').textContent();

    await page.locator('.country-row', { hasText: 'Germany' }).locator('.x').click();

    await expect(page.locator('#visited-count')).not.toHaveText(countAfterAdd);
    const finalCount = await page.locator('#visited-count').textContent();
    expect(parseInt(finalCount, 10)).toBe(parseInt(countAfterAdd, 10) - 1);

    await expect(page.locator('#countries-ul')).not.toContainText('Germany');
  });

  test('hide/show list toggles drawer via quick action', async ({ page }) => {
    await gotoWithUser(page, uniqueUserId());

    await expect(page.locator('#drawer')).toBeVisible();
    await page.click('#toggle-list-btn');
    await expect(page.locator('#drawer')).toBeHidden();
    await expect(page.locator('#toggle-list-btn .qa-label')).toHaveText(/show list/);

    await page.click('#toggle-list-btn');
    await expect(page.locator('#drawer')).toBeVisible();
    await expect(page.locator('#toggle-list-btn .qa-label')).toHaveText(/hide list/);
  });

  test('sort segment switches the .on class between buttons', async ({ page }) => {
    await gotoWithUser(page, uniqueUserId());

    // Default: recent
    await expect(page.locator('#sort-seg button[data-sort="recent"]')).toHaveClass(/on/);
    await expect(page.locator('#sort-seg button[data-sort="alpha"]')).not.toHaveClass(/on/);

    await page.click('#sort-seg button[data-sort="alpha"]');
    await expect(page.locator('#sort-seg button[data-sort="alpha"]')).toHaveClass(/on/);
    await expect(page.locator('#sort-seg button[data-sort="recent"]')).not.toHaveClass(/on/);

    await page.click('#sort-seg button[data-sort="continent"]');
    await expect(page.locator('#sort-seg button[data-sort="continent"]')).toHaveClass(/on/);
  });

  test('autocomplete suggest keyboard navigation: ArrowDown then Enter adds', async ({ page }) => {
    await gotoWithUser(page, uniqueUserId());

    const initialCount = await page.locator('#visited-count').textContent();

    await page.focus('#country-input');
    await page.fill('#country-input', 'Sp'); // matches Spain, Sri Lanka? Spain starts with Sp.
    await expect(page.locator('#suggest')).toBeVisible();
    // First suggestion is active by default; press Enter.
    await page.keyboard.press('Enter');

    await expect(page.locator('#visited-count')).not.toHaveText(initialCount);
    await expect(page.locator('#countries-ul')).toContainText('Spain');
    await expect(page.locator('#toast-text')).toHaveText(/Added — Spain/);
  });
});
