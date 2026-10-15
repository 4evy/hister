<script lang="ts">
  const queries = [
    {
      query: '"self-hosted"',
      title: 'Remember a phrase?',
      copy: 'Put it in quotes to find those words together.',
    },
    {
      query: 'extractors domain:hister.org',
      title: 'Remember the site?',
      copy: 'Search within one domain, including the text inside its pages.',
    },
    {
      query: 'type:file',
      title: 'Looking for a local document?',
      copy: 'Limit the results to files you’ve added to your index.',
    },
    {
      query: 'added:<7d',
      title: 'Added it recently?',
      copy: 'Find documents first added during the last seven days.',
    },
  ];
</script>

<section id="features" class="landing-section">
  <h2 class="landing-heading">Search by phrase, site, or date</h2>
  <p class="intro">A few words are enough to start. Add details when you need to narrow it down.</p>
  <ul class="queries" role="list">
    {#each queries as example}
      <li class="query">
        <div>
          <h3>{example.title}</h3>
          <p>{example.copy}</p>
        </div>
        <code>{example.query}</code>
      </li>
    {/each}
  </ul>
  <a class="query-guide" href="/docs/query-language">Explore the query language</a>
  <div class="beyond-browser">
    <section>
      <h3>Prefer a terminal?</h3>
      <p>
        The terminal interface searches the same index. Use the CLI when you want to work from a
        script.
      </p>
      <a href="/docs/terminal-client">Terminal & CLI guide</a>
    </section>
    <section>
      <h3>Let your tools search it, too</h3>
      <p>
        Give an MCP client access to your saved pages, or query Hister directly through its HTTP
        API.
      </p>
      <a href="/docs/mcp">Connect an MCP client</a>
    </section>
  </div>
</section>

<style>
  .intro {
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 1.8;
    margin: 22px 0 34px;
  }
  .queries {
    list-style: none;
    margin: 0;
    padding: 0;
    border-top: 1px solid var(--border);
  }
  .query {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 32px;
    padding: 24px 0;
    border-bottom: 1px solid var(--border);
    align-items: center;
  }
  code {
    font:
      13px 'Fira Code',
      monospace;
    color: var(--hister-indigo);
    overflow-wrap: anywhere;
  }
  h3 {
    font:
      500 17px 'Inter',
      sans-serif;
    margin: 0 0 8px;
    letter-spacing: -0.025em;
  }
  p {
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 1.8;
    margin: 0;
  }
  a {
    display: inline-block;
    margin-top: 20px;
    font-size: 13px;
    text-underline-offset: 5px;
  }
  .beyond-browser {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 64px;
    margin-top: 56px;
  }
  .beyond-browser p {
    max-width: 420px;
  }
  @media (max-width: 650px) {
    .query {
      grid-template-columns: 1fr;
      gap: 14px;
    }
    .beyond-browser {
      grid-template-columns: 1fr;
      gap: 32px;
      margin-top: 40px;
    }
  }
</style>
