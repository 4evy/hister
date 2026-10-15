import { appendQueryFilters, queryFilters, removeQueryFilters } from './query-filters.js';
import type { QuerySyntax } from './query-tokens.js';

interface TimeFilter {
  comparison: string;
  value: string;
}

export function shiftISODate(value: string, days: number): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return '';
  const date = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(date.getTime()) || date.toISOString().slice(0, 10) !== value) return '';
  date.setUTCDate(date.getUTCDate() + days);
  return date.toISOString().slice(0, 10);
}

function isValidTimeValue(value: string): boolean {
  return /^\d+[smhdw]$/i.test(value) || shiftISODate(value, 0) !== '';
}

function validTimeFilters(text: string, field: string, syntax?: QuerySyntax) {
  return queryFilters(text, [field], syntax).flatMap((filter) => {
    const match = filter.value.match(/^(<=|>=|<|>)(\d+[smhdwSMHDW]|\d{4}-\d{2}-\d{2})$/);
    if (filter.negated || !match || !isValidTimeValue(match[2])) return [];
    return [{ filter, comparison: match[1], value: match[2] }];
  });
}

export function timeFilters(text: string, field: string, syntax?: QuerySyntax): TimeFilter[] {
  return validTimeFilters(text, field, syntax).map(({ comparison, value }) => ({
    comparison,
    value,
  }));
}

export function removeTimeFilters(text: string, field: string, syntax?: QuerySyntax): string {
  return removeQueryFilters(
    text,
    validTimeFilters(text, field, syntax).map(({ filter }) => filter),
    syntax,
  );
}

export function replaceTimeFilters(
  text: string,
  field: string,
  filters: string[],
  syntax?: QuerySyntax,
): string {
  if (!field) return text;
  const remaining = removeTimeFilters(text, field, syntax);
  const tokens = filters.map((filter) => `${field}:${filter}`);
  return appendQueryFilters(remaining, tokens, syntax);
}

export function customDatesFromQuery(
  text: string,
  field: string,
  syntax?: QuerySyntax,
): { from: string; to: string } {
  let from = '';
  let to = '';
  for (const filter of timeFilters(text, field, syntax).filter((filter) =>
    /^\d{4}-\d{2}-\d{2}$/.test(filter.value),
  )) {
    if (filter.comparison === '>=') from = filter.value;
    if (filter.comparison === '<') to = shiftISODate(filter.value, -1);
  }
  return { from, to };
}
