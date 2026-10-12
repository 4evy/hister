import { expect, test } from '@playwright/test';
import { queryFilters, removeQueryFilters, toggleQueryFilter } from '../../src/lib/query-filters';
import { replaceTimeFilters, timeFilters } from '../../src/lib/time-filters';
import { sortDirectiveFromQuery } from '../../src/lib/sort-directive';
import { emptySearchCapabilities } from '../../src/lib/search-schema';

test('facet filters constrain every OR branch', () => {
  for (const query of ['privacy OR security', 'privacy|security']) {
    expect(toggleQueryFilter(query, 'domain', 'example.com')).toBe(`(${query}) domain:example.com`);
  }
  expect(queryFilters('domain:example.com OR security', ['domain'])).toEqual([]);
  expect(toggleQueryFilter('title:(privacy OR security)', 'domain', 'example.com')).toBe(
    'title:(privacy OR security) domain:example.com',
  );
  expect(toggleQueryFilter('privacy', 'label', 'a|b')).toBe('privacy label:"a|b"');
});

test('removing filters preserves Boolean syntax', () => {
  for (const query of [
    'privacy AND domain:example.com',
    'domain:example.com AND privacy',
    'privacy AND NOT domain:example.com',
  ]) {
    const filters = queryFilters(query, ['domain']);
    expect(filters).toHaveLength(1);
    expect(removeQueryFilters(query, filters)).toBe('privacy');
  }
  expect(queryFilters('NOT domain:example.com', ['domain'])[0].negated).toBe(true);
});

test('date filters respect groups and Boolean alternatives', () => {
  expect(timeFilters('title:privacy OR updated:<7d', 'updated')).toEqual([]);
  expect(timeFilters('(privacy updated:<7d security)', 'updated')).toEqual([]);
  expect(replaceTimeFilters('privacy OR security', 'updated', ['<7d'])).toBe(
    '(privacy OR security) updated:<7d',
  );
  expect(replaceTimeFilters('privacy AND updated:<7d', 'updated', ['<30d'])).toBe(
    'privacy updated:<30d',
  );
});

test('sort directives inside groups remain search text', () => {
  const sort = {
    field: 'sort',
    label: 'Sort',
    description: '',
    options: [{ value: 'date', label: 'Date', visible: true }],
  };
  const syntax = { ...emptySearchCapabilities(), sort };
  expect(sortDirectiveFromQuery('(privacy sort:date security)', syntax)).toBeNull();
  expect(sortDirectiveFromQuery('NOT sort:date', syntax)).toBeNull();
  expect(sortDirectiveFromQuery('(privacy OR security) sort:date', syntax)).toBe('date');
  expect(toggleQueryFilter('privacy OR security sort:date', 'domain', 'example.com', syntax)).toBe(
    '(privacy OR security) domain:example.com sort:date',
  );
  expect(replaceTimeFilters('privacy OR security sort:date', 'updated', ['<7d'], syntax)).toBe(
    '(privacy OR security) updated:<7d sort:date',
  );
});

test('quoted operators and regexp pipes are not Boolean alternatives', () => {
  for (const query of [
    '"privacy OR security"',
    'url_re:privacy|security',
    'title:(privacy|security)',
  ]) {
    expect(toggleQueryFilter(query, 'domain', 'example.com')).toBe(`${query} domain:example.com`);
  }
  expect(toggleQueryFilter('privacy', 'label', '-archived')).toBe('privacy label:"-archived"');
});

test('spaces after field names preserve scope when editing filters', () => {
  const syntax = emptySearchCapabilities();
  syntax.fields = ['title', 'domain', 'updated', 'url_re'].map((name) => ({
    name,
    label: name,
    description: '',
    kind: 'text',
  }));
  syntax.sort = {
    field: 'sort',
    label: 'Sort',
    description: '',
    options: [{ value: 'date', label: 'Date', visible: true }],
  };
  expect(queryFilters('title: domain:example.com', ['domain'], syntax)).toEqual([]);
  expect(queryFilters('domain: example.com', ['domain'], syntax)[0].value).toBe('example.com');
  expect(sortDirectiveFromQuery('title: sort:date', syntax)).toBeNull();
  for (const query of ['title: OR security', 'url_re: privacy|security']) {
    expect(toggleQueryFilter(query, 'domain', 'example.com', syntax)).toBe(
      `${query} domain:example.com`,
    );
  }
  expect(toggleQueryFilter('unknown: OR security', 'domain', 'example.com', syntax)).toBe(
    '(unknown: OR security) domain:example.com',
  );
});

test('negation inside field values produces the correct filter chip', () => {
  for (const query of ['domain:-example.com', 'NOT domain:--example.com']) {
    expect(queryFilters(query, ['domain'])[0]).toMatchObject({
      value: 'example.com',
      negated: true,
    });
  }
  expect(queryFilters('NOT domain:-example.com', ['domain'])[0]).toMatchObject({
    value: 'example.com',
    negated: false,
  });
  expect(queryFilters('domain:"-example.com"', ['domain'])[0]).toMatchObject({
    value: '-example.com',
    negated: false,
  });
  expect(queryFilters('domain:--(example.com OR example.org)', ['domain'])).toEqual([]);
});
