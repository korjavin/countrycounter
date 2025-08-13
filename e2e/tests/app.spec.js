const { test, expect } = require('@playwright/test');

test.describe('Country Counter App', () => {
  const appUrl = 'http://localhost:8080';

  test('should load the page and display the map', async ({ page }) => {
    await page.goto(appUrl);
    await expect(page.locator('#map')).toBeVisible();
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
});
