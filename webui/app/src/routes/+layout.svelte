<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { ModeWatcher, modeStorageKey, setMode } from 'mode-watcher';
  import SiteHeader from '$lib/components/SiteHeader.svelte';
  import { Toaster, toast } from '@hister/components/ui/sonner';
  import { fetchConfig, logout, resetConfig, type AppConfig } from '$lib/api';
  import { setFlashMessage, showFlashMessage } from '$lib/flash';
  import { base } from '$app/paths';
  import '../style.css';

  let { children } = $props();

  let config = $state<AppConfig | null>(null);
  // The library persists its generated system default during initialization.
  // Only dark and light indicate an explicit visitor choice.
  const storedMode = localStorage.getItem(modeStorageKey.current);
  const hasStoredModeOverride = storedMode === 'dark' || storedMode === 'light';

  onMount(() => {
    void showQueuedNotice();

    fetchConfig()
      .then((c) => {
        config = c;
        if (!hasStoredModeOverride) {
          setMode(c.colorScheme === 'automatic' ? 'system' : c.colorScheme);
        }
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
    window.location.href = isPublic ? base + '/' : base + '/auth';
  }
</script>

{#if config}
  <ModeWatcher defaultMode={config.colorScheme === 'automatic' ? 'system' : config.colorScheme} />
{/if}
<Toaster position="top-center" richColors offset="4.75rem" />

<div class="flex h-dvh flex-col overflow-hidden">
  <SiteHeader {config} onLogout={handleLogout} />

  <main class="flex min-h-0 flex-1 flex-col overflow-clip">
    {@render children()}
  </main>
</div>
