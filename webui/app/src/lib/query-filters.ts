// SPDX-License-Identifier: AGPL-3.0-or-later

import { removeSortDirectives, sortDirectiveFromQuery } from './sort-directive.js';
import { hasTopLevelOr, queryTokens, type QuerySyntax } from './query-tokens.js';

export interface QueryFilter {
  start: number;
  end: number;
  field: string;
  value: string;
  negated: boolean;
}

// Only mandatory top level filters are safe to display as removable chips
export function queryFilters(query: string, fields: string[], syntax?: QuerySyntax): QueryFilter[] {
  if (hasTopLevelOr(query, syntax)) return [];
  const allowedFields = new Set(fields);
  const filters: QueryFilter[] = [];
  let notStart: number | null = null;
  let negated = false;
  for (const token of queryTokens(query, syntax)) {
    if (token.text === 'NOT') {
      notStart ??= token.start;
      negated = !negated;
      continue;
    }
    const match = token.text.match(/^(-*)([a-z_]+):\s*(.+)$/s);
    if (match && allowedFields.has(match[2]) && !/^-*\(/.test(match[3])) {
      let rawValue = match[3];
      let minusCount = match[1].length;
      while (rawValue.startsWith('-') && rawValue.length > 1) {
        minusCount += 1;
        rawValue = rawValue.slice(1);
      }
      let value = '';
      let escaped = false;
      let quoted = false;
      for (const char of rawValue) {
        if (escaped) {
          value += char;
          escaped = false;
        } else if (char === '\\') {
          escaped = true;
        } else if (char === '"') {
          quoted = !quoted;
        } else {
          value += char;
        }
      }
      if (value && !quoted && !escaped) {
        filters.push({
          start: notStart ?? token.start,
          end: token.end,
          field: match[2],
          value,
          negated: negated !== (minusCount % 2 === 1),
        });
      }
    }
    notStart = null;
    negated = false;
  }
  return filters;
}

export function removeQueryFilters(
  query: string,
  filters: QueryFilter[],
  syntax?: QuerySyntax,
): string {
  let remaining = query;
  for (const filter of [...filters].sort((a, b) => b.start - a.start)) {
    let before = remaining.slice(0, filter.start).trimEnd();
    let after = remaining.slice(filter.end).trimStart();
    const preceding = queryTokens(before, syntax).at(-1);
    const following = queryTokens(after, syntax)[0];
    if (preceding?.text === 'AND') {
      before = before.slice(0, preceding.start).trimEnd();
    } else if (following?.text === 'AND') {
      after = after.slice(following.end).trimStart();
    }
    remaining = [before, after].filter(Boolean).join(' ');
  }
  return remaining.trim();
}

export function appendQueryFilters(query: string, filters: string[], syntax?: QuerySyntax): string {
  if (filters.length === 0) return query.trim();
  let base = query.trim();
  let directive = '';
  if (hasTopLevelOr(base, syntax)) {
    const value = syntax ? sortDirectiveFromQuery(base, syntax) : null;
    if (syntax && value !== null) {
      base = removeSortDirectives(base, syntax);
      directive = `${syntax.sort.field}:${value}`;
    }
    base = `(${base})`;
  }
  return [base, ...filters, directive].filter(Boolean).join(' ');
}

export function toggleQueryFilter(
  query: string,
  field: string,
  value: string,
  syntax?: QuerySyntax,
): string {
  const matches = queryFilters(query, [field], syntax).filter(
    (filter) => !filter.negated && filter.value === value,
  );
  if (matches.length > 0) return removeQueryFilters(query, matches, syntax) || '*';
  const escapedValue =
    /[\s"\\()|]/.test(value) || value.startsWith('-')
      ? `"${value.replace(/["\\]/g, '\\$&')}"`
      : value;
  return appendQueryFilters(query, [`${field}:${escapedValue}`], syntax);
}
