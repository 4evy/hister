<script lang="ts">
  import { Button } from '../ui/button';
  import { X, Maximize2, Minimize2 } from '@lucide/svelte';
  let {
    onclose,
    fullscreen = false,
    onfullscreentoggle,
  }: {
    onclose: () => void;
    fullscreen?: boolean;
    onfullscreentoggle?: () => void;
  } = $props();
</script>

{#if onfullscreentoggle}
  <Button
    variant="ghost"
    size="icon-sm"
    class="text-text-brand-muted hover:text-text-brand"
    onclick={onfullscreentoggle}
    aria-label={fullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
    title={fullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
  >
    {#if fullscreen}<Minimize2 class="size-4" />{:else}<Maximize2 class="size-4" />{/if}
  </Button>
{/if}
<Button
  variant="ghost"
  size="icon-sm"
  class="text-text-brand-muted hover:text-text-brand"
  aria-label="Close preview"
  onclick={onclose}><X class="size-4" /></Button
>
