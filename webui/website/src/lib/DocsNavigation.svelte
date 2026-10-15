<script lang="ts">
  interface DocEntry {
    slug: string;
    title: string;
  }

  interface DocCategory {
    name: string;
    color: string;
    docs: DocEntry[];
  }

  let {
    categories,
    currentSlug,
  }: {
    categories: DocCategory[];
    currentSlug: string;
  } = $props();
</script>

<div class="flex flex-col gap-4">
  <a
    href="/docs"
    class="font-inter flex items-center gap-2 px-3 py-2.5 text-sm text-(--text-secondary) no-underline transition-colors hover:bg-(--muted-surface)"
  >
    All documentation
  </a>

  <nav aria-label="Documentation" class="flex flex-col gap-5">
    {#each categories as category}
      <section>
        <h2
          class="font-inter mb-1.5 px-3 text-xs font-semibold {category.docs.some(
            (doc) => doc.slug === currentSlug,
          )
            ? 'text-(--text-primary)'
            : 'text-(--text-secondary)'}"
        >
          {category.name}
        </h2>

        <ul class="flex flex-col gap-0.5">
          {#each category.docs as doc}
            <li>
              <a
                href="/docs/{doc.slug}"
                aria-current={doc.slug === currentSlug ? 'page' : undefined}
                class="font-inter flex items-center justify-between gap-2 border-l-2 px-3 py-2 text-sm no-underline transition-colors {doc.slug ===
                currentSlug
                  ? 'border-hister-indigo bg-hister-indigo/10 font-semibold text-(--text-primary)'
                  : 'border-transparent text-(--text-secondary) hover:border-(--border) hover:bg-(--muted-surface) hover:text-(--text-primary)'}"
              >
                <span>{doc.title}</span>
              </a>
            </li>
          {/each}
        </ul>
      </section>
    {/each}
  </nav>
</div>
