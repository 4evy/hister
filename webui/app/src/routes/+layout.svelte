<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { ModeWatcher, setMode } from 'mode-watcher';
  import SiteHeader from '#lib/components/SiteHeader.svelte';
  import { Toaster, toast } from '@hister/components/ui/sonner';
  import { fetchConfig, logout, resetConfig, type AppConfig } from '#lib/api.js';
  import { setFlashMessage, showFlashMessage } from '#lib/flash.js';
  import { getStoredThemePreference } from '#lib/theme.js';
  import { resolve } from '$app/paths';
  import '../style.css';

  let { children } = $props();

  let config = $state<AppConfig | null>(null);
  const storedThemePreference = getStoredThemePreference();

  onMount(() => {
    void showQueuedNotice();

    fetchConfig()
      .then((c) => {
        config = c;
        setMode(
          storedThemePreference ?? (c.colorScheme === 'automatic' ? 'system' : c.colorScheme),
        );
      })
      .catch(() => {});
  });

  async function showQueuedNotice() {
    if (await showFlashMessage()) return;
    await tick();
    const params = new URLSearchParams(window.location.search);
    if (params.get('notice') === 'logged_out') {
      toast.success('You have been logged out.');
      params.delete('notice');
      const qs = params.toString();
      window.history.replaceState({}, '', `${window.location.pathname}${qs ? `?${qs}` : ''}`);
    }
  }

  async function handleLogout() {
    await logout();
    const isPublic = config?.public;
    resetConfig();
    config = null;
    setFlashMessage('You have been logged out.');
    window.location.href = isPublic ? resolve('/') : resolve('auth');
  }
</script>

{#if config}
  <ModeWatcher
    defaultMode={storedThemePreference ??
      (config.colorScheme === 'automatic' ? 'system' : config.colorScheme)}
  />
{/if}
<Toaster position="top-center" richColors offset="4.75rem" />

<div class="flex h-dvh flex-col overflow-hidden">
  <SiteHeader {config} onLogout={handleLogout} />

  <main class="flex min-h-0 flex-1 flex-col overflow-clip">
    {@render children()}
  </main>
</div>
