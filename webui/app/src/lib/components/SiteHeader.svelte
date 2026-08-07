<script lang="ts">
  import { page } from '$app/stores';
  import { base } from '$app/paths';
  import { Button } from '@hister/components/ui/button';
  import * as DropdownMenu from '@hister/components/ui/dropdown-menu';
  import { toggleMode, mode } from 'mode-watcher';
  import {
    ExternalLink,
    Keyboard,
    LogIn,
    LogOut,
    Menu,
    Moon,
    Sun,
    UserRound,
  } from '@lucide/svelte';
  import type { AppConfig } from '$lib/api';
  import { showHelp } from '$lib/stores';

  let { config, onLogout }: { config: AppConfig | null; onLogout: () => void } = $props();

  const navItems = [
    { label: 'History', href: 'history', color: 'var(--hister-indigo)' },
    { label: 'Rules', href: 'rules', color: 'var(--hister-teal)' },
    { label: 'Add', href: 'add', color: 'var(--hister-coral)' },
  ];

  const secondaryItems = [
    { label: 'Help', href: 'help', color: 'var(--hister-indigo)' },
    { label: 'Extractors', href: 'extractors', color: 'var(--hister-cyan)' },
    { label: 'About', href: 'about', color: 'var(--hister-teal)' },
    { label: 'API', href: 'api-docs', color: 'var(--hister-coral)' },
    {
      label: 'GitHub',
      href: 'https://github.com/asciimoo/hister/',
      color: 'var(--hister-amber)',
      external: true,
    },
  ];

  const menuItem =
    'font-space text-text-brand-muted data-[highlighted]:bg-muted-surface data-[highlighted]:text-text-brand cursor-pointer rounded-none px-3 py-2 text-xs font-semibold tracking-wider uppercase';

  const showWriteNav = $derived(!config?.public || !!config?.canWrite);
  const showLogin = $derived(
    !!config &&
      !config.authenticated &&
      (config.authMode === 'user' || config.authMode === 'token'),
  );
  const showLogout = $derived(
    !!config && config.authenticated && (config.authMode === 'user' || config.authMode === 'token'),
  );
  const showProfile = $derived(
    config?.authMode === 'user' && !!config.authenticated && !!config.username,
  );
  const appTitle = $derived(config?.title ?? 'Hister');
</script>

<header
  class="site-header bg-brutal-bg border-brutal-border sticky top-0 z-50 flex h-12 shrink-0 items-center justify-between gap-3 overflow-hidden border-b-[3px] px-3 md:h-16 md:px-6"
>
  <h1 class="flex min-w-0 items-center gap-1.5 md:gap-2">
    <a
      data-sveltekit-reload
      href="./"
      class="group flex min-w-0 items-center gap-1.5 no-underline md:gap-2"
    >
      <img src="static/logo.png" alt={`${appTitle} logo`} class="h-6 w-6 md:h-8 md:w-8" />
      <span
        class="font-space text-text-brand truncate text-lg font-extrabold tracking-[1px] uppercase group-hover:underline md:text-[28px] md:tracking-[2px]"
      >
        {appTitle}
      </span>
    </a>
  </h1>

  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button
          {...props}
          variant="ghost"
          size="icon"
          class="text-text-brand-muted hover:text-hister-indigo hover:bg-muted-surface size-8 shrink-0 transition-all md:size-10"
          title="Open menu"
          aria-label="Open navigation menu"
        >
          <Menu class="size-5 md:size-6" />
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>

    <DropdownMenu.Content
      align="end"
      sideOffset={8}
      class="border-brutal-border bg-card-surface w-64 rounded-none border-[3px] p-2 shadow-[4px_4px_0_var(--brutal-shadow)]"
    >
      {#if showWriteNav}
        {#each navItems as item (item.href)}
          {@const active = $page.route.id === `/${item.href}`}
          <DropdownMenu.Item
            class="primary-menu-item font-space cursor-pointer rounded-none border-l-4 px-3 py-2.5 text-sm font-extrabold tracking-widest uppercase {active
              ? 'is-active text-text-brand'
              : 'text-text-brand-muted'}"
            style="--menu-color: {item.color};"
          >
            {#snippet child({ props })}
              <a {...props} aria-current={active ? 'page' : undefined} href={item.href}>
                {item.label}
              </a>
            {/snippet}
          </DropdownMenu.Item>
        {/each}
        <DropdownMenu.Separator class="bg-border-brand-muted mx-0 my-2 h-[2px]" />
      {/if}

      {#each secondaryItems as item (item.href)}
        {@const active = !item.external && $page.route.id === `/${item.href}`}
        <DropdownMenu.Item
          class="secondary-menu-item {menuItem} {active ? 'is-active text-text-brand' : ''}"
          style="--menu-color: {item.color};"
        >
          {#snippet child({ props })}
            <a
              {...props}
              aria-current={active ? 'page' : undefined}
              href={item.href}
              target={item.external ? '_blank' : undefined}
              rel={item.external ? 'noopener' : undefined}
            >
              {item.label}
              {#if item.external}<ExternalLink class="ml-auto size-3.5" />{/if}
            </a>
          {/snippet}
        </DropdownMenu.Item>
      {/each}

      <DropdownMenu.Separator class="bg-border-brand-muted mx-0 my-2 h-[2px]" />

      {#if $page.route.id === '/'}
        <DropdownMenu.Item
          class={menuItem}
          onSelect={() => ($showHelp = !$showHelp)}
          textValue="Keyboard shortcuts"
        >
          <Keyboard class="size-4" />
          Keyboard shortcuts
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item class={menuItem} onSelect={() => toggleMode()} textValue="Toggle theme">
        {#if mode.current === 'dark'}
          <Sun class="size-4" />
          Light theme
        {:else}
          <Moon class="size-4" />
          Dark theme
        {/if}
      </DropdownMenu.Item>

      {#if showProfile || showLogin || showLogout}
        <DropdownMenu.Separator class="bg-border-brand-muted mx-0 my-2 h-[2px]" />
      {/if}
      {#if showProfile}
        <DropdownMenu.Item
          class={menuItem}
          onSelect={() => (window.location.href = base + '/profile')}
        >
          <UserRound class="size-4" />
          Profile
        </DropdownMenu.Item>
      {:else if showLogin}
        <DropdownMenu.Item
          class={menuItem}
          onSelect={() => (window.location.href = base + '/auth')}
        >
          <LogIn class="size-4" />
          Login
        </DropdownMenu.Item>
      {/if}
      {#if showLogout}
        <DropdownMenu.Item class={menuItem} onSelect={() => onLogout()}>
          <LogOut class="size-4" />
          {config?.username ? `Logout ${config.username}` : 'Logout'}
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</header>

<style>
  .site-header {
    box-shadow: 0 1px 0 color-mix(in srgb, white 6%, transparent) inset;
  }

  .primary-menu-item {
    border-left-color: var(--menu-color);
  }

  .primary-menu-item[data-highlighted],
  .primary-menu-item.is-active {
    color: color-mix(in srgb, var(--menu-color) 76%, var(--text-primary-brand));
    background: color-mix(in srgb, var(--menu-color) 9%, transparent);
  }

  .secondary-menu-item[data-highlighted],
  .secondary-menu-item.is-active {
    color: color-mix(in srgb, var(--menu-color) 78%, var(--text-primary-brand));
  }

  :global(.dark) .site-header {
    box-shadow: 0 1px 0 color-mix(in srgb, white 8%, transparent) inset;
  }

  :global(.dark) .primary-menu-item[data-highlighted],
  :global(.dark) .primary-menu-item.is-active {
    color: color-mix(in srgb, var(--menu-color) 84%, white);
    background: color-mix(in srgb, var(--menu-color) 13%, transparent);
  }

  :global(.dark) .secondary-menu-item[data-highlighted],
  :global(.dark) .secondary-menu-item.is-active {
    color: color-mix(in srgb, var(--menu-color) 86%, white);
  }
</style>
