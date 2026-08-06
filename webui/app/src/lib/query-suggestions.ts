import type { FacetsResult, TermCount } from '$lib/search';

export type QuerySuggestionKind = 'alias' | 'facet' | 'field' | 'recent' | 'sort' | 'spelling';

export interface QuerySuggestion {
  id: string;
  group: string;
  kind: QuerySuggestionKind;
  label: string;
  detail: string;
  insertText: string;
  replacement: 'query' | 'token';
  appendSpace?: boolean;
  keepOpen?: boolean;
}

interface QueryToken {
  start: number;
  end: number;
  value: string;
}

interface QuerySuggestionOptions {
  query: string;
  cursor: number;
  aliases: Record<string, string>;
  recentSearches: string[];
  facets?: FacetsResult | null;
  serverSuggestion?: string;
  limit?: number;
}

export interface FacetSuggestionContext {
  baseQuery: string;
  facetName: string;
  key: string;
}

interface SearchField {
  name: string;
  label: string;
  detail: string;
}

const SEARCH_FIELDS: SearchField[] = [
  { name: 'domain', label: 'Domain', detail: 'Limit results to one domain' },
  { name: 'title', label: 'Title', detail: 'Search only page titles' },
  { name: 'url', label: 'URL', detail: 'Search only page addresses' },
  { name: 'text', label: 'Text', detail: 'Search only document content' },
  { name: 'type', label: 'Type', detail: 'Filter web or local documents' },
  { name: 'language', label: 'Language', detail: 'Filter by document language' },
  { name: 'label', label: 'Label', detail: 'Search assigned labels' },
  { name: 'visits', label: 'Visits', detail: 'Filter by visit count' },
  { name: 'updated', label: 'Updated', detail: 'Filter by last update time' },
  { name: 'added', label: 'Added', detail: 'Filter by indexed time' },
  { name: 'sort', label: 'Sort', detail: 'Change result ordering' },
];

const SORT_VALUES = [
  { value: 'relevance', label: 'Relevance' },
  { value: 'date', label: 'Newest first' },
  { value: '-date', label: 'Oldest first' },
  { value: 'visits', label: 'Most visited' },
  { value: '-visits', label: 'Least visited' },
  { value: 'domain', label: 'Domain A to Z' },
  { value: '-domain', label: 'Domain Z to A' },
];

const TIME_VALUES = [
  { value: '<24h', label: 'Within the last 24 hours' },
  { value: '<7d', label: 'Within the last 7 days' },
  { value: '<30d', label: 'Within the last 30 days' },
  { value: '<365d', label: 'Within the last year' },
  { value: '>365d', label: 'More than one year ago' },
];

const DEFAULT_TYPE_VALUES: TermCount[] = [
  { term: 'web', count: 0 },
  { term: 'local', count: 0 },
];

const DEFAULT_VISIT_VALUES: TermCount[] = [
  { term: '1', count: 0, label: '1 visit' },
  { term: '2..4', count: 0, label: '2 to 4 visits' },
  { term: '5..9', count: 0, label: '5 to 9 visits' },
  { term: '10..', count: 0, label: '10 or more visits' },
];

const FACET_NAMES: Record<string, string> = {
  domain: 'domains',
  language: 'languages',
  type: 'types',
  visits: 'visits',
};

function isWhitespace(value: string): boolean {
  return /\s/.test(value);
}

function isInsideQuotes(query: string, cursor: number): boolean {
  let quoted = false;
  let escaped = false;
  for (const char of query.slice(0, cursor)) {
    if (escaped) {
      escaped = false;
      continue;
    }
    if (char === '\\') {
      escaped = true;
      continue;
    }
    if (char === '"') quoted = !quoted;
  }
  return quoted;
}

export function queryTokenAt(query: string, cursor: number): QueryToken {
  const safeCursor = Math.max(0, Math.min(cursor, query.length));
  if (
    safeCursor > 0 &&
    isWhitespace(query[safeCursor - 1]) &&
    (safeCursor === query.length || isWhitespace(query[safeCursor]))
  ) {
    return { start: safeCursor, end: safeCursor, value: '' };
  }

  let start = safeCursor;
  let end = safeCursor;
  while (start > 0 && !isWhitespace(query[start - 1])) start -= 1;
  while (end < query.length && !isWhitespace(query[end])) end += 1;
  return { start, end, value: query.slice(start, end) };
}

export function facetSuggestionContext(
  query: string,
  cursor: number,
): FacetSuggestionContext | null {
  if (isInsideQuotes(query, cursor)) return null;
  const token = queryTokenAt(query, cursor);
  const match = token.value.match(/^-?(domain|language|type|visits):(.*)$/i);
  if (!match) return null;

  const field = match[1].toLowerCase();
  const facetName = FACET_NAMES[field];
  const baseQuery = `${query.slice(0, token.start)} ${query.slice(token.end)}`
    .replace(/\s+/g, ' ')
    .trim();
  const normalizedBase = baseQuery || '*';
  return {
    baseQuery: normalizedBase,
    facetName,
    key: `${field}:${normalizedBase}`,
  };
}

function fieldValueSuggestions(
  field: string,
  valueFragment: string,
  tokenPrefix: string,
  facets?: FacetsResult | null,
): QuerySuggestion[] {
  const facetName = FACET_NAMES[field] ?? '';
  let values = facetName ? (facets?.terms?.[facetName]?.terms ?? []) : [];
  if (field === 'type') values = mergeDefaultValues(values, DEFAULT_TYPE_VALUES);
  if (field === 'visits') values = mergeDefaultValues(values, DEFAULT_VISIT_VALUES);

  const normalizedFragment = valueFragment.toLowerCase();
  return values
    .filter(
      ({ term, label }) =>
        term.toLowerCase().includes(normalizedFragment) ||
        (label ?? '').toLowerCase().includes(normalizedFragment),
    )
    .slice(0, 10)
    .map(({ term, count, label }) => ({
      id: `facet:${field}:${term}`,
      group: `${field[0].toUpperCase()}${field.slice(1)} values`,
      kind: 'facet' as const,
      label: label ?? term,
      detail: count > 0 ? `${count} results` : `${field}:${term}`,
      insertText: `${tokenPrefix}${field}:${term}`,
      replacement: 'token' as const,
      appendSpace: true,
    }));
}

function mergeDefaultValues(values: TermCount[], defaults: TermCount[]): TermCount[] {
  const merged = new Map(defaults.map((value) => [value.term, value]));
  for (const value of values) merged.set(value.term, value);
  return [...merged.values()];
}

function contextualValueSuggestions(
  tokenValue: string,
  facets?: FacetsResult | null,
): QuerySuggestion[] | null {
  const match = tokenValue.match(/^(-?)([a-z_]+):(.*)$/i);
  if (!match) return null;

  const tokenPrefix = match[1];
  const field = match[2].toLowerCase();
  const valueFragment = match[3];
  if (field === 'sort') {
    const normalizedFragment = valueFragment.toLowerCase();
    return SORT_VALUES.filter(
      ({ value, label }) =>
        value.toLowerCase().includes(normalizedFragment) ||
        label.toLowerCase().includes(normalizedFragment),
    ).map(({ value, label }) => ({
      id: `sort:${value}`,
      group: 'Sort values',
      kind: 'sort' as const,
      label,
      detail: `sort:${value}`,
      insertText: `${tokenPrefix}sort:${value}`,
      replacement: 'token' as const,
      appendSpace: true,
    }));
  }

  if (field === 'updated' || field === 'added') {
    const normalizedFragment = valueFragment.toLowerCase();
    return TIME_VALUES.filter(
      ({ value, label }) =>
        value.toLowerCase().includes(normalizedFragment) ||
        label.toLowerCase().includes(normalizedFragment),
    ).map(({ value, label }) => ({
      id: `time:${field}:${value}`,
      group: `${field[0].toUpperCase()}${field.slice(1)} values`,
      kind: 'facet' as const,
      label,
      detail: `${field}:${value}`,
      insertText: `${tokenPrefix}${field}:${value}`,
      replacement: 'token' as const,
      appendSpace: true,
    }));
  }

  if (['domain', 'language', 'type', 'visits'].includes(field)) {
    return fieldValueSuggestions(field, valueFragment, tokenPrefix, facets);
  }
  return [];
}

export function buildQuerySuggestions(options: QuerySuggestionOptions): QuerySuggestion[] {
  const {
    query,
    cursor,
    aliases,
    recentSearches,
    facets,
    serverSuggestion = '',
    limit = 12,
  } = options;
  const suggestions: QuerySuggestion[] = [];

  if (serverSuggestion && serverSuggestion !== query) {
    suggestions.push({
      id: `spelling:${serverSuggestion}`,
      group: 'Suggested query',
      kind: 'spelling',
      label: serverSuggestion,
      detail: 'Server suggestion',
      insertText: serverSuggestion,
      replacement: 'query',
    });
  }

  if (isInsideQuotes(query, cursor)) return suggestions.slice(0, limit);
  const token = queryTokenAt(query, cursor);
  const contextual = contextualValueSuggestions(token.value, facets);
  if (contextual !== null) return [...suggestions, ...contextual].slice(0, limit);

  const normalizedQuery = query.trim().toLowerCase();
  const recentMatches = recentSearches
    .filter((recent, index, all) => all.indexOf(recent) === index)
    .filter((recent) => !normalizedQuery || recent.toLowerCase().includes(normalizedQuery))
    .slice(0, normalizedQuery ? 3 : 5)
    .map((recent): QuerySuggestion => ({
      id: `recent:${recent}`,
      group: 'Recent searches',
      kind: 'recent',
      label: recent,
      detail: 'Run this query again',
      insertText: recent,
      replacement: 'query',
    }));
  suggestions.push(...recentMatches);

  const negated = token.value.startsWith('-');
  const tokenFragment = (negated ? token.value.slice(1) : token.value).toLowerCase();
  const aliasMatches = Object.entries(aliases)
    .filter(([keyword]) => !tokenFragment || keyword.toLowerCase().includes(tokenFragment))
    .slice(0, 4)
    .map(([keyword, expansion]): QuerySuggestion => ({
      id: `alias:${keyword}`,
      group: 'Aliases',
      kind: 'alias',
      label: keyword,
      detail: expansion,
      insertText: `${negated ? '-' : ''}${keyword}`,
      replacement: 'token',
      appendSpace: true,
    }));
  suggestions.push(...aliasMatches);

  const fieldMatches = SEARCH_FIELDS.filter(
    ({ name, label }) =>
      !tokenFragment ||
      name.startsWith(tokenFragment) ||
      label.toLowerCase().startsWith(tokenFragment),
  )
    .slice(0, tokenFragment ? 8 : 6)
    .map(({ name, label, detail }): QuerySuggestion => ({
      id: `field:${name}`,
      group: 'Search fields',
      kind: 'field',
      label,
      detail,
      insertText: `${negated ? '-' : ''}${name}:`,
      replacement: 'token',
      keepOpen: true,
    }));
  suggestions.push(...fieldMatches);
  return suggestions.slice(0, limit);
}

export function applyQuerySuggestion(
  query: string,
  cursor: number,
  suggestion: QuerySuggestion,
): { cursor: number; query: string } {
  if (suggestion.replacement === 'query') {
    return { query: suggestion.insertText, cursor: suggestion.insertText.length };
  }

  const token = queryTokenAt(query, cursor);
  const before = query.slice(0, token.start);
  let after = query.slice(token.end);
  let insertion = suggestion.insertText;
  if (suggestion.appendSpace) {
    insertion += ' ';
    after = after.replace(/^\s+/, '');
  }
  return {
    query: `${before}${insertion}${after}`,
    cursor: before.length + insertion.length,
  };
}
