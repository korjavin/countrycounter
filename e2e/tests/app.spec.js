const { test, expect } = require('@playwright/test');

test.describe('Country Counter App', () => {
  const appUrl = 'http://localhost:8080';

  test('should load the page and display the map', async ({ page }) => {
    await page.goto(appUrl);
    await expect(page.locator('#map')).toBeVisible();
  });

  test('world map renders visited paths matching seeded set', async ({ page }) => {
    await page.goto(appUrl);

    await page.waitForFunction(
      () =>
        typeof window.WorldMap !== 'undefined' &&
        typeof window.d3 !== 'undefined' &&
        typeof window.topojson !== 'undefined',
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

    await expect
      .poll(
        async () => await page.locator('svg.world path.country').count(),
        { timeout: 10000 },
      )
      .toBeGreaterThan(0);

    const visitedCount = await page
      .locator('svg.world path.country.visited')
      .count();
    expect(visitedCount).toBe(seeded.length);
  });

  test('should add a country and update the count', async ({ page }) => {
    await page.goto(appUrl);

    // Get the initial count
    const initialCount = await page.locator('#visited-count').textContent();

    // Select a country from the dropdown
    await page.selectOption('#country-select', 'Canada');

    // Click the "Add Country" button
    await page.click('#add-country-btn');

    // Wait for the count to be updated
    await page.waitForFunction((initialCount) => {
      const newCount = document.querySelector('#visited-count').textContent;
      return newCount > initialCount;
    }, initialCount);

    // Verify the count has been updated
    const newCount = await page.locator('#visited-count').textContent();
    expect(parseInt(newCount)).toBe(parseInt(initialCount) + 1);

    // Verify the country is in the list
    await page.click('#toggle-list-btn');
    await expect(page.locator('#countries-ul')).toContainText('Canada');
  });

  test('should delete a country and update the count', async ({ page }) => {
    await page.goto(appUrl);

    // Add a country first
    await page.fill('#country-input', 'Germany');
    await page.click('#add-country-btn');

    // Wait for the country to be added
    await page.waitForSelector('#countries-ul li:has-text("Germany")');

    // Get the initial count
    const initialCount = await page.locator('#visited-count').textContent();

    // Click the delete button
    await page.click('#countries-ul li:has-text("Germany") .delete-btn');

    // Wait for the count to be updated
    await page.waitForFunction((initialCount) => {
      const newCount = document.querySelector('#visited-count').textContent;
      return newCount < initialCount;
    }, initialCount);

    // Verify the count has been updated
    const newCount = await page.locator('#visited-count').textContent();
    expect(parseInt(newCount)).toBe(parseInt(initialCount) - 1);

    // Verify the country is not in the list
    await expect(page.locator('#countries-ul')).not.toContainText('Germany');
  });
});
