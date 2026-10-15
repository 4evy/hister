<script lang="ts">
  import type { Snippet } from 'svelte';
  let {
    url,
    href = url,
    title,
    headingLevel = 2,
    htmlSnippet = '',
    active = false,
    color = 'var(--hister-cyan)',
    class: className = '',
    target,
    onclick,
    onauxclick,
    leading,
    actions,
    metadata,
  }: {
    url: string;
    href?: string;
    title: string;
    headingLevel?: 2 | 3 | 4 | 5 | 6;
    // Accepts the app’s sanitized result HTML or trusted static excerpts
    htmlSnippet?: string;
    active?: boolean;
    color?: string;
    class?: string;
    target?: string;
    onclick?: (event: MouseEvent) => void;
    onauxclick?: (event: MouseEvent) => void;
    leading?: Snippet;
    actions?: Snippet;
    metadata?: Snippet;
  } = $props();
</script>

<article
  data-result
  class="result-card w-full scroll-my-[6em] space-y-0.5 {className}"
  class:result-card-active={active}
  style:--result-color={color}
>
  <div class="flex min-w-0 items-start gap-1.5">
    {@render leading?.()}
    <svelte:element this={`h${headingLevel}`} class="min-w-0 flex-1">
      <a
        data-result-link={url}
        {href}
        {title}
        {target}
        {onclick}
        {onauxclick}
        class="result-title font-outfit text-md line-clamp-3 font-semibold hover:underline md:text-xl"
        >{title}</a
      >
    </svelte:element>
    {@render actions?.()}
  </div>
  {@render metadata?.()}
  {#if htmlSnippet}<p
      class="result-snippet font-inter text-text-brand-secondary line-clamp-3 text-sm leading-[1.45] md:text-base"
    >
      {@html htmlSnippet}
    </p>{/if}
</article>

<style>
  .result-card {
    position: relative;
    container-type: inline-size;
    min-width: 0;
    max-width: 100%;
    overflow: hidden;
    border: 0;
    background: transparent;
    padding: 0.2rem 0;
    box-shadow: none;
  }

  .result-card:hover {
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--result-color) 4%, transparent),
      transparent 44%
    );
  }

  .result-card :global(.result-domain) {
    color: color-mix(in srgb, var(--result-color) 70%, var(--text-primary-brand));
  }

  .result-card :global(.result-meta) {
    color: var(--text-muted-brand);
  }

  .result-card-active {
    border-inline-start: 3px solid var(--result-color);
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--result-color) 8%, transparent),
      color-mix(in srgb, var(--result-color) 2%, transparent) 62%,
      transparent
    );
    padding-inline-start: 0.75rem;
    box-shadow: none;
  }

  :global(.results-list-panel) .result-card:not(.result-card-active),
  :global(.results-list-panel) .result-card-active {
    width: calc(100% - 1.5rem);
    max-width: 100%;
    overflow: hidden;
  }

  :global(.results-list-panel) .result-card-active {
    box-shadow: none;
  }

  .result-title {
    color: color-mix(in srgb, var(--result-color) 82%, var(--text-primary-brand));
    overflow-wrap: anywhere;
  }

  .result-title:hover {
    color: var(--result-color);
  }

  .result-snippet {
    overflow-wrap: anywhere;
  }

  .result-card :global(.result-label) {
    overflow-wrap: anywhere;
    white-space: normal;
  }

  .result-card :global(.result-url-text),
  .result-card :global(.result-url-text) :global(.result-domain),
  .result-card :global(.result-url-text) :global(.result-path) {
    overflow-wrap: normal;
    white-space: nowrap;
    word-break: normal;
  }

  @container (max-width: 360px) {
    .result-card :global(.result-meta) {
      flex-wrap: wrap;
      overflow: visible;
    }

    .result-card :global(.result-url-line) {
      flex-basis: 100%;
      width: 100%;
    }

    .result-card :global(.result-secondary-meta) {
      flex-basis: 100%;
    }
  }

  .result-card :global(.result-snippet mark) {
    color: var(--text-primary-brand);
    background: color-mix(in srgb, var(--hister-amber) 18%, transparent);
    padding: 0 0.12em;
  }

  :global(.dark) .result-card {
    box-shadow: none;
  }

  :global(.dark) .result-card:hover {
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--result-color) 7%, transparent),
      transparent 46%
    );
  }

  :global(.dark) .result-title {
    color: color-mix(in srgb, var(--result-color) 88%, white);
  }

  :global(.dark) .result-card-active {
    border-inline-start-color: var(--result-color);
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--result-color) 10%, transparent),
      color-mix(in srgb, var(--result-color) 4%, transparent) 58%,
      transparent
    );
    box-shadow: none;
  }

  :global(.dark .results-list-panel .result-card-active) {
    box-shadow: none;
  }
</style>
