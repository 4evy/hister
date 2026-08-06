<script lang="ts">
  import { tick } from 'svelte';
  import { ArrowUpDown, Clock3, Filter, Link2, SearchCheck, TextCursorInput } from '@lucide/svelte';
  import type { QuerySuggestion } from '$lib/query-suggestions';

  interface Props {
    activeIndex: number;
    floating?: boolean;
    id: string;
    loading?: boolean;
    open: boolean;
    suggestions: QuerySuggestion[];
    onactivechange: (index: number) => void;
    onselect: (suggestion: QuerySuggestion) => void;
  }

  let {
    activeIndex,
    floating = true,
    id,
    loading = false,
    open,
    suggestions,
    onactivechange,
    onselect,
  }: Props = $props();

  const suggestionIcons = {
    alias: Link2,
    facet: Filter,
    field: TextCursorInput,
    recent: Clock3,
    sort: ArrowUpDown,
    spelling: SearchCheck,
  };

  let listboxEl: HTMLDivElement | undefined = $state();

  $effect(() => {
    const index = activeIndex;
    if (!open || index < 0) return;
    tick().then(() => {
      const option = listboxEl?.querySelector<HTMLElement>(`[data-suggestion-index="${index}"]`);
      if (!listboxEl || !option) return;

      const optionTop = option.offsetTop;
      const optionBottom = optionTop + option.offsetHeight;
      if (optionTop < listboxEl.scrollTop) {
        listboxEl.scrollTop = optionTop;
      } else if (optionBottom > listboxEl.scrollTop + listboxEl.clientHeight) {
        listboxEl.scrollTop = optionBottom - listboxEl.clientHeight;
      }
    });
  });
</script>

{#if open && (suggestions.length > 0 || loading)}
  <div
    bind:this={listboxEl}
    {id}
    role="listbox"
    aria-label="Search suggestions"
    class="border-brutal-border bg-card-surface z-[70] overflow-y-auto py-2 {floating
      ? 'absolute top-full right-0 left-0 mt-2 max-h-[min(24rem,55vh)] border-[3px] shadow-[5px_5px_0_var(--brutal-shadow)]'
      : 'relative max-h-[min(18rem,40vh)] border-x-0 border-t-0 border-b-[3px] shadow-none'}"
  >
    {#each suggestions as suggestion, index (suggestion.id)}
      {@const Icon = suggestionIcons[suggestion.kind]}
      {#if index === 0 || suggestions[index - 1].group !== suggestion.group}
        <div
          class="font-space text-text-brand-muted px-3 pt-2 pb-1 text-[10px] font-bold tracking-[1.5px] uppercase first:pt-0"
        >
          {suggestion.group}
        </div>
      {/if}
      <button
        id={`${id}-option-${index}`}
        data-suggestion-index={index}
        type="button"
        role="option"
        aria-selected={index === activeIndex}
        class="font-inter flex w-full cursor-pointer items-center gap-3 px-3 py-2 text-left transition-colors {index ===
        activeIndex
          ? 'bg-hister-indigo/10 text-text-brand'
          : 'text-text-brand-secondary hover:bg-muted-surface hover:text-text-brand'}"
        onpointerdown={(event) => event.preventDefault()}
        onmouseenter={() => onactivechange(index)}
        onclick={() => onselect(suggestion)}
      >
        <span class="text-hister-indigo flex size-5 shrink-0 items-center justify-center">
          <Icon class="size-4" />
        </span>
        <span class="min-w-0 flex-1">
          <span class="block truncate text-sm font-semibold">{suggestion.label}</span>
          <span class="text-text-brand-muted block truncate text-xs">{suggestion.detail}</span>
        </span>
        <code
          class="font-fira bg-muted-surface text-text-brand-muted hidden max-w-52 shrink-0 truncate px-1.5 py-0.5 text-[10px] md:block"
          >{suggestion.insertText}</code
        >
      </button>
    {/each}
    {#if loading && suggestions.length === 0}
      <div class="font-inter text-text-brand-muted flex items-center gap-2 px-3 py-3 text-sm">
        <span class="bg-hister-indigo size-2 animate-pulse"></span>
        Loading values…
      </div>
    {/if}
    <div
      class="border-border-brand-muted font-inter text-text-brand-muted mt-1 flex items-center gap-3 border-t px-3 pt-2 text-[10px]"
    >
      <span>↑ ↓ choose</span>
      <span>Tab insert</span>
      <span>Esc close</span>
    </div>
  </div>
{/if}
