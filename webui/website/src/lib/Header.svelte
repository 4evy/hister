<script lang="ts">
  import CodebergIcon from '#lib/CodebergIcon.svelte';
  import GitHubIcon from '#lib/GitHubIcon.svelte';
  import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
  import ThemeSwitch from '#lib/ThemeSwitch.svelte';
  import { page } from '$app/state';
  import { afterNavigate } from '$app/navigation';
  let menuOpen = $state(false);
  let menuButton: HTMLButtonElement;
  const links = [
    { href: '/docs', label: 'Documentation' },
    { href: '/datasets', label: 'Datasets' },
    { href: '/posts', label: 'Posts' },
    { href: '/support', label: 'Support' },
  ];
  afterNavigate(() => (menuOpen = false));
  function closeOnEscape(event: KeyboardEvent) {
    if (event.key === 'Escape' && menuOpen) {
      menuOpen = false;
      menuButton.focus();
    }
  }
</script>

<svelte:window onkeydown={closeOnEscape} />
<header>
  <div class="header-bar">
    <a class="wordmark" href="/" aria-label="Hister home"
      ><img src="/logo.png" alt="" width="28" height="28" />hister</a
    >
    <nav class="desktop-links" aria-label="Primary">
      {#each links as link}<a
          href={link.href}
          aria-current={page.url.pathname.startsWith(link.href) ? 'location' : undefined}
          >{link.label}</a
        >{/each}
    </nav>
    <div class="utilities">
      <a
        class="source-link"
        href="https://github.com/asciimoo/hister"
        aria-label="Hister source on GitHub"
        ><GitHubIcon />GitHub<ArrowUpRight size={14} aria-hidden="true" /></a
      >
      <a
        class="source-link"
        href="https://codeberg.org/asciimoo/hister"
        aria-label="Hister source on Codeberg"
        ><CodebergIcon />Codeberg<ArrowUpRight size={14} aria-hidden="true" /></a
      >
      <ThemeSwitch />
    </div>
    <button
      class="menu-button"
      bind:this={menuButton}
      onclick={() => (menuOpen = !menuOpen)}
      aria-expanded={menuOpen}
      aria-controls="mobile-navigation">{menuOpen ? 'Close −' : 'Menu +'}</button
    >
  </div>
  {#if menuOpen}<nav id="mobile-navigation" aria-label="Primary">
      {#each links as link}<a
          href={link.href}
          aria-current={page.url.pathname.startsWith(link.href) ? 'location' : undefined}
          onclick={() => (menuOpen = false)}>{link.label}</a
        >{/each}<a class="mobile-source" href="https://github.com/asciimoo/hister"
        ><GitHubIcon />Source on GitHub<ArrowUpRight size={14} aria-hidden="true" /></a
      >
      <a class="mobile-source" href="https://codeberg.org/asciimoo/hister"
        ><CodebergIcon />Source on Codeberg<ArrowUpRight size={14} aria-hidden="true" /></a
      >
    </nav>{/if}
</header>

<style>
  header {
    border-bottom: 1px solid var(--border);
  }
  :global(.home-theme) header {
    border-bottom: 0;
  }
  :global(.home-theme) .header-bar {
    max-width: 1240px;
  }
  .header-bar {
    max-width: 1240px;
    margin: auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 21px 48px;
    gap: 32px;
  }
  a {
    text-decoration: none;
  }
  .wordmark {
    font:
      800 29px 'Outfit',
      sans-serif;
    letter-spacing: -1.5px;
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .utilities {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-left: auto;
  }
  .desktop-links {
    margin-left: auto;
    display: flex;
    gap: 29px;
  }
  .desktop-links a,
  .source-link,
  .menu-button {
    font-size: 13px;
  }
  .desktop-links a:hover,
  .desktop-links a[aria-current],
  .source-link:hover {
    text-decoration: underline;
    text-underline-offset: 5px;
  }
  .source-link,
  .mobile-source {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 44px;
    white-space: nowrap;
  }
  .mobile-source:has(+ .mobile-source) {
    margin-top: 8px;
    border-top: 1px solid var(--border);
  }
  .menu-button {
    display: none;
    cursor: pointer;
    padding: 10px 0 10px 16px;
  }
  #mobile-navigation {
    padding: 15px 24px 25px;
    border-top: 1px solid var(--border);
    display: grid;
  }
  #mobile-navigation a {
    padding: 12px 0;
    font-size: 14px;
  }
  @media (max-width: 940px) {
    .header-bar {
      padding: 18px 24px;
    }
    .desktop-links,
    .source-link {
      display: none;
    }
    .menu-button {
      display: block;
    }
  }
  @media (min-width: 941px) {
    #mobile-navigation {
      display: none;
    }
  }
</style>
