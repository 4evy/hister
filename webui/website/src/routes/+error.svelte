<script lang="ts">
  import { page } from '$app/state';
  import DocsSearch from '#lib/DocsSearch.svelte';

  const notFound = $derived(page.status === 404);
  const title = $derived(notFound ? 'Page not found' : 'Something went wrong');
  const description = $derived(
    notFound
      ? 'This page may have moved or the link may be out of date. Find what you need below.'
      : 'Hister could not load this page. Try again or continue from one of the links below.',
  );
  const destinations = [
    { href: '/docs', label: 'Documentation', description: 'Install, configure, and use Hister' },
    {
      href: '/datasets',
      label: 'Datasets',
      description: 'Find collections for your search engine',
    },
    { href: '/posts', label: 'Posts', description: 'Read news, ideas, and practical guides' },
    { href: '/support', label: 'Support', description: 'Get involved and help Hister grow' },
  ];
</script>

<svelte:head>
  <title>{page.status} {title} | Hister</title>
  <meta name="description" content={description} />
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<section class="reading-index error-page">
  <header class="reading-header">
    <p class="reading-kicker">Error {page.status}</p>
    <h1>{title}</h1>
    <p>{description}</p>
    <div class="error-actions">
      <a href="/">Go home</a>
      {#if !notFound}<a href={page.url.pathname} data-sveltekit-reload>Try again</a>{/if}
    </div>
    <DocsSearch />
  </header>
  <nav aria-label="Explore Hister">
    <h2>Explore Hister</h2>
    <ul>
      {#each destinations as destination}
        <li>
          <a href={destination.href}
            ><h3>{destination.label}</h3>
            <p>{destination.description}</p></a
          >
        </li>
      {/each}
    </ul>
  </nav>
</section>

<style>
  .error-page {
    max-width: 1000px;
  }
  .error-actions {
    display: flex;
    gap: 24px;
    font-size: 14px;
  }
  .error-actions a {
    color: var(--text-primary);
    text-underline-offset: 4px;
  }
  nav {
    border-top: 1px solid var(--border);
    padding-top: 28px;
  }
  nav h2 {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary);
    margin: 0 0 24px;
  }
  nav ul {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 28px 40px;
    padding: 0;
    margin: 0;
    list-style: none;
  }
  nav a {
    display: block;
    color: var(--text-primary);
    text-decoration: none;
  }
  nav a:hover h3 {
    text-decoration: underline;
    text-underline-offset: 4px;
  }
  nav h3 {
    font-size: 16px;
    font-weight: 500;
    margin: 0 0 8px;
  }
  nav p {
    font-size: 13px;
    line-height: 1.6;
    color: var(--text-secondary);
    margin: 0;
  }
  @media (max-width: 600px) {
    nav ul {
      grid-template-columns: 1fr;
    }
  }
</style>
