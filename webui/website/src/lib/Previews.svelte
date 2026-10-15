<script lang="ts">
  import { tick } from 'svelte';
  import { exampleDocuments as documents } from './example-documents';
  import {
    Button,
    Input,
    SearchResultCard,
    PreviewHeader,
    PreviewContent,
    PreviewControls,
  } from '@hister/components';
  import { Eye, Globe, Search } from '@lucide/svelte';

  const lessons = [
    {
      title: 'Search what you remember',
      text: 'Search for raw HTML, even if you don’t remember the page title. The matching words are highlighted; use Preview on any result to read its stored text.',
      action: 'Next: narrow the search',
    },
    {
      title: 'Narrow to one site',
      text: 'Put “raw HTML” in quotes to match the phrase, then add domain:hister.org to search only that site. The two Markdown references disappear; the two Hister pages remain.',
      action: 'Start again',
    },
  ];
  let step = $state(0);
  const filtered = $derived(step > 0);
  const query = $derived(filtered ? '"raw HTML" domain:hister.org' : 'raw HTML');
  const results = $derived(
    documents.filter((result) => !filtered || result.domain === 'hister.org'),
  );
  let selected = $state<number | null>(0);

  function selectStep(next: number) {
    step = next;
    selected = 0;
    lastTrigger = undefined;
  }
  const document = $derived(selected === null ? null : documents[selected]);
  let stepButtons = $state<HTMLButtonElement[]>([]);
  let previewHeading = $state<HTMLHeadingElement>();
  let lastTrigger: HTMLElement | undefined;

  async function openPreview(index: number, event: MouseEvent) {
    lastTrigger = event.currentTarget as HTMLElement;
    selected = index;
    await tick();
    previewHeading?.focus();
  }

  async function closePreview() {
    selected = null;
    await tick();
    (lastTrigger ?? stepButtons[step])?.focus();
  }
</script>

<section id="product" class="product" aria-label="Hister web interface">
  <header class="product-top">
    <h2>Preserve Knowledge. Find It Again</h2>
    <p>You remember the words “raw HTML”. Find the page they came from.</p>
  </header>

  <nav class="lesson-steps" aria-label="Search walkthrough">
    {#each lessons as lesson, index}
      <button
        bind:this={stepButtons[index]}
        aria-current={step === index ? 'step' : undefined}
        onclick={() => selectStep(index)}
      >
        <span>0{index + 1}</span>{lesson.title}
      </button>
    {/each}
  </nav>
  <div class="lesson" aria-live="polite">
    <p>{lessons[step].text}</p>
    <Button variant="outline" size="sm" onclick={() => selectStep((step + 1) % lessons.length)}
      >{lessons[step].action}</Button
    >
  </div>
  <div class="example">
    <header class="example-header">
      <a href="https://demo.hister.org/" aria-label="Hister live demo"
        ><img src="/logo.png" alt="" width="24" height="24" /><span>HISTER</span></a
      ><span class="example-label">INTERACTIVE EXAMPLE</span>
    </header>
    <search class="query-bar" aria-label="Example search">
      <Search size={20} aria-hidden="true" />
      <Input
        aria-label="Example search query"
        value={query}
        readonly
        class="h-11 border-0 bg-transparent px-0 font-mono text-sm shadow-none focus-visible:ring-0"
      />
      <Button
        variant="outline"
        size="sm"
        href={`https://demo.hister.org/?q=${encodeURIComponent(query)}`}>Try this search</Button
      >
    </search>
    <div class="filter-bar">
      <span>Example collection</span>
      <Button variant="ghost" size="sm" aria-pressed={!filtered} onclick={() => selectStep(0)}
        >All sites <span>4</span></Button
      >
      <Button variant="ghost" size="sm" aria-pressed={filtered} onclick={() => selectStep(1)}
        >hister.org <span>2</span></Button
      >
      <output class="filter-feedback"
        >{filtered ? '2 of 4 pages match this site' : '4 pages · 3 sites'}</output
      >
    </div>
    <div class="panes" class:has-preview={document !== null}>
      <div class="results">
        <p class="result-count">{results.length} example results for <strong>{query}</strong></p>
        {#each results as result (result.path)}
          {@const index = documents.indexOf(result)}
          <SearchResultCard
            url={result.url}
            href={result.href}
            title={result.title}
            headingLevel={3}
            htmlSnippet={result.snippet}
            active={selected === index}
            class="mb-7"
          >
            {#snippet leading()}{#if result.domain === 'hister.org'}<img
                  src="/logo.png"
                  width="18"
                  height="18"
                  class="mt-1 shrink-0"
                  alt=""
                />{:else}<Globe size={18} class="mt-1 shrink-0" aria-hidden="true" />{/if}{/snippet}
            {#snippet metadata()}
              <div class="result-meta flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1">
                <span
                  class="result-url-text min-w-0 overflow-hidden text-xs text-ellipsis whitespace-nowrap"
                  title={result.url}
                  ><span class="result-domain font-semibold">{result.domain}</span><span
                    class="text-text-brand-muted">{result.path}</span
                  ></span
                >
                <Button
                  variant="link"
                  size="sm"
                  class="text-text-brand-muted hover:text-hister-indigo h-auto gap-1 p-0 text-xs"
                  aria-expanded={selected === index}
                  aria-controls="example-preview"
                  onclick={(event) => openPreview(index, event)}
                  ><Eye size={13} /><span>Preview<span class="sr-only"> {result.title}</span></span
                  ></Button
                >
              </div>
            {/snippet}
          </SearchResultCard>
        {/each}
        <p class="example-note">
          {filtered
            ? 'Filtered out: CommonMark Spec and Markdown Syntax Documentation. Remove the site filter to bring them back.'
            : 'The words match the content of all four pages, even though none of their titles contains “raw HTML”.'}
        </p>
      </div>
      {#if document}
        <section id="example-preview" class="preview" aria-label="Document preview">
          <PreviewHeader
            title={document.title}
            url={document.href}
            headingLevel={3}
            bind:heading={previewHeading}
          >
            {#snippet actions()}
              <PreviewControls onclose={closePreview} />
            {/snippet}
          </PreviewHeader>
          <PreviewContent>
            <span class="excerpt-label">STORED TEXT · EXCERPT</span>
            {#each document.paragraphs as paragraph, index}
              {#if index === 1}<h4>{document.heading}</h4>{/if}
              <p>
                {#each paragraph.split(/(raw HTML)/gi) as part}{#if part.toLowerCase() === 'raw html'}<mark
                      >{part}</mark
                    >{:else}{part}{/if}{/each}
              </p>
            {/each}
            <a href={document.href}>Continue reading</a>
          </PreviewContent>
        </section>
      {/if}
    </div>
  </div>
</section>

<style>
  .lesson-steps {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px;
  }
  .lesson-steps button {
    text-align: left;
    padding: 16px 0;
    border-bottom: 2px solid var(--border);
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
  }
  .lesson-steps button[aria-current='step'] {
    color: var(--text-primary);
    border-color: var(--hister-coral);
  }
  .lesson-steps span {
    margin-right: 12px;
    font-family: 'Fira Code', monospace;
    color: var(--hister-coral);
  }
  .lesson {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 32px;
    padding: 22px 0 26px;
    min-height: 124px;
  }
  .lesson p {
    max-width: 740px;
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 1.8;
    margin: 0;
  }
  .lesson :global(button) {
    flex-shrink: 0;
  }
  .query-bar {
    box-shadow: inset 3px 0 var(--hister-coral);
  }
  @media (max-width: 800px) {
    .lesson-steps {
      gap: 12px;
    }
    .lesson-steps button {
      font-size: 11px;
      line-height: 1.6;
    }
    .lesson-steps span {
      display: block;
    }
    .lesson {
      align-items: flex-start;
      flex-direction: column;
      gap: 16px;
    }
  }

  .product {
    max-width: 1240px;
    margin: auto;
    padding: 0 48px 64px;
  }
  .product-top {
    border-top: 1px solid var(--border);
    padding: 28px 0;
  }
  h2 {
    font:
      500 clamp(24px, 3vw, 34px)/1.2 'Inter',
      sans-serif;
    letter-spacing: -0.04em;
    margin: 0;
    max-width: 680px;
  }
  .product-top p {
    font-size: 13px;
    color: var(--text-secondary);
    margin: 14px 0 0;
    line-height: 1.7;
  }
  .example {
    border: 2px solid var(--brutal-border);
    background: var(--card-surface);
    color: var(--text-primary-brand);
  }
  .example-header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 2px solid var(--brutal-border);
  }
  .example-header a {
    display: flex;
    align-items: center;
    gap: 8px;
    font:
      800 23px 'Outfit',
      sans-serif;
    text-decoration: none;
  }
  .example-label {
    font:
      9px 'Fira Code',
      monospace;
    letter-spacing: 0.07em;
    color: var(--text-muted-brand);
  }
  .query-bar {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: 12px;
    padding: 8px 20px;
    border-bottom: 2px solid var(--border-muted-brand);
  }
  .query-bar > :global(svg) {
    flex-shrink: 0;
    color: var(--text-muted-brand);
  }
  .panes {
    display: grid;
    grid-template-columns: 1fr;
    min-height: 420px;
  }
  .panes.has-preview {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .results {
    min-width: 0;
    padding: 22px 24px;
  }
  .result-count {
    font:
      12px 'Inter',
      sans-serif;
    color: var(--hister-indigo);
    margin: 0 0 28px;
  }
  .example-note {
    border-top: 1px solid var(--border);
    padding-top: 20px;
    color: var(--text-muted-brand);
    font-size: 12px;
    line-height: 1.7;
  }
  .preview {
    min-width: 0;
    border-left: 2px solid var(--border-muted-brand);
  }
  .preview :global(p) {
    margin: 0 0 20px;
    line-height: 1.8;
  }
  .preview h4 {
    font:
      700 20px 'Outfit',
      sans-serif;
    margin: 0 0 14px;
  }
  .filter-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    padding: 12px 20px;
    border-bottom: 1px solid var(--border-muted-brand);
    font-size: 11px;
    color: var(--text-muted-brand);
  }
  .filter-bar > span:first-child {
    margin-right: 8px;
  }
  .filter-bar :global(button[aria-pressed='true']) {
    background: color-mix(in srgb, var(--hister-indigo) 14%, transparent);
    color: var(--text-primary-brand);
    box-shadow: inset 0 -2px var(--hister-indigo);
  }
  .filter-feedback {
    margin-left: auto;
  }
  .excerpt-label {
    display: block;
    font:
      10px 'Fira Code',
      monospace;
    color: var(--text-muted-brand);
    margin-bottom: 20px;
  }
  .preview mark {
    color: var(--text-primary-brand);
    background: color-mix(in srgb, var(--hister-amber) 38%, transparent);
    padding: 1px 3px;
  }
  @media (prefers-reduced-motion: no-preference) {
    .preview {
      animation: reveal 220ms ease-out;
    }
    @keyframes reveal {
      from {
        opacity: 0;
        transform: translateX(12px);
      }
    }
  }
  @media (max-width: 800px) {
    .product {
      padding: 16px 24px 40px;
    }
    .panes,
    .panes.has-preview {
      min-height: 0;
      grid-template-columns: 1fr;
    }
    .preview {
      border-left: 0;
      border-top: 2px solid var(--border-muted-brand);
    }
    .filter-bar > span:first-child,
    .filter-feedback {
      flex-basis: 100%;
      margin: 0;
    }
    .query-bar {
      grid-template-columns: auto minmax(0, 1fr);
      gap: 8px;
      padding: 12px 16px;
    }
    .query-bar :global(input) {
      font-size: 12px;
    }
    .query-bar :global([data-slot='button']) {
      grid-column: 2;
      justify-self: start;
    }
    .example-header {
      padding: 14px 16px;
    }
    .results {
      padding: 20px 16px;
    }
    .example-label {
      font-size: 8px;
    }
  }
</style>
