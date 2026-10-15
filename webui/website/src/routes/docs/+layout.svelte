<script lang="ts">
  import { page } from '$app/state';
  import BookOpenIcon from '@lucide/svelte/icons/book-open';
  import { afterNavigate } from '$app/navigation';
  import DocsNavigation from '#lib/DocsNavigation.svelte';
  import DocsSearch from '#lib/DocsSearch.svelte';

  let { children, data } = $props();
  let navigationOpen = $state(false);
  afterNavigate(() => (navigationOpen = false));

  const isIndex = $derived(page.url.pathname === '/docs' || page.url.pathname === '/docs/');

  const currentDoc = $derived(
    !isIndex ? data.docs.find((d) => page.url.pathname === `/docs/${d.slug}`) : null,
  );
  const currentCategory = $derived(
    !isIndex
      ? data.categories.find((category) =>
          category.docs.some((doc) => doc.slug === currentDoc?.slug),
        )
      : null,
  );
</script>

{#if isIndex}
  {@render children()}
{:else}
  <div class="docs-search-bar"><DocsSearch /></div>

  <div class="bg-brutal-bg sticky top-0 z-30 border-b border-(--border) md:hidden">
    <details class="group" bind:open={navigationOpen}>
      <summary
        class="flex cursor-pointer list-none items-center gap-3 px-6 py-3.5 [&::-webkit-details-marker]:hidden"
      >
        <BookOpenIcon aria-hidden="true" size={19} class="text-hister-indigo shrink-0" />
        <span class="min-w-0 flex-1">
          <span
            class="font-fira block text-[9px] font-bold tracking-[1.5px] text-(--text-secondary) uppercase"
            >{currentCategory?.name ?? 'Documentation'}</span
          >
          <span class="font-inter block truncate text-sm font-semibold text-(--text-primary)"
            >{currentDoc?.title}</span
          >
        </span>
        <span
          class="font-fira text-[9px] font-bold tracking-[1.25px] text-(--text-secondary) uppercase"
          >Browse</span
        >
        <span aria-hidden="true">{navigationOpen ? '−' : '+'}</span>
      </summary>
      <div class="max-h-[70dvh] overflow-y-auto border-t border-(--border) bg-(--page-bg) p-4">
        <DocsNavigation categories={data.categories} currentSlug={currentDoc?.slug ?? ''} />
      </div>
    </details>
  </div>

  <div class="docs-shell">
    <aside class="docs-sidebar">
      <div class="sticky top-6 max-h-[calc(100dvh-3rem)] overflow-y-auto pr-2 pb-4">
        <DocsNavigation categories={data.categories} currentSlug={currentDoc?.slug ?? ''} />
      </div>
    </aside>

    <div class="docs-main">
      <header class="reading-header docs-header">
        <nav aria-label="Breadcrumb" class="reading-breadcrumb">
          <a href="/docs">Documentation</a><span aria-hidden="true">/</span>
          <span>{currentCategory?.name}</span>
        </nav>
        <h1>{currentDoc?.title}</h1>
        {#if currentDoc?.description}<p>{currentDoc.description}</p>{/if}
      </header>

      {@render children()}
    </div>
  </div>
{/if}

<style>
  .docs-search-bar {
    max-width: 1440px;
    padding: 24px 48px;
    margin: auto;
    border-bottom: 1px solid var(--border);
  }
  .docs-search-bar :global(search) {
    max-width: min(640px, calc(100% - 256px));
    margin-left: 256px;
  }
  .docs-shell {
    display: grid;
    grid-template-columns: 208px minmax(0, 1fr);
    gap: 48px;
    max-width: 1440px;
    margin: auto;
    padding: 40px 48px 80px;
  }
  .docs-main {
    min-width: 0;
  }
  .docs-header {
    max-width: 760px;
    padding: 0;
    border: 0;
    margin: 0 0 40px;
  }
  .docs-header h1 {
    font-size: clamp(32px, 3.5vw, 46px);
    letter-spacing: -0.04em;
  }
  .docs-header > p {
    font-size: 18px;
    line-height: 1.7;
    margin-bottom: 0;
  }
  .reading-breadcrumb {
    margin-bottom: 24px;
  }
  @media (max-width: 1000px) {
    .docs-shell {
      grid-template-columns: 180px minmax(0, 1fr);
      gap: 32px;
      padding-inline: 32px;
    }
    .docs-search-bar {
      padding-inline: 32px;
    }
    .docs-search-bar :global(search) {
      max-width: min(640px, calc(100% - 212px));
      margin-left: 212px;
    }
  }
  @media (max-width: 767px) {
    .docs-shell {
      display: block;
      padding: 32px 24px 56px;
    }
    .docs-sidebar {
      display: none;
    }
    .docs-search-bar {
      padding: 24px;
    }
    .docs-search-bar :global(search) {
      max-width: 100%;
      margin-left: 0;
    }
    .docs-header > p {
      font-size: 17px;
    }
  }
</style>
