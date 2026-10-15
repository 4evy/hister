<script lang="ts">
  import ListIcon from '@lucide/svelte/icons/list';
  import ImageLightbox from '#lib/ImageLightbox.svelte';
  import Seo from '#lib/Seo.svelte';

  let { data } = $props();

  interface TocEntry {
    id: string;
    text: string;
    level: number;
  }

  let toc = $state<TocEntry[]>([]);
  let activeId = $state('');
  const activeEntry = $derived(toc.find((entry) => entry.id === activeId) ?? toc[0]);

  function closePageNavigation(event: MouseEvent) {
    (event.currentTarget as HTMLElement).closest('details')?.removeAttribute('open');
  }

  $effect(() => {
    // Track data.content as a dependency so this re-runs when navigating between docs
    void data.content;
    activeId = '';

    const article = document.querySelector('[data-doc-content]');
    if (!article) return;

    const headings = Array.from(article.querySelectorAll('h2, h3')).filter(
      (heading) => !heading.closest('.config-option'),
    );
    const nextToc = Array.from(headings).map((h) => ({
      id: h.id,
      text: h.textContent ?? '',
      level: h.tagName === 'H2' ? 2 : 3,
    }));
    toc = nextToc;
    activeId = nextToc[0]?.id ?? '';

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            activeId = entry.target.id;
          }
        }
      },
      { rootMargin: '-80px 0px -60% 0px', threshold: 0 },
    );

    headings.forEach((h) => observer.observe(h));
    return () => observer.disconnect();
  });
</script>

{#snippet mobilePageOutline()}
  <details class="page-outline group">
    <summary
      class="flex cursor-pointer list-none items-center gap-3 px-4 py-3 [&::-webkit-details-marker]:hidden"
    >
      <ListIcon aria-hidden="true" size={18} class="text-hister-indigo shrink-0" />
      <span class="min-w-0 flex-1">
        <span
          class="font-fira block text-[9px] font-bold tracking-[1.5px] text-(--text-secondary) uppercase"
          >On this page</span
        >
        <span class="font-inter block truncate text-sm font-semibold text-(--text-primary)"
          >{activeEntry?.text}</span
        >
      </span>
      <span aria-hidden="true" class="group-open:hidden">+</span>
      <span aria-hidden="true" class="hidden group-open:inline">−</span>
    </summary>
    <nav
      aria-label="On this page"
      class="max-h-[50dvh] overflow-y-auto border-t border-(--border) p-2"
    >
      {#each toc as entry}
        <a
          href="#{entry.id}"
          aria-current={activeId === entry.id ? 'location' : undefined}
          onclick={closePageNavigation}
          class="font-inter block border-l-2 py-2 text-sm no-underline transition-colors {entry.level ===
          3
            ? 'pl-6'
            : 'pl-3'} {activeId === entry.id
            ? 'border-hister-indigo bg-hister-indigo/10 font-semibold text-(--text-primary)'
            : 'border-transparent text-(--text-secondary) hover:bg-(--muted-surface) hover:text-(--text-primary)'}"
        >
          {entry.text}
        </a>
      {/each}
    </nav>
  </details>
{/snippet}

<Seo
  title={`${data.meta.title} | Hister Documentation`}
  description={(data.meta.description as string) ??
    `Learn about ${data.meta.title} in the Hister documentation.`}
  path={`/docs/${data.slug}`}
  type="article"
  publishedTime={data.meta.date as string | undefined}
  modifiedTime={(data.meta.updated as string | undefined) ?? (data.meta.date as string | undefined)}
  author="Hister contributors"
  section="Documentation"
  schemaType="TechArticle"
  schemaTitle={data.meta.title as string}
  breadcrumbs={[
    { name: 'Home', path: '/' },
    { name: 'Documentation', path: '/docs' },
    { name: data.meta.title as string, path: `/docs/${data.slug}` },
  ]}
/>

<div class="doc-layout">
  <article class="doc-article" data-doc-content>
    {#if toc.length > 0}
      {@render mobilePageOutline()}
    {/if}

    <div class="content doc-content">
      <data.content />
    </div>

    <nav aria-label="Documentation pagination" class="doc-pagination">
      {#if data.prev}
        <a href="/docs/{data.prev.slug}" rel="prev">
          <span>Previous</span>
          {data.prev.title}
        </a>
      {/if}
      {#if data.next}
        <a href="/docs/{data.next.slug}" rel="next">
          <span>Next</span>
          {data.next.title}
        </a>
      {/if}
    </nav>
  </article>

  <!-- TOC Sidebar (xl only) -->
  {#if toc.length > 0}
    <aside class="doc-toc">
      <nav
        aria-label="On this page"
        class="sticky top-6 flex max-h-[calc(100dvh-3rem)] flex-col gap-0.5 overflow-y-auto pb-4"
      >
        <span
          class="font-fira mb-3 text-[10px] font-bold tracking-[2px] text-(--text-secondary) uppercase"
          >On This Page</span
        >
        {#each toc as entry}
          <a
            href="#{entry.id}"
            aria-current={activeId === entry.id ? 'location' : undefined}
            class="font-inter h-auto justify-start rounded-none border-l-2 py-1 text-left text-[13px] whitespace-normal no-underline transition-colors
              {entry.level === 3 ? 'pl-5' : 'pl-3'}
              {activeId === entry.id
              ? 'border-hister-indigo font-medium text-(--text-primary)'
              : 'border-transparent text-(--text-secondary) hover:border-(--border) hover:text-(--text-primary)'}"
          >
            {entry.text}
          </a>
        {/each}
      </nav>
    </aside>
  {/if}
</div>

<ImageLightbox contentKey={data.slug} />

<style>
  .doc-layout {
    display: grid;
    grid-template-columns: minmax(0, 760px);
    gap: 40px;
  }
  .doc-article {
    min-width: 0;
  }
  .doc-toc {
    min-width: 0;
    display: none;
  }
  .doc-toc a {
    overflow-wrap: anywhere;
  }
  .page-outline {
    margin-bottom: 36px;
    border-block: 1px solid var(--border);
  }
  @media (min-width: 1440px) {
    .doc-layout {
      grid-template-columns: minmax(0, 760px) 176px;
    }
    .doc-toc {
      display: block;
    }
    .page-outline {
      display: none;
    }
  }

  .doc-pagination {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 24px;
    border-top: 1px solid var(--border);
    margin-top: 48px;
    padding-top: 24px;
  }
  .doc-pagination a {
    padding: 8px 0;
    font-size: 14px;
    text-decoration: none;
  }
  .doc-pagination a:hover {
    color: var(--hister-indigo);
  }
  .doc-pagination span {
    display: block;
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 6px;
  }
  .doc-pagination [rel='next'] {
    grid-column: 2;
    text-align: right;
  }
</style>
