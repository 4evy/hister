import { contentSlug } from '#lib/content.js';
import { error } from '@sveltejs/kit';
import type { Component } from 'svelte';

export const prerender = true;

const modules = import.meta.glob('../../../content/posts/*.md', { eager: true });

const posts = Object.fromEntries(
  Object.entries(modules).map(([path, mod]) => {
    const slug = contentSlug(path);
    return [slug, mod];
  }),
) as Record<string, { default: Component; metadata?: Record<string, unknown> }>;

export function entries() {
  return Object.keys(posts).map((slug) => ({ slug }));
}

function formattedDate(value?: string): string {
  if (!value) return '';
  return new Date(value).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

function postDetails(metadata: Record<string, unknown>) {
  const publishedTime = metadata.date as string | undefined;
  return {
    date: formattedDate(publishedTime),
    publishedTime,
    modifiedTime: (metadata.updated as string | undefined) ?? publishedTime,
    author: (metadata.author as string | undefined) ?? 'Adam Tauber',
  };
}

export async function load({ params }: { params: { slug: string } }) {
  const post = posts[params.slug];
  if (!post) {
    error(404, `Post "${params.slug}" not found`);
  }
  const metadata = post.metadata ?? {};
  return {
    content: post.default,
    meta: post.metadata ?? {},
    slug: params.slug,
    ...postDetails(metadata),
  };
}
