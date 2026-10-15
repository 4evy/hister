<script lang="ts">
  const linkCols = [
    {
      title: 'Product',
      links: [
        { label: 'Features', href: '/#features' },
        { label: 'Documentation', href: '/docs' },
        { label: 'Demo', href: 'https://demo.hister.org/' },
      ],
    },
    {
      title: 'Resources',
      links: [
        { label: 'GitHub', href: 'https://github.com/asciimoo/hister' },
        { label: 'Codeberg', href: 'https://codeberg.org/asciimoo/hister' },
        { label: 'Posts', href: '/posts' },
        { label: 'Datasets', href: '/datasets' },
      ],
    },
    {
      title: 'Extensions',
      links: [
        {
          label: 'Chrome',
          href: 'https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb',
        },
        { label: 'Firefox', href: 'https://addons.mozilla.org/en-US/firefox/addon/hister/' },
      ],
    },
    {
      title: 'Community',
      links: [
        { label: 'Mastodon', href: 'https://chaos.social/@hister' },
        { label: 'Discord', href: 'https://discord.gg/beEyuHxRSs' },
        { label: 'IRCNet: #hister', href: '' },
      ],
    },
  ];
</script>

<footer>
  <div class="footer-main">
    <div class="footer-brand">
      <a href="/">hister</a>
      <p>Self-hosted full-text search.<br />Free software under AGPLv3.</p>
    </div>
    <nav aria-label="Footer">
      {#each linkCols as col}<div>
          <h2>{col.title}</h2>
          <ul>
            {#each col.links as link}<li>
                {#if link.href}<a href={link.href}>{link.label}</a>{:else}<span>{link.label}</span
                  >{/if}
              </li>{/each}
          </ul>
        </div>{/each}
    </nav>
  </div>
  <div class="footer-bottom">
    <span>© 2026 Hister</span><a href="https://www.gnu.org/licenses/#AGPL">Free software / AGPLv3</a
    ><a href="#main-content">Back to top</a>
  </div>
</footer>

<style>
  footer {
    max-width: 1240px;
    width: 100%;
    margin: auto;
    padding: 60px 48px 25px;
  }
  :global(.home-theme) footer {
    max-width: 1240px;
    border-top: 1px solid var(--border);
  }
  a {
    text-decoration: none;
  }
  a:hover {
    text-decoration: underline;
    text-underline-offset: 4px;
  }
  .footer-main {
    display: flex;
    justify-content: space-between;
    gap: 70px;
    padding-bottom: 55px;
  }
  .footer-brand > a {
    font:
      800 42px 'Outfit',
      sans-serif;
    letter-spacing: -2px;
  }
  .footer-brand p {
    font-size: 12px;
    line-height: 1.8;
    color: var(--text-secondary);
    margin-top: 12px;
  }
  nav {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 48px;
  }
  h2 {
    font:
      11px 'Fira Code',
      monospace;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin: 10px 0 20px;
  }
  ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  li {
    font-size: 13px;
    margin-top: 8px;
    color: var(--text-secondary);
  }
  li a {
    display: inline-block;
    padding: 4px 0;
  }
  .footer-bottom {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 18px;
    border-top: 1px solid var(--border);
    padding-top: 23px;
    font:
      11px 'Fira Code',
      monospace;
    color: var(--text-secondary);
  }
  @media (max-width: 900px) {
    .footer-main {
      flex-direction: column;
      gap: 28px;
    }
    nav {
      gap: 24px;
    }
  }
  @media (max-width: 600px) {
    footer {
      padding: 48px 24px 24px;
    }
    nav {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style>
