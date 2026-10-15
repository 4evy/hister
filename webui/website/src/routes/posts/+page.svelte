<script lang="ts">
  import Rss from '@lucide/svelte/icons/rss';
  import Seo from '#lib/Seo.svelte';

  let { data } = $props();
  const months = $derived([
    ...new Set(data.posts.map((post) => post.date?.slice(0, 7) ?? 'Undated')),
  ]);
</script>

<Seo
  title="Posts | Hister"
  description="Read Hister news, guides, technical articles, and practical stories about private search and personal knowledge management."
  path="/posts"
/>

<section class="reading-index posts-index">
  <header class="reading-header">
    <h1>Posts</h1>
    <p>News, ideas, and practical guides for making the web your own.</p>
    <a href="/posts/rss.xml" data-sveltekit-reload class="feed-link"
      ><Rss size={16} aria-hidden="true" /> Subscribe via RSS</a
    >
  </header>
  {#each months as month}
    <section class="post-month" aria-labelledby={`posts-${month}`}>
      <h2 id={`posts-${month}`}>
        {month === 'Undated'
          ? month
          : new Date(`${month}-01T00:00:00Z`).toLocaleDateString('en-US', {
              month: 'long',
              year: 'numeric',
              timeZone: 'UTC',
            })}
      </h2>
      <ul class="post-list">
        {#each data.posts.filter((post) => (post.date?.slice(0, 7) ?? 'Undated') === month) as post}
          <li>
            <a href="/posts/{post.slug}">
              {#if post.date}<time datetime={post.date}
                  >{new Date(post.date).toLocaleDateString('en-US', {
                    month: 'long',
                    day: 'numeric',
                    year: 'numeric',
                    timeZone: 'UTC',
                  })}</time
                >{/if}
              <h3>{post.title}</h3>
              {#if post.description}<p>{post.description}</p>{/if}
            </a>
          </li>
        {/each}
      </ul>
    </section>
  {/each}
</section>

<style>
  .post-month {
    margin-top: 40px;
  }
  .post-month > h2 {
    font:
      13px 'Fira Code',
      monospace;
    color: var(--text-secondary);
    margin-bottom: 8px;
  }
</style>
