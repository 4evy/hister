// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SearchResult, SearchResults, SemanticHit } from './search';

export interface MergedResult extends SearchResult {
  semanticScore?: number;
  finalScore: number;
  sourceType: 'keyword' | 'semantic' | 'both';
}

interface MergeOptions {
  semanticEnabled: boolean;
  weight: number;
  sort: string;
  userId?: number;
}

function compareStrings(a: string, b: string): number {
  return a < b ? -1 : a > b ? 1 : 0;
}

function resultDate(result: SearchResult): number {
  return result.updated ?? result.added ?? 0;
}

function compareDates(a: SearchResult, b: SearchResult): number {
  return resultDate(b) - resultDate(a);
}

const resultComparators: Record<string, (a: MergedResult, b: MergedResult) => number> = {
  date: compareDates,
  visits: (a, b) => (b.add_count ?? 1) - (a.add_count ?? 1) || compareDates(a, b),
  domain: (a, b) => compareStrings(a.domain, b.domain),
  score: (a, b) => b.finalScore - a.finalScore || compareDates(a, b),
};

function compareIdentity(a: SearchResult, b: SearchResult): number {
  return compareStrings(a.id ?? a.url, b.id ?? b.url);
}

// Match the ordering and tie breakers in server/indexer/searchschema
function compareResults(a: MergedResult, b: MergedResult, sort: string): number {
  const direction = sort.startsWith('-') ? -1 : 1;
  const field = sort.replace(/^-/, '');
  const compare = Object.hasOwn(resultComparators, field)
    ? resultComparators[field]
    : resultComparators.score;
  return direction * compare(a, b) || compareIdentity(a, b);
}

function documentIdentity(doc: SearchResult, userId?: number): string {
  const owner = doc.user_id ?? userId;
  return doc.id ?? (owner ? `${owner}:${doc.url}` : doc.url);
}

function keywordScore(doc: SearchResult): number {
  return doc.score ?? 0;
}

function blendResult(
  doc: SearchResult,
  semanticScore: number | undefined,
  weight: number,
  maxScore: number,
): MergedResult {
  return {
    ...doc,
    semanticScore,
    finalScore: (1 - weight) * (keywordScore(doc) / maxScore) + weight * (semanticScore ?? 0),
    sourceType: semanticScore === undefined ? 'keyword' : 'both',
  };
}

function addSemanticHits(merged: Map<string, MergedResult>, hits: SemanticHit[], weight: number) {
  for (const hit of hits) {
    const doc = hit.document;
    if (!doc || merged.has(doc.url)) continue;
    merged.set(doc.url, semanticResult(doc, hit, weight));
  }
}

function semanticResult(doc: SearchResult, hit: SemanticHit, weight: number): MergedResult {
  return {
    ...doc,
    id: doc.id || hit.doc_id,
    semanticScore: hit.similarity,
    finalScore: weight * hit.similarity,
    sourceType: 'semantic',
  };
}

function blendResults(docs: SearchResult[], hits: SemanticHit[], options: MergeOptions) {
  const maxScore = Math.max(...docs.map(keywordScore), 1);
  const semanticScores = new Map(hits.map((hit) => [hit.doc_id, hit.similarity]));
  return new Map(
    docs.map((doc) => {
      const semanticScore =
        semanticScores.get(documentIdentity(doc, options.userId)) ?? semanticScores.get(doc.url);
      return [doc.url, blendResult(doc, semanticScore, options.weight, maxScore)];
    }),
  );
}

export function mergeSearchResults(
  docs: SearchResult[],
  hits: SemanticHit[] | undefined,
  options: MergeOptions,
): MergedResult[] {
  if (!options.semanticEnabled || !hits?.length) {
    return docs.map((doc) => ({ ...doc, finalScore: keywordScore(doc), sourceType: 'keyword' }));
  }
  const merged = blendResults(docs, hits, options);
  addSemanticHits(merged, hits, options.weight);
  return [...merged.values()].sort((a, b) => compareResults(a, b, options.sort));
}

function resultDocuments(results: SearchResults): SearchResult[] {
  return [
    ...(results.documents ?? []),
    ...(results.history ?? []),
    ...(results.semantic_hits ?? []).flatMap((hit) => (hit.document ? [hit.document] : [])),
  ];
}

export function removeSearchResults(
  results: SearchResults,
  matches: (doc: SearchResult) => boolean,
): SearchResults {
  const removedURLs = new Set(
    resultDocuments(results)
      .filter(matches)
      .map((doc) => doc.url),
  );
  const keepDocument = (doc: SearchResult) => !removedURLs.has(doc.url);

  return {
    ...results,
    documents: results.documents?.filter(keepDocument),
    history: results.history?.filter(keepDocument),
    semantic_hits: results.semantic_hits?.filter(
      (hit) => !removedURLs.has(hit.document?.url ?? hit.doc_id.replace(/^\d+:/, '')),
    ),
    total: results.total === undefined ? undefined : Math.max(0, results.total - removedURLs.size),
    facets: undefined,
  };
}
