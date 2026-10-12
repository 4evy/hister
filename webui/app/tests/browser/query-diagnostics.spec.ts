import { expect, test } from '@playwright/test';
import { mockAPI } from './fixtures';

test('query errors are visible and clear after correcting the query', async ({ page }) => {
  await mockAPI(page);
  await page.routeWebSocket('**/search', (socket) => {
    socket.onMessage((message) => {
      const query = JSON.parse(String(message));
      socket.send(
        JSON.stringify({
          query,
          documents: [],
          history: [],
          total: 0,
          ...(query.text.endsWith('OR')
            ? { error: 'invalid query at character 8: expected a term after OR' }
            : {}),
        }),
      );
    });
  });
  await page.goto('/?q=needle%20OR');
  await expect(page.getByRole('alert')).toContainText('at character 8: expected a term after OR');
  await expect(page.getByText('No results found for', { exact: false })).toHaveCount(0);
  const input = page.locator('input').filter({ visible: true }).first();
  await expect(input).toHaveValue('needle OR');
  await input.fill('needle');
  await expect(page.getByRole('alert')).toHaveCount(0);
  await expect(page.getByText('No results found for', { exact: false })).toBeVisible();
});
