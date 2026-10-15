<script lang="ts">
  import Search from '@lucide/svelte/icons/search';
  import type { HTMLInputAttributes } from 'svelte/elements';

  let {
    label,
    hint,
    shortcuts = false,
    value = $bindable(''),
    input = $bindable(),
    ...attributes
  }: HTMLInputAttributes & {
    label: string;
    hint?: string;
    shortcuts?: boolean;
    value?: string;
    input?: HTMLInputElement;
  } = $props();
  const id = $props.id();

  function isEditing(target: EventTarget | null): boolean {
    return (
      target instanceof HTMLElement &&
      (target.matches('input, textarea, select') || target.isContentEditable)
    );
  }

  function isCommandShortcut(event: KeyboardEvent): boolean {
    return (event.metaKey || event.ctrlKey) && !event.altKey && event.key.toLowerCase() === 'k';
  }

  function hasModifiers(event: KeyboardEvent): boolean {
    return event.metaKey || event.ctrlKey || event.altKey;
  }

  function isSlashShortcut(event: KeyboardEvent): boolean {
    return event.key === '/' && !hasModifiers(event) && !isEditing(event.target);
  }

  function ignoreShortcut(event: KeyboardEvent): boolean {
    return !shortcuts || event.defaultPrevented || event.isComposing;
  }

  function handleShortcut(event: KeyboardEvent) {
    if (ignoreShortcut(event)) return;
    if (isCommandShortcut(event) || isSlashShortcut(event)) {
      event.preventDefault();
      input?.focus();
    }
  }
</script>

<svelte:window onkeydown={handleShortcut} />

<div class="search-control">
  <div class="search-caption">
    <label for={id}>{label}</label>
    {#if hint || shortcuts}<span>{hint ?? 'Cmd/Ctrl K or /'}</span>{/if}
  </div>
  <div class="search-field">
    <Search size={22} aria-hidden="true" />
    <input
      {...attributes}
      {id}
      type="search"
      aria-keyshortcuts={shortcuts ? 'Meta+K Control+K /' : undefined}
      bind:this={input}
      bind:value
    />
  </div>
</div>

<style>
  .search-control {
    width: 100%;
  }
  .search-caption {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 10px;
    color: var(--text-secondary);
  }
  label {
    font-size: 11px;
    font-weight: 500;
    letter-spacing: 1.4px;
    text-transform: uppercase;
  }
  .search-caption span {
    font:
      9px 'Fira Code',
      monospace;
  }
  .search-field {
    display: flex;
    align-items: center;
    gap: 14px;
    padding-left: 16px;
    color: var(--text-secondary);
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--brutal-card);
  }
  .search-field:focus-within {
    outline: 2px solid var(--hister-indigo);
    outline-offset: 2px;
  }
  .search-field :global(svg) {
    flex-shrink: 0;
  }
  input {
    width: 100%;
    min-width: 0;
    padding: 14px 16px 14px 0;
    color: var(--text-primary);
    background: transparent;
    font:
      500 14px/1.5 'Inter',
      sans-serif;
    outline: none;
  }
  input::placeholder {
    color: var(--text-secondary);
  }
  @media (max-width: 600px) {
    .search-caption span {
      display: none;
    }
  }
</style>
