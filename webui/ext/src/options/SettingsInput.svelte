<script lang="ts">
  import { Input } from '@hister/components/ui/input';
  import { Label } from '@hister/components/ui/label';

  let {
    label,
    value = $bindable(),
    placeholder = '',
    description = '',
    type = 'text',
  }: {
    label: string;
    value: string;
    placeholder?: string;
    description?: string;
    type?: string;
  } = $props();
  const id = $props.id();
</script>

<div class="space-y-2">
  <Label for={id} class="font-outfit text-text-brand text-sm font-bold">{label}</Label>
  <Input
    {id}
    aria-describedby={description ? `${id}-description` : undefined}
    {type}
    bind:value
    {placeholder}
    class="bg-page-bg border-hister-indigo font-fira text-text-brand placeholder:text-text-brand-muted focus-visible:border-hister-coral h-12 w-full border-[3px] px-4 text-sm shadow-none transition-colors focus-visible:ring-0"
  />
  {#if description}
    <p id={`${id}-description`} class="text-text-brand-muted font-inter text-xs">{description}</p>
  {/if}
</div>
