<script lang="ts">
  import { Button, ThemeOptions } from '@hister/components';
  import * as DropdownMenu from '@hister/components/ui/dropdown-menu';
  import { SunMoon } from '@lucide/svelte';
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button {...props} variant="ghost" size="icon-sm" aria-label="Color theme"
        ><SunMoon size={18} /></Button
      >
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content
    align="end"
    class="border-brutal-border bg-card-surface w-44 rounded-none border-[3px] p-2 shadow-[4px_4px_0_var(--brutal-shadow)]"
  >
    <ThemeOptions />
  </DropdownMenu.Content>
</DropdownMenu.Root>
