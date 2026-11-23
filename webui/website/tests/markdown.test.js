// SPDX-License-Identifier: AGPL-3.0-or-later
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import { join, relative, resolve } from 'node:path';
import { after, test } from 'node:test';
import { compile, preprocess } from 'svelte/compiler';
import { createServer, normalizePath, transformWithOxc } from 'vite';

// Starting the dev server also generates the config and would hide an incomplete install.
await test('the installed website can transform its docs layout with Vite', async () => {
  const layout = resolve('src/routes/docs/+layout.ts');
  await assert.doesNotReject(transformWithOxc(readFileSync(layout, 'utf8'), layout));
});

const server = await createServer({
  logLevel: 'silent',
  server: { middlewareMode: true, ws: false, watch: null },
});
after(() => server.close());
const preprocessors = server.config.plugins.find(
  (plugin) => plugin.name === 'vite-plugin-svelte:config',
).api.options.preprocess;
const filename = resolve('src/content/docs/import.md');
const table = '| Source | Details |\n| --- | --- |\n| Browser | Saved pages |\n';

test('Markdown imports Svelte components while prose and code examples stay literal', async (t) => {
  const dir = await mkdtemp(resolve('tests/document-'));
  t.after(() => rm(dir, { recursive: true }));
  await writeFile(
    join(dir, 'Names.svelte'),
    '<script>let { names } = $props();</script><button>{names.join(", ")}</button>',
  );
  const page = join(dir, 'sample.md');
  await writeFile(
    page,
    [
      '---',
      "title: 'Example'",
      '---',
      '<script>import Names from "./Names.svelte";</script>',
      '',
      '<Names names={["Ada", "Jo"]} />',
      '<Names names={["Bea"]} />',
      '',
      'A literal {placeholder} and [a link](https://example.org "look at {placeholder}").',
      '',
      '```unknown',
      '<Names names={values} />',
      'nogon index \\',
      '  https://example.org',
      '```',
    ].join('\n'),
  );
  const { render } = await server.ssrLoadModule('svelte/server');
  const doc = await server.ssrLoadModule('/' + normalizePath(relative(server.config.root, page)));
  assert.equal(doc.metadata.title, 'Example');
  const { body } = render(doc.default);
  assert.match(body, /<button>Ada, Jo<\/button>/);
  assert.match(body, /<button>Bea<\/button>/);
  assert.match(body, /A literal \{placeholder\}/);
  assert.match(body, /title="look at \{placeholder\}"/);
  assert.match(body, /(?:&lt;|&#x3C;)Names names=\{values\}/i);
  assert.match(body, /nogon index \\/);
});

test('standard Markdown preserves intraword bold, italics, tildes and reference links', async () => {
  const source = '**GO**od *news* for **~50%** of [`server` readers].\n\n[`server` readers]: /docs';
  const { code } = await preprocess(source, preprocessors, { filename });
  assert.match(
    code,
    /<strong>GO<\/strong>od <em>news<\/em> for <strong>~50%<\/strong> of <a href="\/docs"><code>server<\/code> readers<\/a>\./,
  );
});

for (const [name, markdown] of [
  ['a table', table],
  [
    'a table inside a blockquote',
    table
      .split('\n')
      .map((line) => `> ${line}`)
      .join('\n'),
  ],
  ['the import guide', readFileSync(filename, 'utf8')],
]) {
  test(`${name} compiles without noninteractive tabindex warnings`, async () => {
    const { code } = await preprocess(markdown, preprocessors, { filename });
    const { warnings } = compile(code, { filename });
    assert.deepEqual(
      warnings
        .filter((warning) => warning.code === 'a11y_no_noninteractive_tabindex')
        .map(({ code, message }) => ({ code, message })),
      [],
    );
    assert.match(
      code,
      /<\w+\b(?=[^>]*\btabindex="0")(?=[^>]*\brole="region")(?=[^>]*\baria-label="Scrollable table")[^>]*>/,
      'Tables must remain named, keyboard focusable scroll regions',
    );
  });
}

test('the table exception does not hide unrelated accessibility warnings', async () => {
  const source = `<div tabindex="0">Before the table</div>\n\n${table}\n\n<div tabindex="0">After the table</div>`;
  const { code } = await preprocess(source, preprocessors, { filename });
  const { warnings } = compile(code, { filename });
  assert.equal(
    warnings.filter((warning) => warning.code === 'a11y_no_noninteractive_tabindex').length,
    2,
  );
});
