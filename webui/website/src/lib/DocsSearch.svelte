<script lang="ts">
  import { goto } from '$app/navigation';
  import SearchField from '#lib/SearchField.svelte';
  import type { DocsSearchEntry, DocsSearchIndex } from '#lib/docs-search.js';

  interface RankedResult {
    entry: DocsSearchEntry;
    label: string;
    excerpt: string;
    score: number;
  }

  interface HighlightPart {
    text: string;
    match: boolean;
  }

  let root: HTMLElement;
  let input = $state<HTMLInputElement>();
  let query = $state('');
  let index = $state<DocsSearchEntry[] | null>(null);
  let loading = $state(false);
  let loadError = $state('');
  let open = $state(false);
  let activeIndex = $state(0);

  const terms = $derived(tokenize(query));
  const results = $derived.by(() => rankResults(index ?? [], terms));
  const showPanel = $derived(open && query.trim().length > 0);

  $effect(() => {
    void query;
    activeIndex = 0;
  });

  function tokenize(value: string): string[] {
    return [...new Set(value.toLocaleLowerCase().match(/[\p{L}\p{N}_.:/-]+/gu) ?? [])];
  }

  function countMatches(value: string, term: string): number {
    let count = 0;
    let offset = 0;
    while ((offset = value.indexOf(term, offset)) !== -1) {
      count += 1;
      offset += term.length;
    }
    return count;
  }

  function excerptStart(source: string, firstMatch: number): number {
    const start = Math.max(0, firstMatch - 70);
    const nextSpace = source.indexOf(' ', start);
    if (start > 0 && nextSpace >= 0 && nextSpace < firstMatch) return nextSpace + 1;
    return start;
  }

  function excerptEnd(source: string, start: number, limit: number): number {
    const end = Math.min(source.length, limit);
    if (end === source.length) return end;
    const previousSpace = source.lastIndexOf(' ', end);
    return previousSpace > start ? previousSpace : end;
  }

  function excerptText(source: string, start: number, end: number): string {
    return `${start > 0 ? '…' : ''}${source.slice(start, end).trim()}${end < source.length ? '…' : ''}`;
  }

  function makeExcerpt(entry: DocsSearchEntry, searchTerms: string[]): string {
    const source = entry.content || entry.description;
    if (source.length <= 190) return source;
    const normalized = source.toLocaleLowerCase();
    const positions = searchTerms
      .map((term) => normalized.indexOf(term))
      .filter((position) => position >= 0);
    const firstMatch = positions.length > 0 ? Math.min(...positions) : 0;
    const start = excerptStart(source, firstMatch);
    const end = excerptEnd(source, start, Math.max(0, firstMatch - 70) + 190);
    return excerptText(source, start, end);
  }

  function titleMatchScore(value: string, term: string, weights: readonly number[]): number {
    if (value === term) return weights[0];
    if (value.startsWith(term)) return weights[1];
    return value.includes(term) ? weights[2] : 0;
  }

  function titleBonus(entry: DocsSearchEntry, title: string, searchTerms: string[]): number {
    if (!searchTerms.every((term) => title.includes(term))) return 0;
    return entry.heading ? 40 : 160;
  }

  function rankEntry(entry: DocsSearchEntry, searchTerms: string[]): RankedResult | null {
    const pageTitle = entry.pageTitle.toLocaleLowerCase();
    const heading = entry.heading?.toLocaleLowerCase() ?? '';
    const description = entry.description.toLocaleLowerCase();
    const content = entry.content.toLocaleLowerCase();
    const combined = `${pageTitle} ${heading} ${description} ${content}`;
    if (!searchTerms.every((term) => combined.includes(term))) return null;
    const score = searchTerms.reduce(
      (total, term) =>
        total +
        titleMatchScore(pageTitle, term, [50, 30, 18]) +
        titleMatchScore(heading, term, [45, 28, 16]) +
        Math.min(countMatches(description, term), 3) * 6 +
        Math.min(countMatches(content, term), 5) * 2,
      titleBonus(entry, pageTitle, searchTerms),
    );
    return {
      entry,
      label: entry.heading ?? entry.pageTitle,
      excerpt: makeExcerpt(entry, searchTerms),
      score,
    };
  }

  function rankResults(entries: DocsSearchEntry[], searchTerms: string[]): RankedResult[] {
    if (searchTerms.length === 0) return [];
    return entries
      .map((entry) => rankEntry(entry, searchTerms))
      .filter((result): result is RankedResult => result !== null)
      .sort(
        (a, b) =>
          b.score - a.score ||
          a.entry.pageTitle.localeCompare(b.entry.pageTitle) ||
          a.entry.url.localeCompare(b.entry.url),
      )
      .slice(0, 8);
  }

  function escapeRegex(value: string): string {
    return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  function highlightParts(value: string, searchTerms: string[]): HighlightPart[] {
    if (searchTerms.length === 0) return [{ text: value, match: false }];
    const expression = new RegExp(`(${searchTerms.map(escapeRegex).join('|')})`, 'giu');
    return value
      .split(expression)
      .filter(Boolean)
      .map((text) => ({
        text,
        match: searchTerms.some((term) => text.toLocaleLowerCase() === term),
      }));
  }

  async function fetchIndex(): Promise<DocsSearchIndex> {
    const response = await fetch('/docs/search.json');
    if (!response.ok) throw new Error(`Search index request failed with ${response.status}`);
    return response.json();
  }

  async function loadIndex() {
    if (index || loading) return;
    loading = true;
    loadError = '';
    try {
      const data = await fetchIndex();
      index = data.entries;
    } catch {
      loadError = 'Documentation search could not be loaded.';
    } finally {
      loading = false;
    }
  }

  function handleWindowClick(event: MouseEvent) {
    if (root && !root.contains(event.target as Node)) open = false;
  }

  function moveSelection(event: KeyboardEvent, offset: number) {
    event.preventDefault();
    activeIndex = (activeIndex + offset + results.length) % results.length;
  }

  function openSelection(event: KeyboardEvent) {
    event.preventDefault();
    open = false;
    void goto(results[activeIndex]?.entry.url ?? results[0].entry.url);
  }

  const navigationKeys = new Map<string, (event: KeyboardEvent) => void>([
    ['ArrowDown', (event) => moveSelection(event, 1)],
    ['ArrowUp', (event) => moveSelection(event, -1)],
    ['Enter', openSelection],
  ]);

  function canNavigate() {
    return showPanel && results.length > 0;
  }

  function handleInputKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      open = false;
      input?.blur();
      return;
    }
    if (!canNavigate()) return;
    navigationKeys.get(event.key)?.(event);
  }
</script>

{#snippet resultOptions()}
  {#each results as result, resultIndex (result.entry.url)}
    <a
      id="docs-search-result-{resultIndex}"
      href={result.entry.url}
      role="option"
      aria-selected={activeIndex === resultIndex}
      onmouseenter={() => (activeIndex = resultIndex)}
      onclick={() => (open = false)}
      class="result block border-b border-(--border) px-4 py-3 text-(--text-primary) no-underline last:border-b-0"
    >
      <span class="font-inter block text-sm leading-snug font-semibold">
        {#each highlightParts(result.label, terms) as part}
          {#if part.match}<mark>{part.text}</mark>{:else}{part.text}{/if}
        {/each}
      </span>
      {#if result.entry.heading}
        <span
          class="font-fira mt-1 block text-[9px] font-semibold tracking-[0.8px] text-(--text-secondary) uppercase"
        >
          {result.entry.pageTitle}
        </span>
      {/if}
      <span class="font-inter mt-1.5 block text-xs leading-relaxed text-(--text-secondary)">
        {#each highlightParts(result.excerpt, terms) as part}
          {#if part.match}<mark>{part.text}</mark>{:else}{part.text}{/if}
        {/each}
      </span>
    </a>
  {/each}
{/snippet}

<svelte:window onclick={handleWindowClick} />

<search bind:this={root} aria-label="Documentation" class="docs-search relative w-full">
  <SearchField
    label="Search all documentation"
    shortcuts
    bind:input
    bind:value={query}
    role="combobox"
    aria-autocomplete="list"
    aria-expanded={showPanel}
    aria-controls="docs-search-results"
    aria-activedescendant={showPanel && results.length > 0
      ? `docs-search-result-${activeIndex}`
      : undefined}
    aria-busy={loading}
    autocomplete="off"
    placeholder="Try OAuth, browser import, or semantic search"
    onfocus={() => {
      open = true;
      void loadIndex();
    }}
    oninput={() => (open = true)}
    onkeydown={handleInputKeydown}
  />

  {#if showPanel}
    <div
      id="docs-search-results"
      role="listbox"
      aria-label="Documentation search results"
      class="results bg-brutal-card absolute top-[calc(100%+0.5rem)] right-0 left-0 z-40 max-h-[min(65dvh,34rem)] overflow-y-auto border border-(--border) shadow-lg"
    >
      {#if loading}
        <p class="font-inter m-0 px-4 py-5 text-sm text-(--text-secondary)" role="status">
          Loading documentation search…
        </p>
      {:else if loadError}
        <p class="font-inter m-0 px-4 py-5 text-sm text-(--text-secondary)" role="status">
          {loadError}
        </p>
      {:else if results.length === 0}
        <p class="font-inter m-0 px-4 py-5 text-sm text-(--text-secondary)" role="status">
          No documentation matches “{query.trim()}”.
        </p>
      {:else}
        {@render resultOptions()}
      {/if}
    </div>
  {/if}
</search>

<style>
  .results {
    border-radius: 6px;
  }

  .result[aria-selected='true'],
  .result:hover {
    background: color-mix(in srgb, var(--hister-indigo) 14%, var(--brutal-card));
  }

  mark {
    color: inherit;
    background: color-mix(in srgb, var(--hister-amber) 65%, transparent);
    padding: 0 0.08em;
  }
</style>
