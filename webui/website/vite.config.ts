import adapter from '@sveltejs/adapter-static';
import { mdsvex, escapeSvelte } from 'mdsvex';
import rehypeSlug from 'rehype-slug';
import { createCssVariablesTheme, createHighlighter } from 'shiki';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { cpSync, mkdirSync } from 'fs';
import { resolve } from 'path';
import { defineConfig } from 'vite';
import type { Element, Root } from 'hast';

function addTableAccessibilityComment(parent: Root | Element, index: number) {
  const note =
    'svelte-ignore a11y_no_noninteractive_tabindex (Scrollable tables need keyboard focus)';
  // mdsvex emits root comment values directly when processing frontmatter
  parent.children.splice(index, 0, {
    type: 'comment',
    value: parent.type === 'root' ? `<!-- ${note} -->` : ` ${note} `,
  });
}

function scrollableTables() {
  return (tree: Root) => {
    function wrapTables(parent: Root | Element) {
      for (let index = 0; index < parent.children.length; index++) {
        const node = parent.children[index];
        if (node.type !== 'element') continue;
        if (node.tagName === 'table') {
          parent.children[index] = {
            type: 'element',
            tagName: 'div',
            properties: {
              className: ['table-scroll'],
              tabIndex: 0,
              role: 'region',
              ariaLabel: 'Scrollable table',
            },
            children: [node],
          } satisfies Element;
          addTableAccessibilityComment(parent, index);
          index++;
        } else {
          wrapTables(node);
        }
      }
    }
    wrapTables(tree);
  };
}

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

const themeName = 'hister';
const theme = createCssVariablesTheme({ name: themeName });

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
          rehypePlugins: [rehypeSlug, scrollableTables],
          highlight: {
            highlighter: (code, lang) => {
              const resolved = lang ? langAliases[lang] || lang : 'text';
              const html = escapeSvelte(
                highlighter.codeToHtml(code, {
                  lang: loadedLangs.has(resolved) ? resolved : 'text',
                  theme: themeName,
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
