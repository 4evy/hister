// SPDX-License-Identifier: AGPL-3.0-or-later
import { expect, test } from '@playwright/test';
import { mockAPI, previewPath } from './fixtures';

const issueURL = 'https://github.com/owner/repo/issues/305';
const issue = {
  title: 'Fix <img src=x onerror=alert(1)>',
  number: '305',
  repository: 'owner/repo',
  url: issueURL,
  state: 'Closed',
  labels: ['bug'],
  body: {
    author: 'alice',
    authorURL: 'https://github.com/alice',
    createdAt: '2026-04-09T07:47:32Z',
    url: `${issueURL}#issue-1`,
    bodyHTML: '<p>First paragraph</p><pre><code>go test ./...</code></pre><p>☑ Done</p>',
  },
  comments: [
    {
      author: 'bob',
      authorURL: 'https://github.com/bob',
      createdAt: 'on Apr 10, 2026',
      url: `${issueURL}#issuecomment-2`,
      bodyHTML: '<p>Try <code>--verbose</code></p>',
    },
  ],
};

test.beforeEach(async ({ page }) => {
  await mockAPI(page);
});

test('issue preview shows readable content and attributed comments', async ({ page }) => {
  await page.route('**/api/preview?**', (route) =>
    route.fulfill({
      json: { title: 'Saved issue', template: 'issue', content: JSON.stringify(issue) },
    }),
  );
  await page.goto(previewPath(issueURL));
  const article = page.getByRole('article', { name: 'Issue preview' });
  await expect(article.getByRole('heading', { name: issue.title, exact: true })).toBeVisible();
  await expect(article.locator('img')).toHaveCount(0);
  await expect(article).toContainText('owner/repo');
  await expect(article).toContainText('#305');
  await expect(article).toContainText('Closed');
  await expect(article).toContainText('bug');
  await expect(article.getByRole('link', { name: 'alice', exact: true })).toHaveAttribute(
    'href',
    'https://github.com/alice',
  );
  await expect(article.getByText('First paragraph', { exact: true })).toBeVisible();
  await expect(article.locator('pre code')).toHaveText('go test ./...');
  await expect(article).toContainText('☑ Done');
  const comments = article.getByRole('region', { name: 'Comments' });
  await expect(comments.getByRole('link', { name: 'bob', exact: true })).toHaveAttribute(
    'href',
    'https://github.com/bob',
  );
  await expect(comments.getByRole('link', { name: 'on Apr 10, 2026' })).toHaveAttribute(
    'href',
    `${issueURL}#issuecomment-2`,
  );
  await expect(comments.locator('code')).toHaveText('--verbose');
  await expect(article).toContainText('Comments shown are from the saved page.');
  expect(await article.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(true);
});

test('empty issue and missing attribution remain readable', async ({ page }) => {
  await page.route('**/api/preview?**', (route) =>
    route.fulfill({
      json: {
        template: 'issue',
        content: JSON.stringify({
          title: 'Empty issue',
          body: { bodyHTML: '' },
          comments: [{ bodyHTML: '<p>Anonymous reply</p>' }],
        }),
      },
    }),
  );
  await page.goto(previewPath(issueURL));
  const article = page.getByRole('article', { name: 'Issue preview' });
  await expect(article).toContainText('No description provided.');
  await expect(article).toContainText('Unknown author');
  await expect(article).toContainText('Anonymous reply');
  await expect(article).not.toContainText('undefined');
});

test('malformed issue data offers retry instead of rendering JSON as HTML', async ({ page }) => {
  for (const content of ['{"bodyHTML":"<img src=x>"', '{"title":"Bad","comments":{}}']) {
    await page.route('**/api/preview?**', (route) =>
      route.fulfill({ json: { template: 'issue', content } }),
    );
    await page.goto(previewPath(issueURL));
    await expect(page.getByRole('alert')).toContainText('Could not load this preview.');
    await expect(page.getByRole('button', { name: 'Retry', exact: true })).toBeVisible();
    await expect(page.locator('.preview-content img')).toHaveCount(0);
  }
});
