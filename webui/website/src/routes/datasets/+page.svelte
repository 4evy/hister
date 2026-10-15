<script lang="ts">
  import CodebergIcon from '#lib/CodebergIcon.svelte';
  import GitHubIcon from '#lib/GitHubIcon.svelte';
  import ArrowUpRight from '@lucide/svelte/icons/arrow-up-right';
  import Download from '@lucide/svelte/icons/download';
  import Check from '@lucide/svelte/icons/check';
  import SearchField from '#lib/SearchField.svelte';
  import X from '@lucide/svelte/icons/x';
  import Seo from '#lib/Seo.svelte';
  import DatasetExample from '#lib/DatasetExample.md';
  import { modal } from '#lib/modal.js';
  import type { Dataset } from './+page.ts';

  let { data } = $props();

  const allTags = $derived([...new Set(data.datasets.flatMap((d: Dataset) => d.tags))].sort());
  const allLicenses = $derived([...new Set(data.datasets.map((d: Dataset) => d.license))].sort());
  const allAuthors = $derived([...new Set(data.datasets.map((d: Dataset) => d.author))].sort());

  let searchQuery = $state('');
  let selectedTags = $state<string[]>([]);
  let selectedLicense = $state('');
  let selectedAuthor = $state('');
  let submitModalOpen = $state(false);

  const submissionFields = [
    {
      title: 'About the dataset',
      fields: [
        ['name', 'Display name'],
        ['description', 'A short summary of the contents'],
        ['author', 'Author or publisher'],
        ['license', 'License identifier, such as CC-BY 4.0'],
        ['tags', 'A list of topic strings; optional'],
        ['image', 'Cover image URL; optional, or null'],
      ],
    },
    {
      title: 'Download and updates',
      fields: [
        ['downloadUrl', 'Direct URL to the Hister dataset JSON file'],
        ['diskSizeBytes', 'Download size in bytes (number); optional, or null'],
        ['documentCount', 'Number of indexed documents (number); optional, or null'],
        ['latestUpdate', 'Date string in YYYY-MM-DD format; optional, or null'],
      ],
    },
  ];

  function formatBytes(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  function matchesQuery(d: Dataset): boolean {
    const q = searchQuery.trim().toLowerCase();
    return [d.name, d.description, d.license, d.author, ...d.tags].some((value) =>
      value.toLowerCase().includes(q),
    );
  }

  function matchesTags(d: Dataset): boolean {
    return selectedTags.every((tag) => d.tags.includes(tag));
  }

  function matchesLicense(d: Dataset): boolean {
    return !selectedLicense || d.license === selectedLicense;
  }

  function matchesAuthor(d: Dataset): boolean {
    return !selectedAuthor || d.author === selectedAuthor;
  }

  const queryMatches = $derived(data.datasets.filter(matchesQuery));
  const baseWithoutTags = $derived(queryMatches.filter(matchesLicense).filter(matchesAuthor));
  const baseWithoutLicense = $derived(queryMatches.filter(matchesTags).filter(matchesAuthor));
  const baseWithoutAuthor = $derived(queryMatches.filter(matchesTags).filter(matchesLicense));
  const filtered = $derived(baseWithoutTags.filter(matchesTags));

  // How many results would remain if this tag were toggled on (or kept if already on)
  function tagCount(tag: string): number {
    const tagsToCheck = selectedTags.includes(tag) ? selectedTags : [...selectedTags, tag];
    return baseWithoutTags.filter((d) => tagsToCheck.every((t) => d.tags.includes(t))).length;
  }

  function licenseCount(license: string): number {
    return baseWithoutLicense.filter((d) => d.license === license).length;
  }

  function authorCount(author: string): number {
    return baseWithoutAuthor.filter((d) => d.author === author).length;
  }

  const hasFilters = $derived(
    searchQuery.trim() !== '' ||
      selectedTags.length > 0 ||
      selectedLicense !== '' ||
      selectedAuthor !== '',
  );

  function toggleTag(tag: string) {
    if (selectedTags.includes(tag)) {
      selectedTags = selectedTags.filter((t) => t !== tag);
    } else {
      selectedTags = [...selectedTags, tag];
    }
  }

  function clearFilters() {
    searchQuery = '';
    selectedTags = [];
    selectedLicense = '';
    selectedAuthor = '';
  }

  function closeSubmitModal() {
    submitModalOpen = false;
  }
</script>

{#snippet datasetResults()}
  <ul class="dataset-list">
    {#each filtered as dataset (dataset.slug)}
      <li class="dataset-row">
        <div class="dataset-details">
          {#if dataset.image}<img src={dataset.image} alt="" class="dataset-image" />{/if}
          <h2>{dataset.name}</h2>
          <p class="dataset-author">{dataset.author}</p>
          <p class="dataset-description">{dataset.description}</p>
          <ul class="dataset-topics" aria-label="Topics">
            {#each dataset.tags as tag}<li>
                <button aria-pressed={selectedTags.includes(tag)} onclick={() => toggleTag(tag)}
                  >{#if selectedTags.includes(tag)}<Check
                      size={12}
                      aria-hidden="true"
                    />{/if}{tag}</button
                >
              </li>{/each}
          </ul>
        </div>
        <div class="dataset-facts">
          <dl>
            {#if dataset.documentCount !== null}<div>
                <dt>Documents</dt>
                <dd>{dataset.documentCount.toLocaleString('en-US')}</dd>
              </div>{/if}
            {#if dataset.diskSizeBytes !== null}<div>
                <dt>Download size</dt>
                <dd>{formatBytes(dataset.diskSizeBytes)}</dd>
              </div>{/if}
            <div>
              <dt>License</dt>
              <dd>{dataset.license}</dd>
            </div>
            {#if dataset.latestUpdate !== null}<div>
                <dt>Updated</dt>
                <dd>{dataset.latestUpdate.replaceAll('.', '-')}</dd>
              </div>{/if}
          </dl>
          <a
            class="dataset-download"
            href={dataset.downloadUrl}
            aria-label={`Download ${dataset.name}`}
            download><Download size={16} aria-hidden="true" />Download dataset</a
          >
        </div>
      </li>
    {/each}
  </ul>
{/snippet}

<Seo
  title="Datasets | Hister"
  description="Browse public datasets compatible with Hister. Filter by tag, license or author and download what you need."
  path="/datasets"
/>

<section class="reading-index datasets-page">
  <header class="reading-header">
    <h1>Datasets</h1>
    <p>
      Give your search engine a head start with public collections of documentation and reference
      material.
    </p>
    <div class="dataset-intro-links">
      <a href="/docs/import">How to import a dataset</a>
      <button onclick={() => (submitModalOpen = true)} aria-haspopup="dialog"
        >Submit a dataset</button
      >
    </div>
  </header>

  <search aria-label="Datasets" class="dataset-filters">
    <div class="dataset-search">
      <SearchField
        label="Search datasets"
        shortcuts
        placeholder="Try Python, Mozilla, or documentation"
        bind:value={searchQuery}
      />
    </div>
    <label class="facet-select"
      >Author
      <select bind:value={selectedAuthor}>
        <option value="">All authors</option>
        {#each allAuthors as author}<option value={author}>{author} ({authorCount(author)})</option
          >{/each}
      </select>
    </label>
    <label class="facet-select"
      >License
      <select bind:value={selectedLicense}>
        <option value="">All licenses</option>
        {#each allLicenses as license}<option value={license}
            >{license} ({licenseCount(license)})</option
          >{/each}
      </select>
    </label>
    <fieldset class="tag-filters">
      <legend>Topics</legend>
      {#each allTags as tag}
        <label class:active={selectedTags.includes(tag)}>
          <input type="checkbox" value={tag} bind:group={selectedTags} />
          {tag}<span>{tagCount(tag)}</span>
        </label>
      {/each}
    </fieldset>
  </search>

  <div class="result-summary">
    <output aria-live="polite">{filtered.length} of {data.datasets.length} datasets</output>
    {#if hasFilters}<button onclick={clearFilters}>Clear filters</button>{/if}
  </div>
  {#if filtered.length}
    {@render datasetResults()}
  {:else}
    <div class="dataset-empty">
      <h2>No matching datasets</h2>
      <p>Try a different search or clear your filters to see all collections.</p>
      <button onclick={clearFilters}>Clear filters</button>
    </div>
  {/if}
</section>

<!-- Submit dataset modal -->
{#if submitModalOpen}
  <dialog
    id="submit-dataset-dialog"
    use:modal
    onclose={closeSubmitModal}
    aria-labelledby="submit-modal-title"
    aria-describedby="submit-modal-description"
    class="bg-brutal-card m-auto max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-2xl overflow-y-auto border border-(--border) p-0 text-(--text-primary) shadow-xl md:max-h-[calc(100dvh-5rem)]"
  >
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-(--border) px-6 py-4">
      <h2 id="submit-modal-title" class="font-inter text-xl font-medium text-(--text-primary)">
        Submit a New Dataset
      </h2>
      <button
        onclick={closeSubmitModal}
        aria-label="Close"
        class="cursor-pointer text-(--text-secondary) transition-colors hover:text-(--text-primary)"
      >
        <X size={18} />
      </button>
    </div>

    <!-- Body -->
    <div class="submission-body flex flex-col gap-6 px-6 py-6">
      <p id="submit-modal-description" class="submission-intro">
        Add a JSON file using the example below, give it a descriptive filename, and submit it as a
        pull request.
      </p>
      <div class="submission-actions">
        <a
          href="https://github.com/asciimoo/hister/new/master/webui/website/src/content/datasets?filename=your-dataset-name.json"
          class="submission-action"
        >
          <GitHubIcon />
          Add dataset on GitHub
          <ArrowUpRight size={16} aria-hidden="true" />
        </a>
        <a
          href="https://codeberg.org/asciimoo/hister/_new/master/webui/website/src/content/datasets/"
          class="submission-action"
          ><CodebergIcon />Add dataset on Codeberg<ArrowUpRight size={16} aria-hidden="true" /></a
        >
      </div>

      <figure class="dataset-example content m-0 flex min-w-0 flex-col gap-2">
        <figcaption class="font-inter text-[11px] font-medium text-(--text-secondary)">
          JSON Format
        </figcaption>
        <DatasetExample />
      </figure>

      {#each submissionFields as group}
        <section class="submission-reference" aria-label={group.title}>
          <h3>{group.title}</h3>
          <dl class="submission-fields">
            {#each group.fields as [field, description]}
              <dt><code>{field}</code></dt>
              <dd>{description}</dd>
            {/each}
          </dl>
        </section>
      {/each}
    </div>
  </dialog>
{/if}

<style>
  .dataset-example :global(pre) {
    margin: 0;
    padding: 16px;
    border-radius: 4px;
    --shiki-background: var(--muted-surface);
    font-size: 13px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  dialog::backdrop {
    background: rgb(0 0 0 / 70%);
  }
  .reading-header {
    margin-bottom: 32px;
  }
  .dataset-intro-links {
    display: flex;
    flex-wrap: wrap;
    gap: 12px 24px;
    font-size: 13px;
  }
  .dataset-intro-links a,
  .dataset-intro-links button {
    color: var(--text-primary);
    text-decoration: underline;
    text-underline-offset: 4px;
    cursor: pointer;
  }
  .dataset-search {
    min-width: 0;
  }
  .dataset-filters {
    display: grid;
    grid-template-columns: minmax(0, 2fr) repeat(2, minmax(0, 1fr));
    align-items: end;
    gap: 20px;
    margin-bottom: 24px;
  }
  .facet-select {
    display: grid;
    gap: 8px;
    font-size: 12px;
    color: var(--text-secondary);
  }
  .facet-select select {
    width: 100%;
    min-width: 0;
    padding: 14px 10px;
    border: 1px solid var(--border);
    border-radius: 6px;
    color: var(--text-primary);
    background: var(--brutal-card);
    font-size: 13px;
  }
  .tag-filters {
    grid-column: 1 / -1;
    padding: 0;
    border: 0;
    margin: 0;
  }
  .tag-filters legend {
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 10px;
  }
  .tag-filters label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-height: 28px;
    margin: 0 8px 0 0;
    font-size: 12px;
    color: var(--text-secondary);
    cursor: pointer;
  }
  .tag-filters input {
    accent-color: var(--hister-indigo);
  }
  .tag-filters label.active {
    color: var(--text-primary);
  }
  .tag-filters label span {
    font-size: 10px;
    margin-right: 10px;
  }
  .result-summary {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 16px;
  }
  .result-summary button,
  .dataset-empty button {
    text-decoration: underline;
    text-underline-offset: 4px;
    cursor: pointer;
  }
  .dataset-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  .dataset-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 230px;
    gap: 48px;
    padding: 28px 0;
    border-top: 1px solid var(--border);
  }
  .dataset-details {
    min-width: 0;
  }
  .dataset-row h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 500;
    line-height: 1.4;
    letter-spacing: -0.035em;
  }
  .dataset-author {
    margin: 6px 0 14px;
    color: var(--text-secondary);
    font-size: 12px;
  }
  .dataset-description {
    max-width: 650px;
    margin: 0 0 16px;
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 1.7;
  }
  .dataset-topics {
    list-style: none;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding: 0;
    margin: 0;
  }
  .dataset-topics button {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    min-height: 32px;
    border: 0;
    border-radius: 4px;
    padding: 4px 8px;
    background: var(--muted-surface);
    font-size: 11px;
    color: var(--text-secondary);
    cursor: pointer;
  }
  .dataset-topics button[aria-pressed='true'] {
    background: color-mix(in srgb, var(--hister-indigo) 15%, var(--brutal-card));
    color: var(--text-primary);
    box-shadow: inset 0 0 0 1px var(--hister-indigo);
  }
  .dataset-facts dl {
    margin: 0 0 18px;
    display: grid;
    gap: 7px;
    font-size: 12px;
  }
  .dataset-facts dl > div {
    display: flex;
    justify-content: space-between;
    gap: 16px;
  }
  .dataset-facts dt {
    color: var(--text-secondary);
  }
  .dataset-facts dd {
    font-variant-numeric: tabular-nums;
    margin: 0;
    text-align: right;
  }
  .dataset-download {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 44px;
    padding: 10px 14px;
    border: 1px solid var(--border);
    border-radius: 6px;
    color: var(--text-primary);
    font-size: 12px;
    font-weight: 500;
    text-align: center;
    text-decoration: none;
  }
  .dataset-download:hover {
    background: var(--muted-surface);
    border-color: var(--text-secondary);
  }
  .dataset-image {
    max-width: 100%;
    max-height: 160px;
    object-fit: cover;
    margin-bottom: 16px;
    border-radius: 4px;
  }
  .dataset-empty {
    padding: 32px 0;
    border-top: 1px solid var(--border);
  }
  .dataset-empty h2 {
    font-size: 20px;
    font-weight: 500;
  }
  .dataset-empty p {
    color: var(--text-secondary);
    font-size: 14px;
    margin: 12px 0 20px;
  }
  @media (max-width: 700px) {
    .dataset-search {
      grid-column: 1 / -1;
    }
    .dataset-row {
      grid-template-columns: minmax(0, 1fr);
      gap: 20px;
    }
    .dataset-facts {
      max-width: 340px;
    }
    .dataset-filters {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
  #submit-dataset-dialog {
    border-radius: 8px;
  }
  .submission-intro {
    margin: 0;
    color: var(--text-secondary);
    font-size: 14px;
    line-height: 1.7;
  }
  .submission-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }
  .submission-action {
    display: inline-flex;
    align-items: center;
    align-self: start;
    gap: 10px;
    min-height: 44px;
    padding: 10px 16px;
    border-radius: 4px;
    background: var(--text-primary);
    color: var(--brutal-bg);
    font-size: 13px;
    font-weight: 500;
    text-decoration: none;
  }
  .submission-action:hover {
    background: var(--text-secondary);
  }
  .submission-reference h3 {
    margin: 0 0 12px;
    font-size: 14px;
    font-weight: 600;
  }
  .submission-fields {
    display: grid;
    grid-template-columns: max-content minmax(0, 1fr);
    gap: 10px 20px;
    margin: 0;
    font-size: 13px;
    line-height: 1.6;
  }
  .submission-fields dt code {
    font-size: 12px;
  }
  .submission-fields dd {
    margin: 0;
    color: var(--text-secondary);
  }
  .submission-body {
    overflow-wrap: anywhere;
    min-width: 0;
  }
  @media (max-width: 600px) {
    .submission-fields {
      grid-template-columns: minmax(0, 1fr);
    }
    .submission-fields dt:not(:first-child) {
      margin-top: 12px;
    }
  }
</style>
