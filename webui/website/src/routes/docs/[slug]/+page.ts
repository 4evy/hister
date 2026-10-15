import { contentSlug } from '#lib/content.js';
import { error } from '@sveltejs/kit';
import type { Component } from 'svelte';
import type { DocEntry } from '../+layout.js';

export const prerender = true;

const modules = import.meta.glob('../../../content/docs/*.md', { eager: true });

const docsMap = Object.fromEntries(
  Object.entries(modules).map(([path, mod]) => {
    const slug = contentSlug(path);
    return [slug, mod];
  }),
) as Record<string, { default: Component; metadata?: Record<string, unknown> }>;

export const entries = () => Object.keys(docsMap).map((slug) => ({ slug }));

function adjacentDocs(docs: DocEntry[], slug: string) {
  const currentIndex = docs.findIndex((d) => d.slug === slug);
  const prev = currentIndex > 0 ? docs[currentIndex - 1] : null;
  const next = currentIndex < docs.length - 1 ? docs[currentIndex + 1] : null;

  return { prev, next };
}

export async function load({
  params,
  parent,
}: {
  params: { slug: string };
  parent: () => Promise<{ docs: DocEntry[] }>;
}) {
  const post = docsMap[params.slug];
  if (!post) {
    error(404, `Documentation page "${params.slug}" not found`);
  }

  const { docs } = await parent();
  const navigation = adjacentDocs(docs, params.slug);

  return {
    content: post.default,
    meta: post.metadata ?? {},
    slug: params.slug,
    ...navigation,
  };
}
