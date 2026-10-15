import { contentMetadata } from '#lib/content.js';
import { docsStructure } from '#lib/docs-structure.js';

const modules = import.meta.glob('../../content/docs/*.md', { eager: true });

export interface DocEntry {
  slug: string;
  title: string;
  description: string;
}

export interface DocCategory {
  name: string;
  color: string;
  docs: DocEntry[];
}

export async function load() {
  const docs: DocEntry[] = docsStructure.flatMap(({ slugs }) =>
    slugs.map((slug) => {
      const metadata = contentMetadata<Record<string, string>>(
        modules[`../../content/docs/${slug}.md`],
      );
      return {
        slug,
        title: metadata.title ?? slug.replace(/-/g, ' ').replace(/\b\w/g, (l) => l.toUpperCase()),
        description: metadata.description ?? '',
      };
    }),
  );

  const categories: DocCategory[] = docsStructure
    .map(({ slugs, ...props }) => ({
      docs: docs.filter((d) => slugs.includes(d.slug)),
      ...props,
    }))
    .filter((c) => c.docs.length > 0);

  return { docs, categories };
}
