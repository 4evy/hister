<script lang="ts">
  import { onDestroy } from 'svelte';
  import { Check, Copy } from '@lucide/svelte';

  let { command, label }: { command: string; label: string } = $props();
  let copied = $state(false);
  let failed = $state(false);
  let code: HTMLElement;
  let reset: ReturnType<typeof setTimeout> | undefined;
  let disposed = false;

  onDestroy(() => {
    disposed = true;
    clearTimeout(reset);
  });

  async function copy() {
    clearTimeout(reset);
    try {
      await navigator.clipboard.writeText(command);
      if (disposed) return;
      copied = true;
      failed = false;
      reset = setTimeout(() => (copied = false), 2500);
    } catch {
      if (disposed) return;
      copied = false;
      failed = true;
      window.getSelection()?.selectAllChildren(code);
    }
  }
</script>

<button class="command" class:copied onclick={copy} aria-label={label}>
  <span class="prompt" aria-hidden="true">$</span>
  <code bind:this={code}>{command}</code>
  <span class="copy-feedback" aria-hidden="true">
    {#if copied}<Check size={16} />{:else}<Copy size={16} />{/if}
    <span>{copied ? 'Copied' : 'Copy'}</span>
  </span>
</button>
<output class="sr-only">{copied ? 'Command copied to clipboard' : ''}</output>
{#if failed}<p class="copy-error" role="status">
    Couldn’t copy automatically. The command is selected so you can copy it manually.
  </p>{/if}

<style>
  .command {
    display: flex;
    align-items: start;
    gap: 12px;
    width: 100%;
    padding: 16px;
    border: 1px solid var(--border);
    background: var(--brutal-card);
    text-align: left;
    cursor: pointer;
  }
  .command:hover,
  .command.copied {
    border-color: var(--hister-indigo);
    background: color-mix(in srgb, var(--hister-indigo) 8%, var(--brutal-card));
  }
  .prompt {
    color: var(--hister-coral);
    font:
      13px/1.8 'Fira Code',
      monospace;
  }
  code {
    flex: 1;
    min-width: 0;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
    user-select: text;
    color: var(--text-primary);
    font:
      12px/1.8 'Fira Code',
      monospace;
  }
  .copy-feedback {
    display: flex;
    align-items: center;
    justify-content: end;
    gap: 6px;
    min-width: 63px;
    padding-top: 3px;
    color: var(--text-secondary);
    font:
      11px/1.5 'Inter',
      sans-serif;
  }
  .copied .copy-feedback {
    color: var(--text-primary);
  }
  .copy-error {
    margin: 8px 0 0;
    font-size: 12px;
    color: var(--text-secondary);
  }
  @media (max-width: 500px) {
    .command {
      padding: 14px 12px;
      gap: 8px;
    }
    .copy-feedback {
      min-width: 16px;
    }
    .copy-feedback span {
      display: none;
    }
  }
</style>
