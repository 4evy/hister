import adapter from '@sveltejs/adapter-static';
import { mdsvex, escapeSvelte } from 'mdsvex';
import rehypeSlug from 'rehype-slug';
import { createHighlighter } from 'shiki';
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

const theme = 'github-dark';

const langs = [
  'bash',
  'shell',
  'yaml',
  'json',
  'javascript',
  'typescript',
  'html',
  'css',
  'nginx',
  'nix',
  'go',
  'text',
  'plaintext',
  'markdown',
  'dockerfile',
];

const langAliases: Record<string, string> = { textplain: 'text', caddy: 'text' };
const highlighter = await createHighlighter({ themes: [theme], langs });
const loadedLangs = new Set(highlighter.getLoadedLanguages());

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit({
      extensions: ['.svelte', '.md', '.svx'],
      preprocess: [
        mdsvex({
          extensions: ['.md', '.svx'],
          rehypePlugins: [rehypeSlug],
          highlight: {
            highlighter: (code, lang) => {
              const resolved = lang ? langAliases[lang] || lang : 'text';
              const html = escapeSvelte(
                highlighter.codeToHtml(code, {
                  lang: loadedLangs.has(resolved) ? resolved : 'text',
                  theme,
                }),
              );

              return `{@html \`${html}\` }`;
            },
          },
        }),
      ],
      adapter: adapter({ pages: 'build', assets: 'build', fallback: '404.html' }),
      prerender: { handleHttpError: 'warn', handleMissingId: 'ignore' },
    }),
    copyDatasetJsons(),
  ],
  ssr: {
    noExternal: ['@hister/components', 'bits-ui', 'svelte-toolbelt', 'runed', 'svelte-sonner'],
  },
  build: {
    rolldownOptions: {
      checks: { pluginTimings: false },
    },
  },
});
