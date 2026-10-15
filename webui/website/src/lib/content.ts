export function contentSlug(path: string, extension = '.md'): string {
  return path.split('/').pop()?.replace(extension, '') ?? path;
}

export function contentMetadata<T extends object>(module: unknown): Partial<T> {
  return (module as { metadata?: T } | undefined)?.metadata ?? {};
}
