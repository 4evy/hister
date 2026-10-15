<script lang="ts">
  import ArrowRight from '@lucide/svelte/icons/arrow-right';
  import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
  import BookOpen from '@lucide/svelte/icons/book-open';
  import FolderOpen from '@lucide/svelte/icons/folder-open';
  import Users from '@lucide/svelte/icons/users';
  import Search from '@lucide/svelte/icons/search';
  import Server from '@lucide/svelte/icons/server';
  import Code from '@lucide/svelte/icons/code';
  import Globe from '@lucide/svelte/icons/globe';
  import DocsSearch from '#lib/DocsSearch.svelte';
  import Seo from '#lib/Seo.svelte';

  let { data } = $props();
  const groups = [
    { icon: BookOpen, description: 'Install Hister and find your first page.' },
    { icon: FolderOpen, description: 'Bring websites, files, and bookmarks into your index.' },
    { icon: Users, description: 'Manage accounts, indexing rules, and stored data.' },
    { icon: Search, description: 'Search from the terminal and connect other tools.' },
    { icon: Server, description: 'Configure, deploy, and maintain your server.' },
    { icon: Code, description: 'Work on Hister and build integrations.' },
  ];
</script>

<Seo
  title="Documentation | Hister"
  description="Learn how to install, configure, and use Hister across the web, terminal, browser extensions, API, and MCP."
  path="/docs"
/>

<section class="reading-index docs-index">
  <header class="reading-header">
    <h1>Documentation</h1>
    <p>Set up your search engine, add content, and make it yours.</p>
    <DocsSearch />
  </header>
  <section class="docs-start" aria-labelledby="docs-start-title">
    <div>
      <h2 id="docs-start-title">New to Hister?</h2>
      <p>Start a server, connect your browser, and run your first search.</p>
    </div>
    <a href="/docs/quickstart">Follow the quickstart</a>
  </section>
  <div class="docs-directory">
    {#each data.categories as category, i}
      {@const group = groups[i]}
      <section aria-labelledby={`category-${i}`}>
        <header>
          <group.icon size={20} aria-hidden="true" />
          <h2 id={`category-${i}`}>{category.name}</h2>
        </header>
        <p>{group.description}</p>
        <ul>
          {#each category.docs as doc}<li>
              <a href="/docs/{doc.slug}"
                ><span>{doc.title}</span><ArrowRight size={16} aria-hidden="true" /></a
              >
            </li>{/each}
        </ul>
      </section>
    {/each}
    <section aria-labelledby="extensions-title">
      <header>
        <Globe size={20} aria-hidden="true" />
        <h2 id="extensions-title">Browser extensions</h2>
      </header>
      <p>Save the pages you visit, as you browse.</p>
      <ul>
        <li>
          <a
            href="https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb"
            target="_blank"
            rel="noopener noreferrer"
            ><span>Chrome Extension</span><ArrowUpRight size={16} aria-hidden="true" /></a
          >
        </li>
        <li>
          <a
            href="https://addons.mozilla.org/en-US/firefox/addon/hister/"
            target="_blank"
            rel="noopener noreferrer"
            ><span>Firefox Add-on</span><ArrowUpRight size={16} aria-hidden="true" /></a
          >
        </li>
      </ul>
    </section>
  </div>
</section>

<style>
  .docs-start {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    padding: 24px;
    margin-bottom: 48px;
    border-radius: 6px;
    background: var(--muted-surface);
  }
  .docs-start h2 {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 6px;
  }
  .docs-start p,
  .docs-directory p {
    color: var(--text-secondary);
    font-size: 13px;
    line-height: 1.65;
  }
  .docs-start a {
    flex-shrink: 0;
    padding: 11px 16px;
    background: var(--hister-indigo);
    color: #18191b;
    font-size: 13px;
    font-weight: 500;
    border-radius: 4px;
    text-decoration: none;
  }
  .docs-start a:hover {
    background: var(--hister-coral);
  }
  .docs-directory {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 40px 64px;
  }
  .docs-directory header {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 8px;
  }
  .docs-directory header :global(svg) {
    color: var(--text-secondary);
    flex-shrink: 0;
  }
  .docs-directory h2 {
    font-size: 17px;
    font-weight: 600;
    letter-spacing: -0.02em;
  }
  .docs-directory p {
    margin: 0 0 14px 32px;
    max-width: 340px;
  }
  .docs-directory ul {
    list-style: none;
    padding: 0 0 0 32px;
    margin: 0;
  }
  .docs-directory a {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-height: 40px;
    max-width: 100%;
    padding: 7px 0;
    font-size: 14px;
    text-decoration: none;
  }
  .docs-directory a :global(svg) {
    flex-shrink: 0;
    color: var(--text-secondary);
  }
  .docs-directory a:hover,
  .docs-directory a:focus-visible {
    text-decoration: underline;
    text-underline-offset: 4px;
    color: var(--hister-indigo);
  }
  @media (max-width: 700px) {
    .docs-directory {
      grid-template-columns: minmax(0, 1fr);
      gap: 32px;
    }
    .docs-start {
      align-items: start;
      flex-direction: column;
    }
  }
</style>
