<script lang="ts">
  type ConfigScope = 'client' | 'server';

  interface ConfigReferenceItem {
    name: string;
    type: string;
    defaultValue?: string;
    requirement?: string;
    description: string;
    scopes?: ConfigScope[];
  }

  const referenceId = $props.id();

  let { items, scopes = [] }: { items: ConfigReferenceItem[]; scopes?: ConfigScope[] } = $props();
</script>

<div class="config-options">
  {#each items as item (item.name)}
    {@const optionScopes = item.scopes ?? scopes}
    <section aria-labelledby="{referenceId}-{item.name}" class="config-option">
      <header>
        <h3 id="{referenceId}-{item.name}"><code>{item.name}</code></h3>
        <span class="config-scope"
          >{optionScopes.length ? optionScopes.join(' · ') : 'Config option'}</span
        >
      </header>
      <p>{item.description}</p>
      <dl>
        <div>
          <dt>Type</dt>
          <dd><code>{item.type}</code></dd>
        </div>
        {#if item.defaultValue !== undefined}<div>
            <dt>Default</dt>
            <dd><code>{item.defaultValue}</code></dd>
          </div>{/if}
        {#if item.requirement}<div>
            <dt>Requirement</dt>
            <dd>{item.requirement}</dd>
          </div>{/if}
      </dl>
    </section>
  {/each}
</div>

<style>
  .config-options {
    display: grid;
    gap: 0;
    margin-bottom: 28px;
  }
  .config-option {
    min-width: 0;
    border-top: 1px solid var(--border);
    padding: 24px 0;
  }
  header {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    justify-content: space-between;
    gap: 12px;
  }
  .config-option h3 {
    margin: 0;
    font-size: 18px;
  }
  .config-option h3 code {
    font-size: inherit;
    border: 0;
    padding: 0;
    background: transparent;
  }
  .config-scope,
  dt {
    font:
      12px 'Fira Code',
      monospace;
    color: var(--text-secondary);
  }
  .config-option p {
    margin: 12px 0 16px;
    font-size: 16px;
  }
  dl {
    display: flex;
    flex-wrap: wrap;
    margin: 0;
    gap: 16px 24px;
  }
  dl div {
    min-width: 0;
    flex: 1 1 160px;
  }
  dd {
    margin: 6px 0 0;
    font-size: 14px;
    overflow-wrap: anywhere;
  }
  dd code {
    border: 0;
    padding: 0;
    background: transparent;
  }
</style>
