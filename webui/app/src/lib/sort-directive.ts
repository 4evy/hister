export type SortDirective =
  'relevance' | 'visits' | 'date' | 'domain' | '-relevance' | '-visits' | '-date' | '-domain';

const sortDirectivePattern = /(?:^|\s)sort:(-?(?:relevance|visits|date|domain))(?=\s|$)/g;
const sortDirectiveValues = new Set<string>([
  'relevance',
  'visits',
  'date',
  'domain',
  '-relevance',
  '-visits',
  '-date',
  '-domain',
]);

function isInsideQuotedText(text: string, offset: number): boolean {
  return [...text.slice(0, offset).matchAll(/(?<!\\)(?:\\\\)*"/g)].length % 2 === 1;
}

export function sortDirectiveFromQuery(text: string): SortDirective | null {
  const matches = [...text.matchAll(sortDirectivePattern)].filter((match) => {
    const tokenOffset = (match.index ?? 0) + match[0].indexOf('sort:');
    return !isInsideQuotedText(text, tokenOffset);
  });
  return (matches.at(-1)?.[1] as SortDirective | undefined) ?? null;
}

export function sortValueFromQuery(text: string): string {
  const directive = sortDirectiveFromQuery(text);
  return directive === null || directive === 'relevance' ? '' : directive;
}

export function removeSortDirectives(text: string): string {
  return text
    .replace(sortDirectivePattern, (match, _sort, offset) => {
      const tokenOffset = offset + match.indexOf('sort:');
      return isInsideQuotedText(text, tokenOffset) ? match : '';
    })
    .trim();
}

export function replaceSortDirective(text: string, sort: string): string {
  if (sort && !sortDirectiveValues.has(sort)) return text;
  const remaining = removeSortDirectives(text);
  if (!sort || sort === 'relevance') return remaining;
  return `${remaining ? `${remaining} ` : ''}sort:${sort}`;
}
