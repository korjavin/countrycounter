const { test, expect } = require('@playwright/test');

test.describe('Country Counter App', () => {
  const appUrl = 'http://localhost:8080';
  const testUserId = '123456789';

  test.beforeEach(async ({ page }) => {
    // Mock the Telegram WebApp object
    await page.addInitScript(() => {
      window.Telegram = {
        WebApp: {
          initData: 'user=%7B%22id%22%3A123456789%2C%22first_name%22%3A%22Test%22%2C%22last_name%22%3A%22User%22%2C%22username%22%3A%22testuser%22%2C%22language_code%22%3A%22en%22%7D&chat_instance=-1234567890123456789&chat_type=private&auth_date=1672531200&hash=mock_hash',
          initDataUnsafe: {
            user: {
              id: 123456789,
              first_name: 'Test',
              last_name: 'User',
              username: 'testuser',
              language_code: 'en'
            }
          },
          ready: () => {},
        }
      };
    });

    // Intercept API requests to add the test user ID header
    await page.route('**/api/countries', (route) => {
      route.continue({
        headers: {
          ...route.request().headers(),
          'X-E2E-Test-User-Id': testUserId,
        },
        method: route.request().method(),
        postData: route.request().postData(),
      });
    });
  });

  test('should load the page and display the map', async ({ page }) => {
    await page.goto(appUrl);
    await expect(page.locator('#map')).toBeVisible();
  });

  test('should add a country and update the count', async ({ page }) => {
    await page.goto(appUrl);

    // Get the initial count
    const initialCount = await page.locator('#visited-count').textContent();

    // Type a country in the input
    await page.locator('#country-input').fill('Canada');

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
