// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SearchCapabilities } from '#lib/search-schema.js';

export type QuerySyntax = Pick<SearchCapabilities, 'fields' | 'sort'>;

export interface QueryToken {
  start: number;
  end: number;
  text: string;
}

// Return top level tokens for editing, keeping groups and quoted values intact
// Only the server interprets the matching expression inside each token
export function queryTokens(query: string, syntax?: QuerySyntax): QueryToken[] {
  const tokens: QueryToken[] = [];
  const fields = new Set(syntax?.fields.map((field) => field.name));
  let start = 0;
  while (start < query.length) {
    if (/\s/.test(query[start])) {
      start += 1;
      continue;
    }
    let end = start;
    const prefix = query.slice(start).match(/^-*([a-z_][\w.]*):\s*/);
    const rawRegexp = /^-*url_re:\s*[^"\s]/.test(query.slice(start));
    if (prefix && (fields.has(prefix[1]) || /^metadata\..+/.test(prefix[1]) || rawRegexp)) {
      end += prefix[0].length;
    }
    let quoted = false;
    let escaped = false;
    let depth = 0;
    let inClass = false;
    for (; end < query.length; end += 1) {
      const char = query[end];
      if (escaped) {
        escaped = false;
        continue;
      }
      if (char === '\\') {
        escaped = true;
        continue;
      }
      if (char === '"') quoted = !quoted;
      if (!quoted) {
        if (rawRegexp && char === '[') inClass = true;
        if (rawRegexp && char === ']') inClass = false;
        if (!inClass) {
          if (char === '(') depth += 1;
          if (char === ')') depth -= 1;
        }
        if (/\s/.test(char) && (depth === 0 || rawRegexp)) break;
        if (char === '|' && depth === 0 && !rawRegexp) {
          if (end === start) end += 1;
          break;
        }
      }
    }
    tokens.push({ start, end, text: query.slice(start, end) });
    start = end;
  }
  return tokens;
}

export function hasTopLevelOr(query: string, syntax?: QuerySyntax): boolean {
  return queryTokens(query, syntax).some(({ text }) => text === 'OR' || text === '|');
}

class QueryTokenState {
  quoted = false;
  escaped = false;
  depth = 0;

  advance(char: string) {
    if (this.escaped) {
      this.escaped = false;
      return;
    }
    if (char === '\\') {
      this.escaped = true;
      return;
    }
    if (char === '"') this.quoted = !this.quoted;
    this.updateDepth(char);
  }

  private updateDepth(char: string) {
    if (this.quoted) return;
    if (char === '(') this.depth += 1;
    if (char === ')') this.depth -= 1;
  }

  get complete() {
    return !this.quoted && this.depth === 0;
  }
}

export function isInsideQuotes(query: string, cursor: number): boolean {
  const state = new QueryTokenState();
  for (const char of query.slice(0, cursor)) state.advance(char);
  return state.quoted;
}
