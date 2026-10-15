<script lang="ts">
  import type { Snippet } from 'svelte';
  let {
    title,
    url,
    actions,
    children,
    heading = $bindable(),
    headingLevel = 2,
  }: {
    title: string;
    url: string;
    actions?: Snippet;
    children?: Snippet;
    heading?: HTMLHeadingElement;
    headingLevel?: 2 | 3 | 4 | 5 | 6;
  } = $props();
</script>

<header
  class="preview-header border-border-brand-muted flex shrink-0 flex-col gap-0.5 border-b-[2px] px-4 py-2.5"
>
  <div class="flex items-start gap-2">
    <svelte:element
      this={`h${headingLevel}`}
      bind:this={heading}
      tabindex="-1"
      class="font-outfit text-text-brand max-h-[min(6.875em,25dvh)] min-w-0 flex-1 overflow-y-auto text-lg leading-snug font-bold wrap-anywhere md:text-3xl"
    >
      {#if url.startsWith('remote-file://')}{title}{:else}
        <a href={url} target="_blank" rel="noopener noreferrer" class="hover:underline">{title}</a>
      {/if}
    </svelte:element>
    {#if actions}
      <div
        class="mt-1 flex shrink-0 flex-wrap items-center justify-end gap-1 text-sm leading-normal"
      >
        {@render actions()}
      </div>
    {/if}
  </div>
  {@render children?.()}
</header>
