<script lang="ts">
  import type { Snippet } from 'svelte';
  let { children }: { children: Snippet } = $props();
</script>

<div
  class="preview-content font-inter text-text-brand-secondary prose dark:prose-invert prose-a:text-hister-teal w-full max-w-[60em] p-4 text-sm"
>
  {@render children()}
</div>
