const pages = [
  {
    title: 'Extractors | Hister Documentation',
    path: '/docs/extractors',
    domain: 'hister.org',
    snippet:
      'Extractors are the components responsible for turning <mark>raw HTML</mark> or file content into rich, searchable data. Specialist extractors understand particular sites and file formats.',
    paragraphs: [
      'Extractors are the components responsible for turning raw HTML or file content into rich, searchable data. Specialist extractors understand particular sites and file formats. General extractors handle everything else.',
      'Generic HTML-to-text conversion loses a lot of signal. A Stack Overflow answer, a Go package reference, a local Markdown note, and a news article all have different structure and a one-size-fits-all parser cannot take advantage of that structure.',
    ],
    heading: 'Purpose',
  },
  {
    title: 'CommonMark Spec',
    path: '/0.31.2/#raw-html',
    domain: 'spec.commonmark.org',
    snippet:
      'Text between &lt; and &gt; that looks like an HTML tag is parsed as a <mark>raw HTML</mark> tag…',
    paragraphs: ['Text between < and > that looks like an HTML tag is parsed as a raw HTML tag…'],
    heading: '6.6 Raw HTML',
  },
  {
    title: 'Extractors: An Ecosystem for Better Search',
    path: '/posts/extractors-an-ecosystem-for-better-search',
    domain: 'hister.org',
    snippet:
      'Extractors are the layer between <mark>raw HTML</mark> and the indexed, searchable content you actually want to find.',
    paragraphs: [
      'Every page you add to Hister arrives as raw HTML. Raw HTML is noisy. A Stack Overflow question, a Go package reference, a GitHub issue, and a news article all contain the same structural clutter: navigation bars, cookie banners, sidebars, footers, and advertising wrappers.',
      'Extractors are the layer between raw HTML and the indexed, searchable content you actually want to find.',
    ],
    heading: 'From HTML to searchable content',
  },
  {
    title: 'Markdown Syntax Documentation',
    path: '/projects/markdown/syntax#link',
    domain: 'daringfireball.net',
    snippet: 'In the <mark>raw HTML</mark>, there’s more markup than there is text.',
    paragraphs: ['In the raw HTML, there’s more markup than there is text.'],
    heading: 'Links',
  },
];

export const exampleDocuments = pages.map((page) => ({
  ...page,
  url: `https://${page.domain}${page.path}`,
  href: page.domain === 'hister.org' ? page.path : `https://${page.domain}${page.path}`,
}));
