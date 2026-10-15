<script lang="ts">
  import ImageLightbox from '#lib/ImageLightbox.svelte';
  import Seo from '#lib/Seo.svelte';

  let { data } = $props();
</script>

<Seo
  title={`${data.meta.title} | Hister`}
  description={(data.meta.description as string) ?? `Read ${data.meta.title} on the Hister blog.`}
  path={`/posts/${data.slug}`}
  type="article"
  image={data.meta.image ? `/posts/${data.meta.image as string}` : undefined}
  imageAlt={data.meta.image ? `${data.meta.title} social image` : undefined}
  publishedTime={data.publishedTime}
  modifiedTime={data.modifiedTime}
  author={data.author}
  section="Post"
  schemaType="BlogPosting"
  schemaTitle={data.meta.title as string}
  breadcrumbs={[
    { name: 'Home', path: '/' },
    { name: 'Posts', path: '/posts' },
    { name: data.meta.title as string, path: `/posts/${data.slug}` },
  ]}
/>

<article class="post-article">
  <header class="reading-header">
    <nav aria-label="Breadcrumb" class="reading-breadcrumb">
      <a href="/posts">All posts</a><span aria-hidden="true">/</span><span>Hister journal</span>
    </nav>
    {#if data.date}<time datetime={data.publishedTime}>{data.date}</time>{/if}
    <h1>{data.meta.title}</h1>
    {#if data.meta.description}<p>{data.meta.description}</p>{/if}
  </header>
  <div class="content"><data.content /></div>
  <footer class="article-footer"><a href="/posts">Back to all posts</a></footer>
</article>

<ImageLightbox contentKey={data.slug} />
