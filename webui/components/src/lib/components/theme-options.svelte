<script lang="ts">
  import * as DropdownMenu from './ui/dropdown-menu';
  import { Monitor, Moon, Sun } from '@lucide/svelte';
  import { userPrefersMode, setMode } from 'mode-watcher';
  let { onchange = setMode }: { onchange?: (value: 'system' | 'light' | 'dark') => void } =
    $props();
  const itemClass =
    'font-space text-text-brand-muted focus:bg-muted-surface focus:text-text-brand cursor-pointer rounded-none py-2 text-xs font-semibold tracking-wider uppercase';
</script>

<DropdownMenu.RadioGroup
  aria-label="Appearance"
  value={userPrefersMode.current}
  onValueChange={(value) => onchange(value as 'system' | 'light' | 'dark')}
>
  <DropdownMenu.RadioItem class={itemClass} value="system"
    ><Monitor class="size-4" />Automatic</DropdownMenu.RadioItem
  >
  <DropdownMenu.RadioItem class={itemClass} value="light"
    ><Sun class="size-4" />Light</DropdownMenu.RadioItem
  >
  <DropdownMenu.RadioItem class={itemClass} value="dark"
    ><Moon class="size-4" />Dark</DropdownMenu.RadioItem
  >
</DropdownMenu.RadioGroup>
