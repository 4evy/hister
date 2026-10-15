<script lang="ts">
  import Hero from '#lib/Hero.svelte';
  import HowItWorks from '#lib/HowItWorks.svelte';
  import Previews from '#lib/Previews.svelte';
  import Features from '#lib/Features.svelte';
  import Seo from '#lib/Seo.svelte';
</script>

<Seo
  title="Hister — Your own search engine"
  description="Search the full text of your browsing history and local files with Hister, a self-hosted search engine. Available on the web, in your terminal, and through an API."
  path="/"
  schemaType="SoftwareApplication"
  schemaTitle="Hister"
/>

<Hero />
<Previews />
<Features />
<HowItWorks />
