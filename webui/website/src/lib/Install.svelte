<script lang="ts">
  import { Download } from '@lucide/svelte';
  import CopyCommand from '#lib/CopyCommand.svelte';

  let selected = $state('brew');
  const methodGroup = $props.id();
  const methods = [
    {
      id: 'brew',
      label: 'Homebrew',
      title: 'Install with Homebrew',
      requirement: 'macOS or Linux · Homebrew required',
      command: 'brew install hister',
      next: 'Then start the server',
      start: 'hister listen',
      href: '/docs/installing#homebrew',
      link: 'Homebrew setup',
    },
    {
      id: 'download',
      label: 'Download',
      title: 'Prebuilt binary',
      requirement: 'No build tools needed',
      command: '',
      next: 'Choose your system and processor on the release page, then follow the binary setup steps.',
      start: '',
      href: '/docs/installing#prebuilt-binary',
      link: 'Binary setup instructions',
    },
    {
      id: 'docker',
      label: 'Docker',
      title: 'Pull the container image',
      requirement: 'Docker required',
      command: 'docker pull ghcr.io/asciimoo/hister:latest',
      next: 'Next, configure persistent storage and start the server with Docker Compose.',
      start: '',
      href: '/docs/docker',
      link: 'Docker Compose setup',
    },
    {
      id: 'nix',
      label: 'Nix',
      title: 'Install with Nix',
      requirement: 'Nix with flakes enabled',
      command: 'nix profile install github:asciimoo/hister',
      next: 'Then start the server',
      start: 'hister listen',
      href: '/docs/installing#nix',
      link: 'Nix & service configuration',
    },
    {
      id: 'source',
      label: 'Source',
      title: 'Build from source',
      requirement: 'Go 1.26, npm, and a C compiler required',
      command: 'git clone https://github.com/asciimoo/hister.git\ncd hister\n./manage.sh build',
      next: 'Then start the server',
      start: './hister listen',
      href: '/docs/installing#building-from-source',
      link: 'Build instructions',
    },
  ];
  const method = $derived(methods.find((item) => item.id === selected)!);
</script>

<section class="installer" id="install" aria-labelledby={`${methodGroup}-heading`}>
  <fieldset class="methods">
    <legend class="sr-only">Installation method</legend>
    {#each methods as item}
      <label>
        <input
          class="sr-only"
          type="radio"
          name={methodGroup}
          value={item.id}
          bind:group={selected}
        />
        {item.label}
      </label>
    {/each}
  </fieldset>
  <div class="install-body">
    <header class="method-heading">
      <h2 id={`${methodGroup}-heading`}>{method.title}</h2>
      <p>{method.requirement}</p>
    </header>
    {#key selected}
      {#if method.command}
        <CopyCommand command={method.command} label="Copy installation command" />
      {:else}
        <a class="download" href="https://github.com/asciimoo/hister/releases/latest">
          <Download size={18} aria-hidden="true" />
          <span>Download latest release</span>
          <span class="destination">GitHub</span>
        </a>
      {/if}
      <p class="next-step">{method.next}</p>
      {#if method.start}<CopyCommand command={method.start} label="Copy server command" />{/if}
    {/key}
    <a class="guide" href={method.href}>{method.link}</a>
  </div>
</section>

<style>
  .installer {
    max-width: 620px;
    margin-top: 30px;
    background: var(--brutal-bg);
    border: 1px solid var(--border);
  }
  .methods {
    display: flex;
    min-width: 0;
    margin: 0;
    padding: 0;
    border: 0;
    overflow-x: auto;
    border-bottom: 1px solid var(--border);
  }
  .methods label {
    flex: 1;
    text-align: center;
    min-height: 44px;
    padding: 12px;
    color: var(--text-secondary);
    border-bottom: 2px solid transparent;
    font:
      12px 'Inter',
      sans-serif;
    cursor: pointer;
  }
  .methods label:hover {
    color: var(--text-primary);
    background: var(--brutal-card);
  }
  .methods label:has(input:checked) {
    color: var(--text-primary);
    border-color: var(--hister-coral);
    background: var(--brutal-card);
    font-weight: 600;
  }
  .methods label:has(input:focus-visible) {
    outline: 3px solid var(--hister-indigo);
    outline-offset: -4px;
  }
  .install-body {
    padding: 20px;
  }
  .method-heading {
    margin-bottom: 16px;
  }
  h2 {
    margin: 0 0 5px;
    font-size: 14px;
    font-weight: 600;
  }
  p {
    margin: 0;
    font-size: 12px;
    line-height: 1.7;
    color: var(--text-secondary);
  }
  .next-step {
    margin: 16px 0 10px;
  }
  .guide {
    display: inline-block;
    margin-top: 16px;
    font-size: 12px;
    text-underline-offset: 4px;
  }
  .download {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 54px;
    padding: 14px 16px;
    border: 1px solid var(--border);
    background: var(--brutal-card);
    font-size: 13px;
    font-weight: 500;
    text-decoration: none;
  }
  .download:hover {
    border-color: var(--hister-indigo);
  }
  .destination {
    margin-left: auto;
    font-size: 11px;
    color: var(--text-secondary);
  }
  @media (max-width: 500px) {
    .methods label {
      padding: 12px 8px;
      font-size: 11px;
    }
    .install-body {
      padding: 16px;
    }
    .download {
      padding: 14px 12px;
    }
  }
</style>
