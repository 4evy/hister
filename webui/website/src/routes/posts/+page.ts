import { contentSlug, contentMetadata } from '#lib/content.js';
const modules = import.meta.glob('../../content/posts/*.md', { eager: true });

export async function load() {
  const posts = Object.entries(modules)
    .map(([path, mod]) => {
      const slug = contentSlug(path);
      const metadata = contentMetadata<{ date: string; title: string; description: string }>(mod);
      return {
        slug,
        title: metadata.title ?? slug,
        date: metadata.date,
        description: metadata.description,
      };
    })
    .sort((a, b) => new Date(b.date ?? 0).getTime() - new Date(a.date ?? 0).getTime());

  return { posts };
}
