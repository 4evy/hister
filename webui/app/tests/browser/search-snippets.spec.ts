import { expect, test } from '@playwright/test';

import { mockAPI, original } from './fixtures';

test('long search snippets are fully displayed', async ({ page }) => {
  await mockAPI(page);
  const context = 'Surrounding search result context. '.repeat(25);
  await page.routeWebSocket('**/search', (socket) => {
    socket.onMessage(() =>
      socket.send(
        JSON.stringify({
          documents: [
            {
              id: 'document-1',
              url: original,
              title: 'Article',
              domain: 'example.com',
              text: `<mark>needle</mark> ${context}`,
              added: 1700000000,
            },
          ],
          total: 1,
        }),
      ),
    );
  });
  await page.goto('/?q=needle');
  const snippet = page.locator('.result-snippet');
  await expect(snippet).toContainText(context.trim());
  await expect(snippet.locator('mark')).toHaveText('needle');
  await expect
    .poll(() => snippet.evaluate((element) => element.scrollHeight > element.clientHeight + 1))
    .toBe(false);
  expect(
    await snippet.evaluate(
      (element) => element.clientHeight / parseFloat(getComputedStyle(element).lineHeight),
    ),
  ).toBeGreaterThan(3);
});
