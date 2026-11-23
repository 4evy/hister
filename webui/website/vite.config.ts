import adapter from '@sveltejs/adapter-static';
import { markdown } from './markdown.ts';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { cpSync, mkdirSync } from 'fs';
import { resolve } from 'path';
import { defineConfig } from 'vite';

function copyDatasetJsons() {
  return {
    name: 'copy-dataset-jsons',
    apply: 'build' as const,
    closeBundle() {
      const src = resolve('src/content/datasets');
      const dest = resolve('build/datasets');
      mkdirSync(dest, { recursive: true });
      cpSync(src, dest, { recursive: true });
    },
  };
}

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit({
      extensions: ['.svelte', '.md', '.svx'],
      preprocess: [markdown],
      adapter: adapter({ pages: 'build', assets: 'build', fallback: '404.html' }),
      prerender: { handleHttpError: 'warn', handleMissingId: 'ignore' },
    }),
    copyDatasetJsons(),
  ],
  ssr: {
    noExternal: ['@nogon/components', 'bits-ui', 'svelte-toolbelt', 'runed', 'svelte-sonner'],
  },
  build: {
    rolldownOptions: {
      checks: { pluginTimings: false },
    },
  },
});
