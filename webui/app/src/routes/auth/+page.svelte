<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '@hister/components/ui/button';
  import { Input } from '@hister/components/ui/input';
  import * as Card from '@hister/components/ui/card';
  import { Lock } from '@lucide/svelte';
  import { toast } from '@hister/components/ui/sonner';
  import { fetchConfig, login, loginWithToken, resetConfig } from '#lib/api.js';
  import { setFlashMessage, showFlashMessage } from '#lib/flash.js';
  import { authReturnPath } from '#lib/auth-navigation.js';
  import { resolve } from '$app/paths';

  let authMode = $state<'token' | 'user' | 'none'>('token');
  let token = $state('');
  let username = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);
  let oauthProviders = $state<string[]>([]);
  let oauthOnly = $state(false);
  let basePath = $state(resolve('/').replace(/\/$/, ''));

  function returnDestination(): string {
    return authReturnPath(new URLSearchParams(window.location.search).get('return_to'), basePath);
  }

  onMount(() => {
    void initializeAuthPage();
  });

  async function initializeAuthPage() {
    const flashShown = await showFlashMessage();

    fetchConfig()
      .then((cfg) => {
        authMode = cfg.authMode ?? 'token';
        configureOAuth(cfg);
        basePath = cfg.basePath ?? '';
        redirectAuthenticated(cfg.authenticated);
      })
      .catch(() => {});

    if (flashShown) return;
    const params = new URLSearchParams(window.location.search);
    switch (params.get('reason')) {
      case 'invalid_token':
        toast.error('Your password was invalid or expired. Please enter it again.');
        break;
      case 'auth_required':
        toast.info('Please sign in to continue.');
    }
  }

  async function handleTokenSave() {
    error = '';
    loading = true;
    try {
      await loginWithToken(token);
      resetConfig();
      setFlashMessage('Signed in successfully.');
      window.location.href = returnDestination();
    } catch {
      error = 'Invalid password';
      toast.error(error);
    } finally {
      loading = false;
    }
  }

  async function handleLogin() {
    error = '';
    loading = true;
    try {
      await login(username, password);
      resetConfig();
      setFlashMessage('Signed in successfully.');
      window.location.href = returnDestination();
    } catch {
      error = 'Invalid username or password';
      toast.error(error);
    } finally {
      loading = false;
    }
  }

  function oauthLogin(provider: string) {
    const params = new URLSearchParams({ provider, return_to: returnDestination() });
    window.location.href = `${basePath}/api/oauth?${params}`;
  }

  const providerLabels: Record<string, string> = {
    github: 'GitHub',
    google: 'Google',
    oidc: 'SSO',
  };
  function configureOAuth(cfg: Awaited<ReturnType<typeof fetchConfig>>) {
    oauthProviders = cfg.oauthProviders ?? [];
    oauthOnly = cfg.oauthOnly ?? false;
  }
  function redirectAuthenticated(authenticated: boolean) {
    if (authMode === 'none' || authenticated) window.location.href = returnDestination();
  }
</script>

{#snippet oauthButtons()}
  {#if oauthProviders.length > 0}
    {#if !oauthOnly}
      <div class="relative flex items-center py-2">
        <div class="border-border-brand-muted flex-grow border-t"></div>
        <span class="text-text-brand-muted font-inter mx-3 flex-shrink text-xs uppercase">or</span>
        <div class="border-border-brand-muted flex-grow border-t"></div>
      </div>
    {/if}
    <div class="space-y-2">
      {#each oauthProviders as provider}
        <Button
          onclick={() => oauthLogin(provider)}
          variant="outline"
          class="border-brutal-border font-space h-11 w-full rounded-none border-[3px] font-bold tracking-wider uppercase shadow-[4px_4px_0px_var(--brutal-shadow)] transition-all hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-[2px_2px_0px_var(--brutal-shadow)] active:translate-x-1 active:translate-y-1 active:shadow-none"
        >
          Sign in with {providerLabels[provider] ?? provider}
        </Button>
      {/each}
    </div>
  {/if}
{/snippet}

<svelte:head>
  <title>Authentication - Hister</title>
</svelte:head>

<div class="bg-brutal-bg h-full overflow-y-auto">
  <div class="flex min-h-full items-center justify-center p-4">
    <Card.Root class="w-full max-w-md shadow-[8px_8px_0px_var(--hister-indigo)]">
      <Card.Header
        class="border-border-brand-muted flex-col space-y-4 border-b-[3px] pb-6 text-center"
      >
        <div class="flex justify-center">
          <div
            class="flex size-16 items-center justify-center rounded-full border-[3px]"
            style="background-color: color-mix(in srgb, var(--hister-indigo) 10%, transparent); border-color: var(--hister-indigo);"
          >
            <Lock class="text-hister-indigo size-8" />
          </div>
        </div>
        <h1 class="font-outfit text-text-brand text-2xl font-extrabold tracking-wide uppercase">
          Authentication Required
        </h1>
        <Card.Description class="font-inter text-text-brand-secondary">
          {#if authMode === 'user'}
            Please sign in.
          {:else}
            Please enter your password.
          {/if}
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-6">
        {#if error}
          <p role="alert" class="text-hister-rose font-inter text-sm">{error}</p>
        {/if}
        {#if authMode === 'user'}
          {#if !oauthOnly}
            <form
              method="POST"
              class="space-y-6"
              onsubmit={(event) => {
                event.preventDefault();
                if (!loading && username.trim() && password.trim()) void handleLogin();
              }}
            >
              <div class="space-y-2">
                <label
                  for="username"
                  class="font-space text-text-brand text-sm font-semibold tracking-wider uppercase"
                >
                  Username
                </label>
                <Input
                  id="username"
                  name="username"
                  autocomplete="username"
                  type="text"
                  variant="brutal"
                  bind:value={username}
                  placeholder="Enter your username"
                  class="focus-visible:border-hister-indigo"
                  autofocus
                />
              </div>
              <div class="space-y-2">
                <label
                  for="password"
                  class="font-space text-text-brand text-sm font-semibold tracking-wider uppercase"
                >
                  Password
                </label>
                <Input
                  id="password"
                  name="password"
                  autocomplete="current-password"
                  type="password"
                  variant="brutal"
                  bind:value={password}
                  placeholder="Enter your password"
                  class="focus-visible:border-hister-indigo font-mono"
                />
              </div>
              <Button
                type="submit"
                disabled={!username.trim() || !password.trim() || loading}
                class="bg-hister-indigo hover:bg-hister-indigo/90 border-brutal-border font-space h-12 w-full rounded-none border-[3px] font-bold tracking-wider uppercase shadow-[4px_4px_0px_var(--brutal-shadow)] transition-all hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-[2px_2px_0px_var(--brutal-shadow)] active:translate-x-1 active:translate-y-1 active:shadow-none disabled:cursor-not-allowed disabled:opacity-50"
              >
                {loading ? 'Signing in…' : 'Sign In'}
              </Button>
            </form>
          {/if}
          {@render oauthButtons()}
        {:else}
          <form
            method="POST"
            class="space-y-6"
            onsubmit={(event) => {
              event.preventDefault();
              if (!loading && token) void handleTokenSave();
            }}
          >
            <div class="space-y-2">
              <label
                for="token"
                class="font-space text-text-brand text-sm font-semibold tracking-wider uppercase"
              >
                Password
              </label>
              <Input
                id="token"
                name="token"
                autocomplete="current-password"
                type="password"
                variant="brutal"
                bind:value={token}
                placeholder="Enter your password"
                class="focus-visible:border-hister-indigo font-mono"
                autofocus
              />
            </div>
            <Button
              type="submit"
              disabled={!token || loading}
              class="bg-hister-indigo hover:bg-hister-indigo/90 border-brutal-border font-space h-12 w-full rounded-none border-[3px] font-bold tracking-wider uppercase shadow-[4px_4px_0px_var(--brutal-shadow)] transition-all hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-[2px_2px_0px_var(--brutal-shadow)] active:translate-x-1 active:translate-y-1 active:shadow-none disabled:cursor-not-allowed disabled:opacity-50"
            >
              {loading ? 'Signing in…' : 'Sign In'}
            </Button>
          </form>
        {/if}
      </Card.Content>
    </Card.Root>
  </div>
</div>
