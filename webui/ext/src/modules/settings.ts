export interface StoredSettings {
  histerURL?: string;
  histerToken?: string;
  histerCookies?: string;
  histerCustomHeaders?: { name: string; value: string }[];
  histerLabel?: string;
  histerProfileUserID?: number;
  indexingEnabled?: boolean;
  showIndexedBadge?: boolean;
  submitPublicDocuments?: boolean;
}

const DEFAULT_SERVER_URL = 'http://127.0.0.1:4433/';

async function ensureDefaultServerURL(): Promise<string> {
  const data = await chrome.storage.local.get<StoredSettings>(['histerURL']);
  if (data['histerURL']) {
    return data['histerURL'];
  }
  await chrome.storage.local.set({ histerURL: DEFAULT_SERVER_URL });
  return DEFAULT_SERVER_URL;
}

export { DEFAULT_SERVER_URL, ensureDefaultServerURL };

export function serverBaseURL(url: string): string {
  return url.endsWith('/') ? url : url + '/';
}

export async function profileUserID(response: Response): Promise<number> {
  try {
    const profile = await response.json();
    return Number(profile?.user_id ?? 0);
  } catch {
    return 0;
  }
}

export function connectionSettings(data: StoredSettings) {
  return {
    url: data.histerURL || DEFAULT_SERVER_URL,
    token: data.histerToken || '',
    headers: Array.isArray(data.histerCustomHeaders) ? data.histerCustomHeaders : [],
  };
}
