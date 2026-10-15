type PageData = {
  title: string;
  text: string;
  url: string;
  html: string;
  faviconURL: string;
};

type PageState = Omit<PageData, 'html'> & { metadata: string };

type Result = {
  title: string;
  url: string;
  query: string;
};

type ExtractorCallback = (r: Result) => void;

interface ResultExtractor {
  isMatch(w: Window): boolean;
  setCallback(d: Document, cb: ExtractorCallback): void;
}

class SearchResultExtractor implements ResultExtractor {
  constructor(
    private matches: (location: Location) => boolean,
    private titleSelector: string,
    private resultSelector: string,
    private querySelector: string,
  ) {}

  isMatch(w: Window) {
    return this.matches(w.location);
  }

  setCallback(d: Document, cb: ExtractorCallback) {
    d.body.addEventListener('click', (event) => {
      const title = event.target;
      if (!(title instanceof HTMLElement) || !title.matches(this.titleSelector)) return;
      const result = title.closest<HTMLAnchorElement>(this.resultSelector);
      if (!result) return;
      cb({
        url: result.getAttribute('href')!,
        title: title.innerText,
        query: d.querySelector<HTMLInputElement | HTMLTextAreaElement>(this.querySelector)!.value,
      });
    });
  }
}

const resultExtractors: ResultExtractor[] = [
  new SearchResultExtractor(
    (location) => location.hostname === 'www.google.com' && location.pathname === '/search',
    'h3',
    'a[jsname="UWckNb"]',
    "textarea[name='q']",
  ),
  new SearchResultExtractor(
    (location) =>
      /^(noai\.|www\.)?duckduckgo.com$/.test(location.hostname) && location.pathname === '/',
    'span',
    'a[class="eVNpHGjtxRBq_gLOfGDr LQNqh2U1kzYxREs65IJu"]',
    "input[name='q']",
  ),
];

function getPageURL() {
  let url = new URL(window.location.href);
  const canonicalHref = document
    .querySelector('link[rel~="canonical" i][href]')
    ?.getAttribute('href')
    ?.trim();
  if (canonicalHref) {
    try {
      const canonicalURL = new URL(canonicalHref, document.baseURI);
      // Compare parsed hostnames exactly so another site or subdomain cannot
      // use this page's content to overwrite a document in the index.
      if (
        (canonicalURL.protocol === 'http:' || canonicalURL.protocol === 'https:') &&
        canonicalURL.hostname === url.hostname &&
        !canonicalURL.username &&
        !canonicalURL.password
      ) {
        url = canonicalURL;
      }
    } catch {}
  }
  url.hash = '';
  return url.href;
}

// Read the fields that can justify an automatic update without serializing
// the entire DOM. Other markup changes are captured by periodic preview checks.
function extractPageState(): PageState {
  const url = getPageURL();
  let faviconURL = '';
  try {
    const faviconHref = document.querySelector("link[rel~='icon']")?.getAttribute('href');
    faviconURL = new URL(faviconHref || '/favicon.ico', document.baseURI).href;
  } catch {}

  return {
    text: document.body?.innerText ?? '',
    title: document.querySelector('title')?.innerText ?? document.title,
    url,
    faviconURL,
    metadata: JSON.stringify(
      Array.from(
        document.querySelectorAll(
          'meta[name], meta[property], link[rel~="canonical" i], script[type="application/ld+json"]',
        ),
        (el) => [
          el.tagName,
          el.getAttribute('name'),
          el.getAttribute('property'),
          el.getAttribute('content'),
          el.getAttribute('href'),
          el.tagName === 'SCRIPT' ? el.textContent : null,
        ],
      ),
    ),
  };
}

function extractPageData(state: PageState): PageData {
  const { metadata, ...data } = state;
  return { ...data, html: document.documentElement?.innerHTML ?? '' };
}

function registerResultExtractor(w: Window, cb: ExtractorCallback) {
  for (let ex of resultExtractors) {
    if (ex.isMatch(w)) {
      ex.setCallback(w.document, cb);
      return;
    }
  }
}

export {
  type PageData,
  type PageState,
  registerResultExtractor,
  getPageURL,
  extractPageState,
  extractPageData,
};
