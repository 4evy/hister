<script lang="ts">
  import Install from '#lib/Install.svelte';
  import { Button } from '@hister/components';
  import * as DropdownMenu from '@hister/components/ui/dropdown-menu';
</script>

<section class="hero" aria-labelledby="hero-title">
  <div class="intro">
    <h1 id="hero-title">Your Own<br /> <span>Search Engine</span></h1>
    <p class="description">
      Find a page by what it said. Hister indexes the text of pages you visit and files you add, on
      a machine you control.
    </p>
    <div class="actions">
      <a class="landing-button" href="/docs/quickstart">Quickstart</a>
      <DropdownMenu.Root>
        <DropdownMenu.Trigger>
          {#snippet child({ props })}<Button {...props} variant="outline" class="extension-button"
              >Download extension</Button
            >{/snippet}
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          align="start"
          class="bg-card-surface rounded-md border border-(--border) p-2"
        >
          <DropdownMenu.Item
            >{#snippet child({ props })}<a
                {...props}
                href="https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb"
                >Chrome extension</a
              >{/snippet}</DropdownMenu.Item
          >
          <DropdownMenu.Item
            >{#snippet child({ props })}<a
                {...props}
                href="https://addons.mozilla.org/en-US/firefox/addon/hister/">Firefox extension</a
              >{/snippet}</DropdownMenu.Item
          >
        </DropdownMenu.Content>
      </DropdownMenu.Root>
      <a class="text-link" href="https://demo.hister.org/">Try the live demo</a>
    </div>
    <Install />
    <p class="platforms">LINUX · MACOS · WINDOWS <span>AGPLv3</span></p>
  </div>
</section>

<style>
  .hero {
    max-width: 1240px;
    margin: auto;
    padding: 84px 48px 72px;
    position: relative;
    isolation: isolate;
  }
  .intro {
    position: relative;
    z-index: 1;
    max-width: 720px;
  }
  h1 {
    font:
      500 clamp(40px, 4.5vw, 65px)/1.1 'Inter',
      sans-serif;
    letter-spacing: -0.055em;
    margin: 0;
  }
  h1 span {
    color: var(--hister-indigo);
  }
  .description {
    max-width: 540px;
    font-size: 15px;
    line-height: 1.85;
    color: var(--text-secondary);
    margin: 26px 0;
  }
  .actions {
    display: flex;
    gap: 16px;
    flex-wrap: wrap;
    align-items: center;
  }
  .actions .text-link {
    color: var(--text-primary);
    font-size: 14px;
    font-weight: 500;
    line-height: 1.5;
    white-space: nowrap;
    text-decoration: underline;
    text-underline-offset: 5px;
    border-bottom: 0;
    background: var(--brutal-bg);
    padding: 12px 4px;
  }
  .actions :global(.extension-button) {
    height: auto;
    padding: 13px 16px;
    font-size: 13px;
  }
  .platforms {
    display: flex;
    justify-content: space-between;
    max-width: 620px;
    margin-top: 15px;
    color: var(--text-secondary);
    font:
      9px 'Fira Code',
      monospace;
    letter-spacing: 0.08em;
  }
  @media (max-width: 800px) {
    .hero {
      padding: 48px 24px;
    }
    h1 {
      font-size: clamp(38px, 7vw, 58px);
    }
    .description {
      max-width: 540px;
    }
  }
</style>
