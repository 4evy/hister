<script lang="ts">
  import { ArrowRight, ArrowUpRight } from '@lucide/svelte';
  import BrowserIcon from '#lib/BrowserIcon.svelte';
  import CopyCommand from '#lib/CopyCommand.svelte';
</script>

<section id="how-it-works" class="landing-section setup">
  <header class="setup-heading">
    <h2 class="landing-heading">From install to your first search</h2>
    <a href="/docs/quickstart">Full setup guide <ArrowRight size={16} aria-hidden="true" /></a>
  </header>
  <ol role="list">
    <li>
      <span class="number">01 / RUN</span>
      <h3>Start Hister</h3>
      <p>After installing, run the server and leave it running while you browse.</p>
      <div class="step-action">
        <CopyCommand command="hister listen" label="Copy start command" />
      </div>
      <p class="note">
        Downloaded the binary? Use <code>./hister listen</code>, or <code>.\hister.exe listen</code>
        on Windows.
      </p>
      <p class="help"><a href="/docs/docker">Docker setup guide</a></p>
    </li>
    <li>
      <span class="number">02 / COLLECT</span>
      <h3>Connect your browser</h3>
      <p>
        The extension adds pages as you visit them. Its settings let you exclude sites you don’t
        want indexed.
      </p>
      <div class="step-action extensions">
        <a href="https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb">
          <BrowserIcon browser="chrome" />
          <span>Get for Chrome</span>
          <ArrowUpRight size={14} aria-hidden="true" />
        </a>
        <a href="https://addons.mozilla.org/en-US/firefox/addon/hister/">
          <BrowserIcon browser="firefox" />
          <span>Get for Firefox</span>
          <ArrowUpRight size={14} aria-hidden="true" />
        </a>
      </div>
      <p class="note">The default extension settings connect to the server on your computer.</p>
    </li>
    <li>
      <span class="number">03 / FIND</span>
      <h3>Search something you’ve read</h3>
      <p>Visit a few pages, then open Hister and search for a word from one of them.</p>
      <div class="step-action">
        <a class="local" href="http://localhost:4433">
          <span>Open Hister <span class="address">localhost:4433</span></span>
          <ArrowRight size={18} aria-hidden="true" />
        </a>
      </div>
      <p class="note">
        Already have a collection? <a href="/docs/import">Import bookmarks, history, or files</a>.
      </p>
    </li>
  </ol>
</section>

<style>
  .setup-heading {
    display: flex;
    justify-content: space-between;
    gap: 32px;
    align-items: end;
    margin-bottom: 38px;
  }
  .setup-heading h2 {
    max-width: 650px;
  }
  .setup-heading > a {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
  }
  ol {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 36px;
  }
  li {
    display: grid;
    grid-template-rows: subgrid;
    grid-row: span 6;
    row-gap: 0;
    border-top: 1px solid var(--border);
    padding-top: 24px;
  }
  .number {
    color: var(--hister-indigo);
    font:
      10px 'Fira Code',
      monospace;
    letter-spacing: 0.06em;
  }
  h3 {
    font:
      500 21px 'Inter',
      sans-serif;
    margin: 20px 0 14px;
    letter-spacing: -0.035em;
  }
  p {
    font-size: 14px;
    line-height: 1.8;
    color: var(--text-secondary);
    margin: 0;
  }
  a {
    font-size: 13px;
    text-underline-offset: 4px;
    text-decoration: underline;
    text-decoration-thickness: 1px;
  }
  a:hover {
    color: var(--hister-indigo);
  }
  .step-action {
    margin-top: 22px;
  }
  .local,
  .extensions a {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    min-height: 56px;
    padding: 14px;
    background: var(--brutal-card);
    border: 1px solid var(--border);
    text-decoration: none;
    font-weight: 500;
  }
  .local:hover,
  .extensions a:hover {
    color: var(--text-primary);
    border-color: var(--hister-indigo);
    background: color-mix(in srgb, var(--hister-indigo) 8%, var(--brutal-card));
  }
  .local > span {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 6px 12px;
  }
  .address {
    color: var(--text-secondary);
    font:
      11px 'Fira Code',
      monospace;
  }
  .help {
    margin-top: 10px;
    font-size: 12px;
  }
  .help a {
    font-size: inherit;
  }
  .note {
    margin-top: 22px;
    font-size: 12px;
  }
  .note code {
    font-size: 12px;
    white-space: nowrap;
  }
  .note a {
    font-size: inherit;
  }
  .extensions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .extensions a {
    flex: 1 1 auto;
    padding-inline: 12px;
  }
  .extensions a span {
    flex: 1;
    white-space: nowrap;
  }
  @media (max-width: 800px) {
    .setup-heading {
      display: block;
    }
    .setup-heading > a {
      display: inline-flex;
      margin-top: 22px;
    }
    li {
      display: block;
    }
    ol {
      grid-template-columns: 1fr;
      gap: 32px;
    }
  }
</style>
