// SPDX-License-Identifier: AGPL-3.0-or-later
import { expect, test, type Page } from '@playwright/test';

const original = 'https://example.com/article';
const preview = {
  title: 'Saved article',
  content: '<p>Saved article content</p>',
  added: 1700000000,
  version_count: 1,
};

function previewPath(url = original, version?: number) {
  const params = new URLSearchParams({ id: url, document_id: 'document-1', title: 'Article' });
  if (version) params.set('version', String(version));
  return `/preview?${params}`;
}

async function mockAPI(page: Page) {
  await page.routeWebSocket('**/search', () => {});
  await page.route('**/api/**', async (route) => {
    const path = new URL(route.request().url()).pathname;
    const data: Record<string, unknown> = {
      '/api/config': {
        wsUrl: 'ws://127.0.0.1:4174/search',
        title: 'Hister',
        searchUrl: 'https://example.com/search?q=',
        hotkeys: {},
        authMode: 'none',
        authenticated: true,
        canWrite: true,
        historyEnabled: true,
        colorScheme: 'automatic',
        search: {
          version: 0,
          fields: [],
          facets: [],
          sort: { field: '', label: '', description: '', options: [] },
          valueSets: {},
        },
      },
      '/api/stats': {
        doc_count: 2,
        recent_searches: [{ query: 'svelte' }, { query: 'golang' }],
      },
      '/api/rules': { aliases: {} },
      '/api/preview': preview,
      '/api/versions': [],
      '/api/extractors': [
        { name: 'basic', description: 'Stored content' },
        { name: 'readability', description: 'Readable content' },
      ],
    };
    await route.fulfill({ json: data[path] ?? {} });
  });
}

test.beforeEach(async ({ page }) => {
  await mockAPI(page);
});

test('recent searches explain browser scope and hiding never deletes server history', async ({
  page,
}) => {
  const mutations: string[] = [];
  page.on('request', (request) => {
    if (request.url().includes('/api/') && request.method() !== 'GET') {
      mutations.push(request.url());
    }
  });
  await page.goto('/');
  const recents = page.getByRole('region', { name: 'Recent searches' });
  await expect(recents).toContainText('Your history remains on the server.');
  await recents.getByRole('button', { name: 'Hide recent search on this browser: svelte' }).click();
  await expect(recents.getByRole('button', { name: 'svelte', exact: true })).toHaveCount(0);
  await expect(recents.getByRole('button', { name: 'golang', exact: true })).toBeVisible();
  await page.reload();
  await expect(recents.getByRole('button', { name: 'golang', exact: true })).toBeVisible();
  await expect(recents.getByRole('button', { name: 'svelte', exact: true })).toHaveCount(0);
  await recents.getByRole('button', { name: 'Hide all on this browser' }).click();
  await expect(recents).toHaveCount(0);
  await page.reload();
  await expect(page.getByRole('link', { name: 'Browse all indexed pages' })).toBeVisible();
  await expect(recents).toHaveCount(0);
  expect(mutations).toEqual([]);
});

for (const failure of ['server', 'network'] as const) {
  test(`a ${failure} preview failure can retry the same archived document`, async ({ page }) => {
    const requests: URLSearchParams[] = [];
    await page.route('**/api/preview?**', async (route) => {
      requests.push(new URL(route.request().url()).searchParams);
      if (requests.length === 1) {
        if (failure === 'network') await route.abort();
        else await route.fulfill({ status: 500, body: 'Internal server details' });
      } else {
        await route.fulfill({
          json: { ...preview, version_id: 7, version_created_at: '2023-11-14T22:13:20Z' },
        });
      }
    });
    await page.goto(previewPath(original, 7));
    await expect(page.getByRole('alert')).toContainText('Could not load this preview.');
    await expect(page.getByRole('link', { name: 'Open original', exact: true })).toHaveAttribute(
      'href',
      original,
    );
    await expect(page.locator('body')).not.toContainText('Internal server details');
    await page.getByRole('button', { name: 'Retry', exact: true }).click();
    await expect(page.getByText('Saved article content', { exact: true })).toBeVisible();
    await expect(page.getByText('Viewing archived version from', { exact: false })).toBeVisible();
    await expect(page.getByRole('alert')).toHaveCount(0);
    expect(requests).toHaveLength(2);
    for (const request of requests) {
      expect(request.get('url')).toBe(original);
      expect(request.get('document_id')).toBe('document-1');
      expect(request.get('version')).toBe('7');
    }
  });

  test(`a ${failure} version history failure has a retry instead of an empty list`, async ({
    page,
  }) => {
    let attempts = 0;
    await page.route('**/api/versions?**', async (route) => {
      attempts++;
      if (attempts === 1) {
        if (failure === 'network') await route.abort();
        else await route.fulfill({ status: 500 });
      } else {
        await route.fulfill({
          json: [{ id: 7, created_at: '2023-11-14T22:13:20Z', text_diff: '+Older content' }],
        });
      }
    });
    await page.goto(previewPath());
    await page.getByRole('button', { name: '1 previous version', exact: true }).click();
    await expect(page.getByRole('alert')).toContainText('Could not load previous versions.');
    await page.getByRole('button', { name: 'Retry', exact: true }).click();
    await expect(page.getByRole('button', { name: 'show this version' })).toBeVisible();
    await expect(page.getByRole('alert')).toHaveCount(0);
    expect(attempts).toBe(2);
  });
}

test('version history distinguishes loading from an empty successful response', async ({
  page,
}) => {
  let finish: () => void = () => {};
  const pending = new Promise<void>((resolve) => {
    finish = resolve;
  });
  await page.route('**/api/versions?**', async (route) => {
    await pending;
    await route.fulfill({ json: [] });
  });
  await page.goto(previewPath());
  await page.getByRole('button', { name: '1 previous version', exact: true }).click();
  await expect(page.getByRole('status')).toHaveText('Loading previous versions…');
  finish();
  await expect(page.getByText('No previous versions available.', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: '1 previous version', exact: true }).click();
  await expect(page.getByText('Saved article content', { exact: true })).toBeVisible();
});

test('preview errors do not offer unusable or unsafe original links', async ({ page }) => {
  await page.route('**/api/preview?**', (route) => route.fulfill({ status: 500 }));
  for (const url of [
    'remote-file://host/file.txt',
    'file:///private/file.txt',
    'javascript:alert(1)',
  ]) {
    await page.goto(previewPath(url));
    await expect(page.getByRole('alert')).toContainText('Could not load this preview.');
    await expect(page.getByRole('button', { name: 'Retry', exact: true })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Open original', exact: true })).toHaveCount(0);
  }
});

test('retry preserves an explicitly selected extractor', async ({ page }) => {
  const requests: URLSearchParams[] = [];
  await page.route('**/api/preview?**', async (route) => {
    requests.push(new URL(route.request().url()).searchParams);
    if (requests.length === 2) await route.fulfill({ status: 500 });
    else await route.fulfill({ json: preview });
  });
  await page.goto(previewPath());
  await page.getByRole('button', { name: 'Change extractor' }).click();
  await page.getByRole('menuitemradio', { name: 'readability', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('Could not load this preview.');
  await page.getByRole('button', { name: 'Retry', exact: true }).click();
  await expect(page.getByText('Saved article content', { exact: true })).toBeVisible();
  expect(requests).toHaveLength(3);
  expect(requests[1].get('extractor')).toBe('readability');
  expect(requests[2].toString()).toBe(requests[1].toString());
});

test('a delayed preview failure does not replace a newly selected document', async ({ page }) => {
  let finish: () => void = () => {};
  const pending = new Promise<void>((resolve) => {
    finish = resolve;
  });
  await page.route('**/api/preview?**', async (route) => {
    if (new URL(route.request().url()).searchParams.get('url') === original) {
      await pending;
      await route.fulfill({ status: 500 });
    } else {
      await route.fulfill({ json: preview });
    }
  });
  await page.goto(previewPath());
  await expect(page.getByRole('status')).toHaveText('Loading…');
  // The standalone preview uses popstate to select a different document in the same panel.
  await page.evaluate((path) => {
    history.pushState({}, '', path);
    window.dispatchEvent(new PopStateEvent('popstate'));
  }, previewPath('https://example.com/another-article'));
  await expect(page.getByText('Saved article content', { exact: true })).toBeVisible();
  finish();
  await page.waitForLoadState('networkidle');
  await expect(page.getByRole('alert')).toHaveCount(0);
  await expect(page.getByText('Saved article content', { exact: true })).toBeVisible();
});
