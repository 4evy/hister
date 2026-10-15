import type { StoredSettings } from '../modules/settings';
import type { PageData } from '../modules/extract';
import { fetchAPI, sendPageData, sendPDFData, sendResult } from '../modules/network';
import { ensureDefaultServerURL } from '../modules/settings';
import { getTabPageURL } from '../modules/tabs';

void ensureDefaultServerURL();

const missingURLMsg = {
  error: 'Missing or invalid Hister server URL. Configure it in the addon popup.',
};

interface BackgroundRequest {
  action?: string;
  tabId?: number;
  url: string;
  sourceURL?: string;
  pattern: string;
  deleteQuery?: string;
  pageData?: PageData & { metadata?: Record<string, unknown>; label?: string };
  resultData?: unknown;
}

type CustomHeader = { name: string; value: string };
type SkipRuleType = 'url' | 'domain';
type IndexingResult = { status?: string; status_code?: number; error?: string };

function escapeRegex(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function buildUrlSkipPattern(url: string): string {
  return escapeRegex(url) + '$';
}

function buildDomainSkipPattern(url: string): string {
  try {
    return escapeRegex(new URL(url).origin);
  } catch (_) {
    return escapeRegex(url);
  }
}

// --- Badge helpers ---

function setErrorBadge(tabId: number) {
  chrome.action.setBadgeText({ text: '!', tabId }, () => void chrome.runtime.lastError);
  chrome.action.setBadgeBackgroundColor(
    { color: '#ff4444', tabId },
    () => void chrome.runtime.lastError,
  );
}

function setPreviouslyIndexedBadge(tabId: number) {
  chrome.action.setBadgeText({ text: '✓', tabId }, () => void chrome.runtime.lastError);
  chrome.action.setBadgeBackgroundColor(
    { color: '#44aa44', tabId },
    () => void chrome.runtime.lastError,
  );
  try {
    void chrome.action.setBadgeTextColor?.({ color: '#ffffff', tabId })?.catch(() => {});
  } catch (_) {}
}

function clearBadge(tabId: number) {
  chrome.action.setBadgeText({ text: '', tabId }, () => void chrome.runtime.lastError);
}

// --- Icon helpers ---

let greyIconCache: Record<number, ImageData> | null = null;
let normalIconCache: Record<number, ImageData> | null = null;

async function buildIcons(grey: boolean): Promise<Record<number, ImageData>> {
  const response = await fetch(chrome.runtime.getURL('assets/icons/icon128.png'));
  const blob = await response.blob();
  const bitmap = await createImageBitmap(blob);
  const result: Record<number, ImageData> = {};
  for (const size of [16, 32]) {
    const canvas = new OffscreenCanvas(size, size);
    const ctx = canvas.getContext('2d')!;
    ctx.drawImage(bitmap, 0, 0, size, size);
    const imageData = ctx.getImageData(0, 0, size, size);
    if (grey) {
      for (let i = 0; i < imageData.data.length; i += 4) {
        const lum =
          imageData.data[i] * 0.299 + imageData.data[i + 1] * 0.587 + imageData.data[i + 2] * 0.114;
        imageData.data[i] = lum;
        imageData.data[i + 1] = lum;
        imageData.data[i + 2] = lum;
        imageData.data[i + 3] = Math.round(imageData.data[i + 3] * 0.5);
      }
    }
    result[size] = imageData;
  }
  return result;
}

async function setGreyIcon(tabId: number): Promise<void> {
  clearBadge(tabId);
  try {
    if (!greyIconCache) greyIconCache = await buildIcons(true);
    chrome.action.setIcon({ imageData: greyIconCache, tabId }, () => void chrome.runtime.lastError);
  } catch (_) {}
}

async function setNormalIcon(tabId: number): Promise<void> {
  try {
    if (!normalIconCache) normalIconCache = await buildIcons(false);
    chrome.action.setIcon(
      { imageData: normalIconCache, tabId },
      () => void chrome.runtime.lastError,
    );
  } catch (_) {}
}

// --- Per-tab sensitive-content rejection state ---

// Maps tabId → URL that was last rejected due to sensitive content.
const tabSensitiveState = new Map<number, string>();

chrome.tabs.onRemoved.addListener((tabId) => {
  tabSensitiveState.delete(tabId);
  forgetNavigationStatus(tabId);
});

// --- Main frame HTTP status tracking ---

// `PerformanceNavigationTiming.responseStatus` is only implemented by Chromium,
// so the content script cannot tell an error page from a successful one on
// Firefox. Capture the status of main frame navigations here instead and gate
// document submission on it.

type NavigationStatus = { url: string; statusCode: number };

// Maps tabId → the most recent main frame responses of that tab, oldest first.
// A short history is needed because a page submits a final snapshot while it is
// being replaced, by which time the next navigation is already recorded.
// Mirrored in chrome.storage.session (when available) so the state survives a
// suspended MV3 background script.
const MAX_TRACKED_NAVIGATIONS = 5;
const tabNavigationStatus = new Map<number, NavigationStatus[]>();

// `chrome.storage.session` requires Chrome >= 102 / Firefox >= 115.
const sessionStore = chrome.storage.session;

function navigationStatusKey(tabId: number): string {
  return `navStatus:${tabId}`;
}

// The content script strips the fragment from the submitted URL, and
// webRequest URLs never carry one.
function normalizeNavigationURL(url: string): string {
  const i = url.indexOf('#');
  return i === -1 ? url : url.slice(0, i);
}

// Redirects (3xx) are followed by the browser and overwritten by the status of
// the redirect target, and 304 responses render a perfectly valid cached page,
// so only 4xx/5xx are treated as unsuccessful.
function isErrorStatus(statusCode: number): boolean {
  return statusCode >= 400;
}

function recordNavigationStatus(tabId: number, url: string, statusCode: number) {
  const entry: NavigationStatus = { url: normalizeNavigationURL(url), statusCode };
  const history = (tabNavigationStatus.get(tabId) ?? [])
    .filter((e) => e.url !== entry.url)
    .concat(entry)
    .slice(-MAX_TRACKED_NAVIGATIONS);
  tabNavigationStatus.set(tabId, history);
  if (sessionStore) {
    void sessionStore.set({ [navigationStatusKey(tabId)]: history }).catch(() => {});
  }
}

function forgetNavigationStatus(tabId: number) {
  tabNavigationStatus.delete(tabId);
  if (sessionStore) {
    void sessionStore.remove(navigationStatusKey(tabId)).catch(() => {});
  }
}

async function readNavigationStatus(tabId: number): Promise<NavigationStatus[]> {
  const cached = tabNavigationStatus.get(tabId);
  if (cached || !sessionStore) {
    return cached ?? [];
  }
  try {
    const key = navigationStatusKey(tabId);
    return ((await sessionStore.get(key))[key] as NavigationStatus[]) ?? [];
  } catch (_) {
    return [];
  }
}

// Returns the unsuccessful status `url` was served with in `tabId`, or
// undefined. An unknown URL is not an error: cached responses, non-HTTP pages
// and history navigations never reach the network.
async function isErrorPage(tabId: number, url: string): Promise<number | undefined> {
  const history = await readNavigationStatus(tabId);
  const wanted = normalizeNavigationURL(url);
  return history.find((e) => e.url === wanted && isErrorStatus(e.statusCode))?.statusCode;
}

// `chrome.webRequest` is undefined when the permission is not granted; in that
// case the status stays unknown and documents are submitted as before.
chrome.webRequest?.onHeadersReceived.addListener(
  (details) => {
    if (details.tabId < 0) return;
    recordNavigationStatus(details.tabId, details.url, details.statusCode);
  },
  { urls: ['<all_urls>'], types: ['main_frame'] },
);

// --- Indexing rules cache ---

interface IndexingRules {
  allow: (RegExp | null)[];
  skip: (RegExp | null)[];
}

interface IndexingRulesCache {
  rules: IndexingRules;
  timestamp: number;
}

const INDEXING_RULES_TTL = 60_000;
let indexingRulesCache: IndexingRulesCache | null = null;

function compilePatterns(patterns: string[]): (RegExp | null)[] {
  return patterns.map((pattern) => {
    try {
      return new RegExp(pattern, 'u');
    } catch (_) {
      // Go patterns unsupported by JavaScript are evaluated by the server.
      return null;
    }
  });
}

function isURLSkipped(url: string, rules: IndexingRules): boolean {
  return (
    (rules.allow.length > 0 && !rules.allow.some((re) => re === null || re.test(url))) ||
    rules.skip.some((re) => re?.test(url))
  );
}

function isPageSkipped(url: string, sourceURL: string, rules: IndexingRules): boolean {
  return isURLSkipped(url, rules) || isURLSkipped(sourceURL.split('#', 1)[0], rules);
}

async function getIndexingRules(
  serverURL: string,
  customHeaders: CustomHeader[],
): Promise<IndexingRules> {
  const now = Date.now();
  if (indexingRulesCache && now - indexingRulesCache.timestamp < INDEXING_RULES_TTL) {
    return indexingRulesCache.rules;
  }
  try {
    const u = baseServerURL(serverURL);
    const r = await fetchAPI(u + 'api/rules', { customHeaders });
    if (!r.ok) return cachedRules();
    const data = await r.json();
    const rules = {
      allow: compilePatterns(data.allow ?? []),
      skip: compilePatterns(data.skip ?? []),
    };
    indexingRulesCache = { rules, timestamp: now };
    return rules;
  } catch (_) {
    return cachedRules();
  }
}

function getCustomHeaders(data: StoredSettings): CustomHeader[] {
  const customHeaders: CustomHeader[] = Array.isArray(data['histerCustomHeaders'])
    ? [...data['histerCustomHeaders']]
    : [];
  if (data['histerToken']) {
    customHeaders.push({ name: 'X-Access-Token', value: data['histerToken'] });
  }
  return customHeaders;
}

function getDocumentSubmissionHeaders(data: StoredSettings): CustomHeader[] {
  const customHeaders = getCustomHeaders(data);
  if (data['submitPublicDocuments'] === true) {
    customHeaders.push({ name: 'X-Hister-Public', value: '1' });
  }
  return customHeaders;
}

async function saveSkipRule(
  baseURL: string,
  customHeaders: CustomHeader[],
  pattern: string,
  deleteQuery?: string,
): Promise<void> {
  const rulesResp = await fetchAPI(baseURL + 'api/rules', { customHeaders });
  if (!rulesResp.ok) {
    throw new Error(`Failed to fetch rules: ${rulesResp.status}`);
  }
  const rulesData = await rulesResp.json();
  const existingSkip: string[] = rulesData.skip ?? [];
  const existingPriority: string[] = rulesData.priority ?? [];
  const existingVersioning: string[] = rulesData.versioning ?? [];
  const newSkip = [...existingSkip, pattern];
  const saveResp = await fetchAPI(baseURL + 'api/rules', {
    formData: {
      skip: newSkip.join(' '),
      priority: existingPriority.join(' '),
      versioning: existingVersioning.join(' '),
    },
    customHeaders,
  });
  if (!saveResp.ok) {
    throw new Error(`Failed to save rule: ${saveResp.status}`);
  }
  if (deleteQuery) {
    const deleteResp = await fetchAPI(baseURL + 'api/delete', {
      body: { query: deleteQuery },
      customHeaders,
    });
    if (!deleteResp.ok) {
      throw new Error(`Failed to delete documents: ${deleteResp.status}`);
    }
  }
  indexingRulesCache = null;
}

// --- Tab icon state ---

async function updateTabIcon(tabId: number, url: string): Promise<void> {
  if (!isNormalPageURL(url)) {
    return;
  }
  const data = await chrome.storage.local.get<StoredSettings>([
    'histerURL',
    'histerToken',
    'indexingEnabled',
    'histerCustomHeaders',
    'showIndexedBadge',
  ]);

  const serverURL: string = data['histerURL'] || '';
  const showIndexedBadge: boolean = data['showIndexedBadge'] === true;
  const customHeaders = getCustomHeaders(data);

  const pageURL = await getTabPageURL(tabId, url);
  if (data['indexingEnabled'] === false) {
    await setGreyIcon(tabId);
    await showPreviousIndex(tabId, pageURL, serverURL, customHeaders, showIndexedBadge);
    return;
  }

  if (!serverURL) {
    setNormalIcon(tabId);
    return;
  }

  const rules = await getIndexingRules(serverURL, customHeaders);
  if (isPageSkipped(pageURL, url, rules)) {
    await setGreyIcon(tabId);
  } else {
    setNormalIcon(tabId);
    clearBadge(tabId);
    await showPreviousIndex(tabId, pageURL, serverURL, customHeaders, showIndexedBadge);
  }
}

// --- PDF tab indexing ---

function isPDFUrl(url: string): boolean {
  try {
    const pathname = new URL(url).pathname.toLowerCase();
    return pathname.endsWith('.pdf');
  } catch (_) {
    return false;
  }
}

async function indexPDFTab(
  tabId: number,
  tab: chrome.tabs.Tab,
  ignoreSkipRules = false,
): Promise<IndexingResult> {
  const data = await chrome.storage.local.get<StoredSettings>([
    'histerURL',
    'histerToken',
    'indexingEnabled',
    'histerCustomHeaders',
    'histerLabel',
    'showIndexedBadge',
    'submitPublicDocuments',
  ]);

  if (data['indexingEnabled'] === false && !ignoreSkipRules) return { status: 'disabled' };

  const serverURL: string = data['histerURL'] || '';
  if (!serverURL) return missingURLMsg;

  const showIndexedBadge: boolean = data['showIndexedBadge'] === true;

  const customHeaders = getDocumentSubmissionHeaders(data);

  const rules = await getIndexingRules(serverURL, customHeaders);
  if (!ignoreSkipRules && isURLSkipped(tab.url!, rules)) {
    await setGreyIcon(tabId);
    return { status: 'ok', status_code: 406 };
  }

  const u = baseServerURL(serverURL);

  try {
    const response = await fetch(tab.url!, { credentials: 'include' });
    if (!response.ok) {
      setErrorBadge(tabId);
      return { error: `Failed to fetch PDF: ${response.status}` };
    }
    const pdfBase64 = encodePDF(await response.arrayBuffer());
    const doc = pdfDocument(tab, ignoreSkipRules, data['histerLabel']);

    const r = await sendPDFData(u + 'api/add_pdf', doc, pdfBase64, customHeaders);
    applyIndexStatus(tabId, tab.url!, r.status, showIndexedBadge);
    return { status: 'ok', status_code: r.status };
  } catch (err) {
    setErrorBadge(tabId);
    return indexingError(err);
  }
}

// --- Tab listeners ---

chrome.tabs.onActivated.addListener(async ({ tabId }) => {
  try {
    const tab = await chrome.tabs.get(tabId);
    if (tab.url) await updateTabIcon(tabId, tab.url);
  } catch (_) {}
});

chrome.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete' && tab.url) {
    await updateTabIcon(tabId, tab.url);
    if (isPDFUrl(tab.url)) {
      await indexPDFTab(tabId, tab);
    }
  }
});

chrome.storage.onChanged.addListener(async (changes, area) => {
  if (area !== 'local') return;
  const connectionChanged = [
    'histerURL',
    'histerToken',
    'histerCookies',
    'histerCustomHeaders',
  ].some((key) => key in changes);
  if (!changesAffectIcon(changes, connectionChanged)) return;
  if (connectionChanged) indexingRulesCache = null;
  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    if (tab?.id && tab.url) await updateTabIcon(tab.id, tab.url);
  } catch (_) {}
});

async function isUrlPreviouslyIndexed(
  url: string,
  serverURL: string,
  customHeaders: CustomHeader[],
): Promise<boolean> {
  try {
    const base = serverURL.endsWith('/') ? serverURL : serverURL + '/';
    const r = await fetchAPI(`${base}api/document?url=${encodeURIComponent(url)}`, {
      method: 'HEAD',
      customHeaders,
    });
    return r.status === 200;
  } catch (_) {
    return false;
  }
}

async function indexCurrentTab(): Promise<IndexingResult> {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id) return { error: 'No active tab found' };

  if (tab.url && isPDFUrl(tab.url)) {
    return indexPDFTab(tab.id, tab, true);
  }

  return new Promise((resolve) => {
    chrome.tabs.sendMessage(tab.id!, { action: 'reindex' }, (response) => {
      const error = chrome.runtime.lastError;
      if (error || response?.status_code !== 201) {
        setErrorBadge(tab.id!);
        resolve(
          error ? { error: error.message } : (response ?? { error: 'No response from page' }),
        );
        return;
      }
      setPreviouslyIndexedBadge(tab.id!);
      setTimeout(() => clearBadge(tab.id!), 2500);
      resolve(response);
    });
  });
}

async function disableIndexingForCurrentTab(type: SkipRuleType): Promise<void> {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.id || !tab.url) return;

  try {
    const data = await chrome.storage.local.get<StoredSettings>([
      'histerURL',
      'histerToken',
      'histerCustomHeaders',
    ]);
    let serverURL: string = data['histerURL'] || '';
    if (!serverURL) {
      setErrorBadge(tab.id);
      return;
    }
    if (!serverURL.endsWith('/')) {
      serverURL += '/';
    }
    const pageURL = await getTabPageURL(tab.id, tab.url);
    const pattern = type === 'url' ? buildUrlSkipPattern(pageURL) : buildDomainSkipPattern(pageURL);
    await saveSkipRule(serverURL, getCustomHeaders(data), pattern);
    await setGreyIcon(tab.id);
    setPreviouslyIndexedBadge(tab.id);
    setTimeout(() => clearBadge(tab.id!), 2500);
  } catch (_) {
    setErrorBadge(tab.id);
  }
}

chrome.commands?.onCommand?.addListener((command) => {
  if (command === 'index-current-page') {
    void indexCurrentTab();
  } else if (command === 'disable-indexing-current-page') {
    void disableIndexingForCurrentTab('url');
  } else if (command === 'disable-indexing-current-domain') {
    void disableIndexingForCurrentTab('domain');
  }
});

// --- Document submission ---

async function submitPageData(
  baseURL: string,
  request: BackgroundRequest,
  sender: chrome.runtime.MessageSender,
  data: StoredSettings,
): Promise<IndexingResult> {
  const force = request.action === 'reindex';
  const tabId = sender.tab?.id;
  if (!force) {
    // The status belongs to the URL that was navigated to, which is not
    // necessarily the submitted URL: a page can declare a canonical URL.
    const sourceURL = sender.url ?? sender.tab?.url ?? request.pageData!.url;
    if (tabId !== undefined) {
      const statusCode = await isErrorPage(tabId, sourceURL);
      if (statusCode !== undefined) {
        return { status: 'http_error', status_code: statusCode };
      }
    }
    const rules = await getIndexingRules(baseURL, getCustomHeaders(data));
    if (isPageSkipped(request.pageData!.url, sourceURL, rules)) {
      if (tabId !== undefined) await setGreyIcon(tabId);
      return { status: 'ok', status_code: 406 };
    }
  }
  const labelData = await chrome.storage.local.get<StoredSettings>(['histerLabel']);
  const pageData = { ...request.pageData! };
  if (force) {
    pageData.metadata = { ...pageData.metadata, ignore_skip_rules: true };
  }
  if (labelData['histerLabel']) {
    pageData.label = labelData['histerLabel'];
  }
  const r = await sendPageData(baseURL + 'api/add', pageData, getDocumentSubmissionHeaders(data));
  if (tabId !== undefined) {
    applyIndexStatus(tabId, sender.tab?.url ?? '', r.status, data['showIndexedBadge'] === true);
  }
  return { status: 'ok', status_code: r.status };
}

// --- Message handler ---

// TODO check source
function cjsMsgHandler(
  request: BackgroundRequest,
  sender: chrome.runtime.MessageSender,
  sendResponse: (response: unknown) => void,
) {
  if (request.action === 'indexCurrentPage') {
    indexCurrentTab()
      .then(sendResponse)
      .catch((error) => sendResponse({ error: error.message }));
    return true;
  }
  chrome.storage.local
    .get<StoredSettings>([
      'histerURL',
      'histerToken',
      'indexingEnabled',
      'histerCustomHeaders',
      'showIndexedBadge',
      'submitPublicDocuments',
    ])
    .then((data) => {
      let u = data['histerURL'] || '';
      const indexingEnabled = data['indexingEnabled'] !== false;
      const customHeaders = getCustomHeaders(data);

      if (handleControlMessage(request, sendResponse, u, customHeaders)) return;

      if (!u) {
        chrome.tabs.sendMessage(sender.tab!.id!, missingURLMsg);
        setErrorBadge(sender.tab!.id!);
        return;
      }
      if (!u.endsWith('/')) {
        u += '/';
      }
      if (request.pageData) {
        if (!indexingEnabled && request.action != 'reindex') {
          sendResponse({ status: 'disabled' });
          return;
        }
        submitPageData(u, request, sender, data)
          .then(sendResponse)
          .catch((err) => {
            if (sender.tab?.id !== undefined) setErrorBadge(sender.tab.id);
            sendResponse({ error: err.message });
          });
        return true;
      }
      if (request.resultData) {
        sendResult(u + 'api/history', request.resultData, customHeaders)
          .then((r) => {
            if (r.status === 201) {
              clearBadge(sender.tab!.id!);
            } else if (r.status != 406) {
              setErrorBadge(sender.tab!.id!);
            }
            sendResponse({ status: 'ok', status_code: r.status });
          })
          .catch((err) => {
            setErrorBadge(sender.tab!.id!);
            sendResponse({ error: err.message });
          });
        return true;
      }
    })
    .catch((error) => {
      chrome.tabs.sendMessage(sender.tab!.id!, missingURLMsg);
      setErrorBadge(sender.tab!.id!);
    });
  return true;
}

chrome.runtime.onMessage.addListener(cjsMsgHandler);

function cachedRules(): IndexingRules {
  return indexingRulesCache?.rules ?? { allow: [], skip: [] };
}

function isNormalPageURL(url: string): boolean {
  return (
    !!url &&
    !['chrome://', 'about:', 'moz-extension://', 'chrome-extension://'].some((prefix) =>
      url.startsWith(prefix),
    )
  );
}

async function showPreviousIndex(
  tabId: number,
  url: string,
  serverURL: string,
  customHeaders: CustomHeader[],
  enabled: boolean,
) {
  if (!enabled || !serverURL) return;
  if (await isUrlPreviouslyIndexed(url, serverURL, customHeaders)) setPreviouslyIndexedBadge(tabId);
}
function applyIndexStatus(tabId: number, url: string, status: number, showIndexedBadge: boolean) {
  if (status === 201) {
    setNormalIcon(tabId);
    if (showIndexedBadge) setPreviouslyIndexedBadge(tabId);
    else clearBadge(tabId);
  } else if (status === 406) {
    indexingRulesCache = null;
    setGreyIcon(tabId);
  } else if (status === 422) {
    tabSensitiveState.set(tabId, url);
    setGreyIcon(tabId);
  } else setErrorBadge(tabId);
}
function pdfDocument(
  tab: chrome.tabs.Tab,
  ignoreSkipRules: boolean,
  label: string | undefined,
): Record<string, unknown> {
  const doc: Record<string, unknown> = { url: tab.url!, title: tab.title || tab.url! };
  if (ignoreSkipRules) doc.metadata = { ignore_skip_rules: true };
  if (label) doc.label = label;
  return doc;
}
function encodePDF(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary);
}
function indexingError(err: unknown): IndexingResult {
  return { error: err instanceof Error ? err.message : String(err) };
}

function baseServerURL(url: string): string {
  return url.endsWith('/') ? url : url + '/';
}

function handleControlMessage(
  request: BackgroundRequest,
  sendResponse: (response: unknown) => void,
  u: string,
  customHeaders: CustomHeader[],
): boolean {
  if (request.action === 'getTabState') {
    const stored = tabSensitiveState.get(request.tabId as number);
    sendResponse({ isSensitive: stored !== undefined && stored === request.url });
    return true;
  }

  if (request.action === 'addSkipRule') {
    if (!u) {
      sendResponse({ error: 'No server URL configured' });
      return true;
    }
    const baseURL = u.endsWith('/') ? u : u + '/';
    (async () => {
      try {
        await saveSkipRule(baseURL, customHeaders, request.pattern, request.deleteQuery);
        sendResponse({ ok: true });
        // Grey out the icon on the active tab immediately
        const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
        if (tab?.id && tab.url) await updateTabIcon(tab.id, tab.url);
      } catch (e) {
        sendResponse({ error: (e as Error).message });
      }
    })();
    return true;
  }

  if (request.action === 'checkSkipRule') {
    if (!u) {
      sendResponse({ isSkipped: false });
      return true;
    }
    const baseURL = u.endsWith('/') ? u : u + '/';
    getIndexingRules(baseURL, customHeaders).then((rules) => {
      sendResponse({
        isSkipped: isPageSkipped(request.url, request.sourceURL ?? request.url, rules),
      });
    });
    return true;
  }

  return false;
}

function changesAffectIcon(changes: Record<string, unknown>, connectionChanged: boolean): boolean {
  return connectionChanged || 'indexingEnabled' in changes || 'showIndexedBadge' in changes;
}
