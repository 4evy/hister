<script lang="ts">
  import Heart from '@lucide/svelte/icons/heart';
  import GitPullRequest from '@lucide/svelte/icons/git-pull-request';
  import Puzzle from '@lucide/svelte/icons/puzzle';
  import Database from '@lucide/svelte/icons/database';
  import Globe from '@lucide/svelte/icons/globe';
  import Network from '@lucide/svelte/icons/network';
  import Bitcoin from '@lucide/svelte/icons/bitcoin';
  import EthereumIcon from '#lib/EthereumIcon.svelte';
  import Copy from '@lucide/svelte/icons/copy';
  import Check from '@lucide/svelte/icons/check';
  import { onDestroy } from 'svelte';
  import Seo from '#lib/Seo.svelte';

  const cryptoWallets = [
    {
      name: 'Bitcoin',
      ticker: 'BTC',
      address: '1JwT5gEFYP358GvcdNCKbwgTKvwFujnHSF',
      href: 'bitcoin:1JwT5gEFYP358GvcdNCKbwgTKvwFujnHSF',
      icon: Bitcoin,
    },
    {
      name: 'Ethereum',
      ticker: 'ETH',
      address: '0x4Eb476D91c92b744600F282F7B727e8b46eDe8Bb',
      href: 'ethereum:0x4Eb476D91c92b744600F282F7B727e8b46eDe8Bb',
      icon: EthereumIcon,
    },
  ];

  let copyError = $state(false);
  let disposed = false;
  onDestroy(() => {
    disposed = true;
    clearTimeout(copyTimer);
  });

  let copiedAddress = $state<string | null>(null);
  let copyTimer: ReturnType<typeof setTimeout> | undefined;

  async function copyWalletAddress(address: string) {
    try {
      await navigator.clipboard.writeText(address);
      if (disposed) return;
      copyError = false;
      copiedAddress = address;
      clearTimeout(copyTimer);
      copyTimer = setTimeout(() => {
        copiedAddress = null;
      }, 1500);
    } catch {
      if (disposed) return;
      copyError = true;
      copiedAddress = null;
    }
  }

  const futurePlans = [
    {
      icon: Network,
      title: 'Federation and Distributed Search',
      desc: 'Connect trusted Hister instances so they can exchange selected public results, datasets, and index metadata while administrators remain in control of what they share and import.',
      href: null,
      linkLabel: null,
    },
    {
      icon: Puzzle,
      title: 'Content-Specific Extractors',
      desc: 'Add extractors for more content types to improve integration and index quality so Hister understands the web deeply, not just superficially.',
      href: '/docs/extractors',
      linkLabel: 'Learn about extractors',
    },
    {
      icon: Database,
      title: 'Thematic Pre-Indexed Databases',
      desc: 'Create infrastructure for importable, pre-indexed databases organized by topic, letting users quickly expand their local index with curated, relevant content.',
      href: '/datasets',
      linkLabel: 'Explore datasets',
    },
    {
      icon: Globe,
      title: 'Public Dataset Imports',
      desc: 'Support importing data from public datasets such as the Open Search Foundation and the Internet Archive, giving Hister access to the breadth of human knowledge.',
      href: null,
      linkLabel: null,
    },
  ];
</script>

<Seo
  title="Support Hister. Help Build the Future of Search"
  description="Support Hister's mission to eliminate dependence on untrusted search services by contributing code or sponsoring ongoing development."
  path="/support"
/>

<div class="support-page">
  <header class="support-intro">
    <h1>Help build the future<br />of <span>private search</span></h1>
    <p class="intro-copy">
      Whether through code, suggestions or sponsorship, every contribution helps make private,
      powerful search available to everyone.
    </p>
  </header>

  <section class="ways" aria-label="Ways to support Hister">
    <article>
      <Heart size={24} class="accent" aria-hidden="true" />
      <h2>Sponsor development</h2>
      <p>
        Financial support funds work on Hister: faster bug fixes, new features, and a more polished
        experience for everyone.
      </p>
      <a
        class="action primary"
        href="https://github.com/sponsors/asciimoo"
        target="_blank"
        rel="noopener noreferrer"
      >
        Sponsor on GitHub
      </a>
    </article>
    <article>
      <GitPullRequest size={24} class="accent" aria-hidden="true" />
      <h2>Contribute code</h2>
      <p>
        Hister is free, open-source software. Whether it's a bug report, a pull request, or
        improving documentation, every contribution moves the project forward.
      </p>
      <a
        class="action"
        href="https://github.com/asciimoo/hister"
        target="_blank"
        rel="noopener noreferrer"
      >
        View on GitHub
      </a>
    </article>
  </section>

  <section class="crypto" aria-labelledby="crypto-heading">
    <header>
      <h2 id="crypto-heading">Donate with cryptocurrency</h2>
      <p>Send BTC or ETH directly to support Hister development.</p>
    </header>
    <ul class="wallets">
      {#each cryptoWallets as wallet (wallet.ticker)}
        <li>
          <h3>
            <wallet.icon size={20} aria-hidden="true" />{wallet.name}<span>{wallet.ticker}</span>
          </h3>
          <a class="wallet-address" href={wallet.href} title="Open {wallet.ticker} wallet"
            ><code>{wallet.address}</code></a
          >
          <button
            type="button"
            class="copy-button"
            aria-label="Copy {wallet.ticker} wallet address"
            onclick={() => copyWalletAddress(wallet.address)}
          >
            {#if copiedAddress === wallet.address}<Check
                size={15}
                aria-hidden="true"
              />Copied{:else}<Copy size={15} aria-hidden="true" />Copy address{/if}
          </button>
        </li>
      {/each}
    </ul>
    <output class="copy-status" class:sr-only={!copyError}
      >{copyError
        ? 'Couldn’t copy automatically. Select the address above to copy it manually.'
        : copiedAddress
          ? 'Wallet address copied to clipboard'
          : ''}</output
    >
  </section>

  <section class="mission" aria-labelledby="mission-heading">
    <header>
      <h2 id="mission-heading">Break free from<br />untrusted search</h2>
    </header>
    <div>
      <p>
        Our goal is to <strong
          >reduce and ultimately eliminate dependence on untrusted search services</strong
        > by providing the best overall private search experience that comes with batteries included and
        supports privacy at every level.
      </p>
      <p>
        This is a fresh, independent approach to search. To make it a reality, we need your support.
      </p>
    </div>
  </section>

  <section class="roadmap" aria-labelledby="roadmap-heading">
    <header>
      <h2 id="roadmap-heading">Where your support takes Hister</h2>
      <p>These are the major pillars of the roadmap.</p>
    </header>
    <ul>
      {#each futurePlans as plan (plan.title)}
        <li>
          <plan.icon size={24} class="accent" aria-hidden="true" />
          <h3>{plan.title}</h3>
          <p>{plan.desc}</p>
          {#if plan.href}<a href={plan.href}>{plan.linkLabel}</a>{/if}
        </li>
      {/each}
    </ul>
  </section>
</div>

<style>
  .support-page {
    max-width: 1240px;
    margin: auto;
    padding: 56px 48px 64px;
    color: var(--text-primary);
  }
  .support-intro {
    max-width: 760px;
    margin-bottom: 56px;
  }
  h1 {
    font:
      500 clamp(38px, 4.5vw, 65px)/1.1 'Inter',
      sans-serif;
    letter-spacing: -0.055em;
    margin: 0;
  }
  h1 span {
    color: var(--hister-indigo);
  }
  h2 {
    font:
      500 clamp(24px, 3vw, 34px)/1.2 'Inter',
      sans-serif;
    letter-spacing: -0.04em;
    margin: 0 0 16px;
  }
  h3 {
    font-size: 17px;
    font-weight: 500;
    letter-spacing: -0.025em;
    margin: 0 0 14px;
  }
  p {
    font-size: 14px;
    line-height: 1.8;
    color: var(--text-secondary);
    margin: 0;
  }
  .intro-copy {
    max-width: 600px;
    font-size: 15px;
    margin-top: 26px;
  }
  strong {
    color: var(--text-primary);
    font-weight: 600;
  }
  a {
    text-underline-offset: 5px;
  }
  .support-page :global(.accent) {
    color: var(--hister-indigo);
  }
  .ways {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 32px;
  }
  .ways article {
    display: flex;
    flex-direction: column;
    align-items: start;
    padding: 28px;
    border-radius: 6px;
    background: var(--muted-surface);
  }
  .ways h2 {
    font-size: 23px;
    margin-top: 22px;
  }
  .ways p {
    margin-bottom: 28px;
  }
  .action {
    display: inline-flex;
    align-items: center;
    gap: 16px;
    padding: 13px 16px;
    border: 1px solid var(--border);
    text-decoration: none;
    font-size: 13px;
    font-weight: 500;
    margin-top: auto;
  }
  .action:hover {
    border-color: var(--hister-indigo);
  }
  .primary {
    background: var(--hister-indigo);
    color: #18191b;
    border-color: transparent;
  }
  .primary:hover {
    background: var(--hister-coral);
    border-color: transparent;
  }
  .crypto,
  .mission,
  .roadmap {
    border-top: 1px solid var(--border);
    margin-top: 56px;
    padding-top: 40px;
  }
  ul {
    list-style: none;
    padding: 0;
    margin: 28px 0 0;
  }
  .wallets {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 32px;
  }
  .wallets li {
    min-width: 0;
    padding: 24px;
    background: var(--brutal-card);
    border: 1px solid var(--border);
  }
  .wallets h3 {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .wallets h3 span {
    margin-left: auto;
    color: var(--text-secondary);
    font:
      11px 'Fira Code',
      monospace;
  }
  .wallet-address {
    display: block;
    overflow-wrap: anywhere;
    font:
      12px/1.9 'Fira Code',
      monospace;
  }
  .copy-button {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 20px;
    padding: 8px 0;
    cursor: pointer;
    font-size: 12px;
    color: var(--text-primary);
  }
  .copy-button:hover {
    text-decoration: underline;
    text-underline-offset: 4px;
  }
  .copy-status {
    display: block;
    margin-top: 12px;
    color: var(--text-secondary);
    font-size: 13px;
  }
  .mission {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 48px;
  }
  .mission p + p {
    margin-top: 18px;
  }
  .roadmap ul {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 36px 48px;
  }
  .roadmap li {
    padding-top: 8px;
  }
  .roadmap h3 {
    margin-top: 18px;
  }
  .roadmap a {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin-top: 20px;
    font-size: 13px;
  }
  @media (max-width: 700px) {
    .support-page {
      padding: 48px 24px;
    }
    .support-intro {
      margin-bottom: 36px;
    }
    .ways,
    .wallets,
    .mission,
    .roadmap ul {
      grid-template-columns: 1fr;
      gap: 24px;
    }
    .ways article,
    .wallets li {
      padding: 22px;
    }
    .crypto,
    .mission,
    .roadmap {
      margin-top: 40px;
      padding-top: 28px;
    }
  }
</style>
