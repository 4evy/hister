// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

import { isInsideQuotes } from './query-tokens.js';

import type { FacetsResult, TermCount } from '#lib/search.js';
import { valuesForField } from '#lib/search-schema.js';
import type { SearchCapabilities, SearchFieldDefinition } from '#lib/search-schema.js';

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
  capabilities: SearchCapabilities;
  facets?: FacetsResult | null;
  serverSuggestion?: string;
  limit?: number;
}

export interface FacetSuggestionContext {
  baseQuery: string;
  facetName: string;
  key: string;
}

function isWhitespace(value: string): boolean {
  return /\s/.test(value);
}

export function queryTokenAt(query: string, cursor: number): QueryToken {
  const safeCursor = Math.max(0, Math.min(cursor, query.length));
  if (isWhitespaceGap(query, safeCursor)) {
    return { start: safeCursor, end: safeCursor, value: '' };
  }
  const left = query.slice(0, safeCursor).search(/\S+$/);
  const right = query.slice(safeCursor).search(/\s/);
  const start = left < 0 ? safeCursor : left;
  const end = right < 0 ? query.length : safeCursor + right;
  return { start, end, value: query.slice(start, end) };
}

function isWhitespaceGap(query: string, cursor: number): boolean {
  return cursor > 0 && isWhitespace(query[cursor - 1]) && atWhitespaceEnd(query, cursor);
}

function atWhitespaceEnd(query: string, cursor: number): boolean {
  return cursor === query.length || isWhitespace(query[cursor]);
}

export function facetSuggestionContext(
  query: string,
  cursor: number,
  capabilities: SearchCapabilities,
): FacetSuggestionContext | null {
  if (isInsideQuotes(query, cursor)) return null;
  const token = queryTokenAt(query, cursor);
  const match = token.value.match(/^-?([a-z_]+):(.*)$/i);
  if (!match) return null;

  const field = match[1].toLowerCase();
  const facetName = suggestibleFacet(field, capabilities);
  if (!facetName) return null;
  const baseQuery = `${query.slice(0, token.start)} ${query.slice(token.end)}`
    .replace(/\s+/g, ' ')
    .trim();
  const normalizedBase = defaultFacetQuery(baseQuery);
  return {
    baseQuery: normalizedBase,
    facetName,
    key: `${field}:${normalizedBase}`,
  };
}

function suggestibleFacet(field: string, capabilities: SearchCapabilities): string | null {
  const facetName = capabilities.fields.find((definition) => definition.name === field)?.facet;
  const facet = capabilities.facets.find(
    (definition) => !!facetName && definition.name === facetName,
  );
  if (!facet) return null;
  return facet.kind === 'date_ranges' ? null : facet.name;
}

function fieldValueSuggestions(
  field: SearchFieldDefinition,
  valueFragment: string,
  tokenPrefix: string,
  capabilities: SearchCapabilities,
  facets?: FacetsResult | null,
): QuerySuggestion[] {
  const facetValues = field.facet ? (facets?.terms?.[field.facet]?.terms ?? []) : [];
  const configuredValues: TermCount[] = valuesForField(capabilities, field).map((value) => ({
    term: value.value,
    count: 0,
    label: value.label,
  }));
  const values = mergeDefaultValues(facetValues, configuredValues);

  const normalizedFragment = valueFragment.toLowerCase();
  return values
    .filter(
      ({ term, label }) =>
        term.toLowerCase().includes(normalizedFragment) ||
        (label ?? '').toLowerCase().includes(normalizedFragment),
    )
    .slice(0, 10)
    .map(({ term, count, label }) => ({
      id: `facet:${field.name}:${term}`,
      group: `${field.label} values`,
      kind: 'facet' as const,
      label: label ?? term,
      detail: count > 0 ? `${count} results` : `${field.name}:${term}`,
      insertText: `${tokenPrefix}${field.name}:${term}`,
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
  capabilities: SearchCapabilities,
  facets?: FacetsResult | null,
): QuerySuggestion[] | null {
  const match = tokenValue.match(/^(-?)([a-z_]+):(.*)$/i);
  if (!match) return null;

  const tokenPrefix = match[1];
  const field = match[2].toLowerCase();
  const valueFragment = match[3];
  if (field === capabilities.sort.field)
    return sortValueSuggestions(valueFragment, tokenPrefix, capabilities);

  const definition = capabilities.fields.find((candidate) => candidate.name === field);
  if (hasSuggestionValues(definition)) {
    return fieldValueSuggestions(definition, valueFragment, tokenPrefix, capabilities, facets);
  }
  return [];
}

function hasSuggestionValues(
  field: SearchFieldDefinition | undefined,
): field is SearchFieldDefinition {
  return Boolean(field && (field.facet || field.valueSet));
}

function sortValueSuggestions(
  valueFragment: string,
  tokenPrefix: string,
  capabilities: SearchCapabilities,
): QuerySuggestion[] {
  const normalizedFragment = valueFragment.toLowerCase();
  return capabilities.sort.options
    .filter((option) => option.visible)
    .filter(
      ({ value, label }) =>
        value.toLowerCase().includes(normalizedFragment) ||
        label.toLowerCase().includes(normalizedFragment),
    )
    .map(({ value, label }) => ({
      id: `sort:${value}`,
      group: `${capabilities.sort.label} values`,
      kind: 'sort' as const,
      label,
      detail: `${capabilities.sort.field}:${value}`,
      insertText: `${tokenPrefix}${capabilities.sort.field}:${value}`,
      replacement: 'token' as const,
      appendSpace: true,
    }));
}

export function buildQuerySuggestions(options: QuerySuggestionOptions): QuerySuggestion[] {
  const {
    query,
    cursor,
    aliases,
    recentSearches,
    capabilities,
    facets,
    serverSuggestion = '',
    limit = 12,
  } = options;
  const suggestions: QuerySuggestion[] = [];
  const normalizedQuery = query.trim().toLowerCase();
  const normalizedServerSuggestion = serverSuggestion.trim().toLowerCase();

  suggestions.push(
    ...spellingSuggestions(serverSuggestion, normalizedServerSuggestion, normalizedQuery),
  );

  if (isInsideQuotes(query, cursor)) return suggestions.slice(0, limit);
  const token = queryTokenAt(query, cursor);
  const contextual = contextualValueSuggestions(token.value, capabilities, facets);
  if (contextual !== null) return [...suggestions, ...contextual].slice(0, limit);

  suggestions.push(
    ...recentSuggestions(recentSearches, normalizedQuery, normalizedServerSuggestion),
  );

  const negated = token.value.startsWith('-');
  const tokenFragment = (negated ? token.value.slice(1) : token.value).toLowerCase();
  suggestions.push(...aliasSuggestions(aliases, tokenFragment, negated));

  suggestions.push(...fieldSuggestions(capabilities, tokenFragment, negated));

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

function recentSuggestions(
  recentSearches: string[],
  normalizedQuery: string,
  normalizedServerSuggestion: string,
): QuerySuggestion[] {
  return recentSearches
    .filter((recent, index, all) => all.indexOf(recent) === index)
    .filter((recent) => {
      const normalizedRecent = recent.trim().toLowerCase();
      return (
        normalizedRecent !== normalizedQuery && normalizedRecent !== normalizedServerSuggestion
      );
    })
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
}
function aliasSuggestions(
  aliases: Record<string, string>,
  tokenFragment: string,
  negated: boolean,
): QuerySuggestion[] {
  return Object.entries(aliases)
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
}
function fieldSuggestions(
  capabilities: SearchCapabilities,
  tokenFragment: string,
  negated: boolean,
): QuerySuggestion[] {
  const searchFields = [
    ...capabilities.fields,
    {
      name: capabilities.sort.field,
      label: capabilities.sort.label,
      description: capabilities.sort.description,
    },
  ].filter((field) => field.name);
  return searchFields
    .filter(
      ({ name, label }) =>
        !tokenFragment ||
        name.startsWith(tokenFragment) ||
        label.toLowerCase().startsWith(tokenFragment),
    )
    .slice(0, tokenFragment ? 8 : 6)
    .map(({ name, label, description }): QuerySuggestion => ({
      id: `field:${name}`,
      group: 'Search fields',
      kind: 'field',
      label,
      detail: description,
      insertText: `${negated ? '-' : ''}${name}:`,
      replacement: 'token',
      keepOpen: true,
    }));
}

function spellingSuggestions(
  serverSuggestion: string,
  normalizedServerSuggestion: string,
  normalizedQuery: string,
): QuerySuggestion[] {
  if (normalizedServerSuggestion && normalizedServerSuggestion !== normalizedQuery) {
    return [
      {
        id: `spelling:${serverSuggestion}`,
        group: 'Suggested query',
        kind: 'spelling',
        label: serverSuggestion,
        detail: 'Server suggestion',
        insertText: serverSuggestion,
        replacement: 'query',
      },
    ];
  }

  return [];
}

function defaultFacetQuery(query: string): string {
  return query || '*';
}
