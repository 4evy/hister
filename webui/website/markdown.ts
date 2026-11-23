// SPDX-License-Identifier: AGPL-3.0-or-later
import type { Element, Nodes, Root } from 'hast';
import type { Root as MarkdownRoot } from 'mdast';
import { toHtml } from 'hast-util-to-html';
import rehypeSlug from 'rehype-slug';
import remarkFrontmatter from 'remark-frontmatter';
import remarkGfm from 'remark-gfm';
import remarkParse from 'remark-parse';
import remarkRehype from 'remark-rehype';
import remarkSmartypants from 'remark-smartypants';
import { createHighlighter } from 'shiki';
import { parse as parseSvelte, type PreprocessorGroup } from 'svelte/compiler';
import { unified } from 'unified';
import { visit } from 'unist-util-visit';
import { VFile } from 'vfile';
import { parse as parseYaml } from 'yaml';

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

function isComponentBlock(source: string): boolean {
  if (!/^<[A-Z]/.test(source) || !source.endsWith('>')) return false;
  try {
    const { fragment } = parseSvelte(source, { modern: true });
    return fragment.nodes.every(
      (part) => part.type === 'Component' || (part.type === 'Text' && !part.data.trim()),
    );
  } catch {
    return false;
  }
}

// CommonMark preserves script imports and ordinary HTML. A component block with
// a JavaScript prop such as scopes={['server', 'client']} can also parse as prose.
function svelteComponents() {
  return (tree: MarkdownRoot, file: VFile) => {
    const source = String(file);
    visit(tree, 'paragraph', (node, index, parent) => {
      if (!parent || !node.position) return;
      const block = source.slice(node.position.start.offset, node.position.end.offset);
      if (isComponentBlock(block)) {
        // visit supplies the sibling index whenever a parent is present.
        parent.children[index!] = { type: 'html', value: block, position: node.position };
      }
    });
  };
}

function scrollableDocTables() {
  return (tree: Root, file: VFile) => {
    if (!file.path.replaceAll('\\', '/').includes('/src/content/docs/')) return;
    visit(tree, 'element', (node, index, parent) => {
      if (node.tagName !== 'table' || !parent || index === undefined) return;
      const wrapper: Element = {
        type: 'element',
        tagName: 'div',
        properties: {
          className: ['table-scroll'],
          tabIndex: 0,
          role: 'region',
          ariaLabel: 'Scrollable table',
        },
        children: [node],
      };
      // Keyboard users need to focus this region to scroll a wide table.
      parent.children.splice(
        index,
        1,
        { type: 'comment', value: ' svelte-ignore a11y_no_noninteractive_tabindex ' },
        wrapper,
      );
      return index + 2;
    });
  };
}

// Only raw nodes contain authored markup or code emitted by the build. Braces in
// Markdown text, link titles and other generated attributes are literal characters.
function svelteElement(element: Element): string {
  const tags = toHtml({ ...element, children: [] }).replaceAll('{', '&#123;');
  if (element.children.length === 0) return tags;
  const closing = tags.lastIndexOf('</');
  return tags.slice(0, closing) + element.children.map(svelteMarkup).join('') + tags.slice(closing);
}

function svelteMarkup(node: Nodes): string {
  switch (node.type) {
    case 'raw':
      return node.value;
    case 'root':
      return node.children.map(svelteMarkup).join('');
    case 'element':
      return svelteElement(node);
    default:
      return toHtml(node).replaceAll('{', '&#123;');
  }
}

const processor = unified()
  .use(remarkParse)
  .use(remarkFrontmatter)
  .use(remarkGfm, { singleTilde: false })
  .use(svelteComponents)
  .use(remarkSmartypants)
  .use(remarkRehype, {
    allowDangerousHtml: true,
    handlers: {
      yaml(_state, node) {
        const metadata = JSON.stringify(parseYaml(node.value)).replaceAll('<', '\\u003c');
        return {
          type: 'raw',
          value: `<script module>export const metadata = ${metadata};</script>`,
        };
      },
      code(_state, node) {
        const lang = node.lang ?? 'text';
        const resolved = langAliases[lang] ?? lang;
        const html = highlighter.codeToHtml(node.value, {
          lang: loadedLangs.has(resolved) ? resolved : 'text',
          theme,
        });
        return { type: 'raw', value: `{@html ${JSON.stringify(html)}}` };
      },
    },
  })
  .use(rehypeSlug)
  .use(scrollableDocTables);

export const markdown: PreprocessorGroup = {
  name: 'remark',
  async markup({ content, filename }) {
    if (!filename || !/\.(md|svx)$/.test(filename)) return;
    const file = new VFile({ value: content, path: filename });
    const tree = await processor.run(processor.parse(file), file);
    return { code: svelteMarkup(tree) };
  },
};
