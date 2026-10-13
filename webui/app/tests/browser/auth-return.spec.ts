// SPDX-License-Identifier: AGPL-3.0-or-later
import { expect, test, type Page } from '@playwright/test';
import { mockAPI, previewPath } from './fixtures';

async function mockLogin(page: Page, authMode: 'token' | 'user', publicMode = false) {
  const config = {
    basePath: '',
    authMode,
    authenticated: false,
    canWrite: false,
    public: publicMode,
    oauthProviders: ['oidc'],
  };
  await mockAPI(page, { config });
  await page.route(`**/api/${authMode === 'token' ? 'token-login' : 'login'}`, async (route) => {
    const body = route.request().postDataJSON();
    if ((body.token ?? body.password) !== 'correct') {
      await route.fulfill({ status: 401 });
      return;
    }
    config.authenticated = true;
    config.canWrite = true;
    await route.fulfill({ json: { username: 'alice' } });
  });
  return config;
}

async function signIn(page: Page, mode: 'token' | 'user', password = 'correct') {
  if (mode === 'user') await page.getByRole('textbox', { name: 'Username' }).fill('alice');
  await page.getByLabel('Password', { exact: true }).fill(password);
  await page.getByRole('button', { name: 'Sign In', exact: true }).click();
}

for (const mode of ['token', 'user'] as const) {
  test(`${mode} login returns to a protected preview after a failed attempt`, async ({ page }) => {
    const config = await mockLogin(page, mode);
    await page.route('**/api/preview?*', async (route) => {
      if (config.authenticated) await route.fallback();
      else await route.fulfill({ status: 403 });
    });
    const target = previewPath() + '#saved-section';
    await page.goto(target);
    await expect(page).toHaveURL(
      (url) => url.pathname === '/auth' && url.searchParams.get('return_to') === target,
    );
    await signIn(page, mode, 'wrong');
    await expect(page).toHaveURL((url) => url.pathname === '/auth');
    await signIn(page, mode);
    await expect(page).toHaveURL(new URL(target, 'http://127.0.0.1:4174').href);
    await expect(page.getByText('Saved article content')).toBeVisible();
  });
}

for (const path of ['/add', '/rules', '/profile']) {
  test(`public mode login returns to ${path}`, async ({ page }) => {
    await mockLogin(page, 'user', true);
    const target = path + '?source=bookmark#saved-section';
    await page.goto(target);
    await expect(page).toHaveURL((url) => url.searchParams.get('return_to') === target);
    await signIn(page, 'user');
    await expect(page).toHaveURL(new URL(target, 'http://127.0.0.1:4174').href);
  });
}

test('a rejected configuration request preserves the search query and fragment', async ({
  page,
}) => {
  const config = await mockLogin(page, 'token');
  await page.route('**/api/config', async (route) => {
    if (!config.authenticated && new URL(page.url()).pathname !== '/auth') {
      await route.fulfill({ status: 403 });
    } else {
      await route.fallback();
    }
  });
  const target = '/?q=needle#results';
  await page.goto(target);
  await expect(page).toHaveURL((url) => url.searchParams.get('return_to') === target);
  await signIn(page, 'token');
  await expect(page).toHaveURL(new URL(target, 'http://127.0.0.1:4174').href);
});

test('OAuth receives the local return destination', async ({ page }) => {
  await mockLogin(page, 'user');
  let returnTo: string | null = null;
  await page.route('**/api/oauth?*', async (route) => {
    returnTo = new URL(route.request().url()).searchParams.get('return_to');
    await route.fulfill({ contentType: 'text/plain', body: 'OAuth login started' });
  });
  const target = previewPath() + '#saved-section';
  await page.goto('/auth?' + new URLSearchParams({ return_to: target }));
  await page.getByRole('button', { name: 'Sign in with SSO' }).click();
  await expect(page.getByText('OAuth login started')).toBeVisible();
  expect(returnTo).toBe(target);
});

test('an authenticated visitor follows the requested page without another login', async ({
  page,
}) => {
  await mockAPI(page);
  await page.goto('/auth?' + new URLSearchParams({ return_to: '/help#query-syntax' }));
  await expect(page).toHaveURL('http://127.0.0.1:4174/help#query-syntax');
});

test('manual login from the menu preserves the current page', async ({ page }) => {
  await mockLogin(page, 'user', true);
  await page.goto('/help#query-syntax');
  await page.getByRole('button', { name: 'Open navigation menu' }).click();
  await page.getByRole('menuitem', { name: 'Login', exact: true }).click();
  await expect(page).toHaveURL((url) => url.searchParams.get('return_to') === '/help#query-syntax');
  await signIn(page, 'user');
  await expect(page).toHaveURL('http://127.0.0.1:4174/help#query-syntax');
});

test('token mode does not send the unavailable profile page through a login loop', async ({
  page,
}) => {
  const config = await mockLogin(page, 'token');
  config.authenticated = true;
  config.canWrite = true;
  await page.goto('/auth?' + new URLSearchParams({ return_to: '/profile' }));
  await expect(page).toHaveURL('http://127.0.0.1:4174/');
});

for (const [target, expected] of [
  ['/hister/preview?id=article#saved-section', '/hister/preview?id=article#saved-section'],
  ['/preview', '/hister/'],
  ['/hister-other/preview', '/hister/'],
  ['/hister/%2e%2e/preview', '/hister/'],
  ['/hister/%2f%2foutside.example', '/hister/'],
  ['/hister/api/oauth?provider=oidc', '/hister/'],
]) {
  test(`base path login validates ${target}`, async ({ page }) => {
    const config = await mockLogin(page, 'token');
    config.basePath = '/hister';
    await page.route('http://127.0.0.1:4174/hister/**', (route) =>
      route.fulfill({ contentType: 'text/plain', body: 'Returned to Hister' }),
    );
    await page.goto('/auth?' + new URLSearchParams({ return_to: target }));
    await signIn(page, 'token');
    await expect(page).toHaveURL(new URL(expected, 'http://127.0.0.1:4174').href);
  });
}

for (const target of [
  '',
  'https://outside.example/',
  '//outside.example/',
  '/\\outside.example/',
  '/%xx',
  '/preview\n',
  '/api/oauth?provider=oidc',
  '/auth',
]) {
  test(`login rejects return destination ${target}`, async ({ page }) => {
    await mockLogin(page, 'token');
    await page.goto('/auth?' + new URLSearchParams({ return_to: target }));
    await signIn(page, 'token');
    await expect(page).toHaveURL('http://127.0.0.1:4174/');
  });
}
