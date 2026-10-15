import { queryTokens, type QuerySyntax } from './query-tokens.js';

function sortDirectives(text: string, syntax: QuerySyntax) {
  const { sort } = syntax;
  const tokens = queryTokens(text, syntax);
  return tokens.flatMap((token, index) => {
    if (tokens[index - 1]?.text === 'NOT') return [];
    const prefix = `${sort.field}:`;
    if (!sort.field || !token.text.startsWith(prefix)) return [];
    const value = token.text.slice(prefix.length);
    return sort.options.some((option) => option.value === value) ? [{ ...token, value }] : [];
  });
}

export function sortDirectiveFromQuery(text: string, syntax: QuerySyntax): string | null {
  return sortDirectives(text, syntax).at(-1)?.value ?? null;
}

export function sortValueFromQuery(text: string, syntax: QuerySyntax): string {
  const directive = sortDirectiveFromQuery(text, syntax);
  if (directive === null) return '';
  return syntax.sort.options.find((option) => option.value === directive)?.default ? '' : directive;
}

export function removeSortDirectives(text: string, syntax: QuerySyntax): string {
  let remaining = text;
  for (const token of sortDirectives(text, syntax).reverse()) {
    remaining = remaining.slice(0, token.start) + remaining.slice(token.end);
  }
  return remaining.trim();
}

export function replaceSortDirective(text: string, value: string, syntax: QuerySyntax): string {
  const { sort } = syntax;
  const definition = sort.options.find((option) => option.value === value);
  if (value && !definition) return text;
  const remaining = removeSortDirectives(text, syntax);
  if (isDefaultSort(value, definition)) return remaining;
  return appendSortDirective(remaining, sort.field, value);
}

function appendSortDirective(text: string, field: string, value: string): string {
  return `${text ? `${text} ` : ''}${field}:${value}`;
}

function isDefaultSort(
  value: string,
  definition: QuerySyntax['sort']['options'][number] | undefined,
): boolean {
  return !value || !!definition?.default;
}
