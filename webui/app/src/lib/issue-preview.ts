// SPDX-License-Identifier: AGPL-3.0-or-later

export interface IssueComment {
  author?: string;
  authorURL?: string;
  createdAt?: string;
  url?: string;
  bodyHTML: string;
}

export interface IssuePreviewData {
  title: string;
  number?: string;
  repository?: string;
  url?: string;
  state?: string;
  labels?: string[];
  body: IssueComment;
  comments: IssueComment[];
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function optionalStrings(value: Record<string, unknown>, keys: string[]): boolean {
  return keys.every((key) => value[key] === undefined || typeof value[key] === 'string');
}

function isComment(value: unknown): value is IssueComment {
  return (
    isRecord(value) &&
    typeof value.bodyHTML === 'string' &&
    optionalStrings(value, ['author', 'authorURL', 'createdAt', 'url'])
  );
}

export function parseIssuePreview(content: string): IssuePreviewData {
  const data: unknown = JSON.parse(content);
  if (
    !isRecord(data) ||
    typeof data.title !== 'string' ||
    !optionalStrings(data, ['number', 'repository', 'url', 'state']) ||
    !isComment(data.body) ||
    !Array.isArray(data.comments) ||
    !data.comments.every(isComment) ||
    (data.labels !== undefined &&
      (!Array.isArray(data.labels) || !data.labels.every((label) => typeof label === 'string')))
  ) {
    throw new Error('Invalid issue preview');
  }
  return data as unknown as IssuePreviewData;
}

export function issueLink(value?: string): string | undefined {
  if (!value) return undefined;
  try {
    const url = new URL(value);
    if ((url.protocol === 'https:' || url.protocol === 'http:') && !url.username && !url.password)
      return url.href;
  } catch {
    return undefined;
  }
  return undefined;
}
