<script lang="ts">
  import type { StoredSettings } from '../modules/settings';

  import {
    serverBaseURL,
    profileUserID as readProfileUserID,
    connectionSettings as readConnectionSettings,
  } from '../modules/settings';

  import { Button } from '@hister/components/ui/button';
  import { Input } from '@hister/components/ui/input';
  import { Label } from '@hister/components/ui/label';
  import { Switch } from '@hister/components/ui/switch';
  import * as Card from '@hister/components/ui/card';
  import SettingsInput from '../options/SettingsInput.svelte';
  import * as Tooltip from '@hister/components/ui/tooltip';
  import { SkipRuleActions, buildUrlSkipPattern, buildDomainSkipPattern } from '@hister/components';
  import { Settings, Sun, Moon, Save, Info, Check } from '@lucide/svelte';
  import { slide } from 'svelte/transition';
  import { ModeWatcher, toggleMode, mode } from 'mode-watcher';
  import { fetchAPI, syncServerCookies } from '../modules/network';
  import { DEFAULT_SERVER_URL } from '../modules/settings';
  import { getTabPageURL } from '../modules/tabs';

  let url = $state(DEFAULT_SERVER_URL);
  let accessToken = $state('');
  let customHeaders: { name: string; value: string }[] = $state([]);
  let indexingEnabled = $state(true);
  let showIndexedBadge = $state(false);
  let submitPublicDocuments = $state(false);
  let profileUserID = $state(0);
  let message = $state('');
  let messageType: 'success' | 'error' | 'info' = $state('success');
  let showSettings = $state(false);
  let isPageSkipped = $state(false);
  let tabURL = $state('');
  let sourceTabURL = $state('');
  let messageKey = $state(0); // to reappear message every time it is updated
  let pageLabel = $state('');

  function setMessage(mType: 'success' | 'error' | 'info', msg: string) {
    message = msg;
    messageType = mType;
    messageKey++;
  }

  function setErrorMessage(msg: string) {
    setMessage('error', msg);
  }

  function setInfoMessage(msg: string) {
    setMessage('info', msg);
  }

  function setSuccessMessage(msg: string) {
    setMessage('success', msg);
  }

  let authCheckPassed = $state<boolean | null>(null);

  function isAuthenticated(userID: number): boolean {
    return userID > 0;
  }

  function setProfileUserID(userID: number) {
    profileUserID = userID;
    chrome.storage.local.set({ histerProfileUserID: userID });
    if (!isAuthenticated(userID) && submitPublicDocuments) {
      submitPublicDocuments = false;
      chrome.storage.local.set({ submitPublicDocuments: false });
    }
  }

  function checkAuth(serverURL: string, token = accessToken): Promise<boolean> {
    const authURL = serverBaseURL(serverURL);
    return fetchAPI(authURL + 'api/profile', { customHeaders, accessToken: token })
      .then(async (r) => {
        if (!r.ok) {
          authCheckPassed = false;
          setProfileUserID(0);
          return false;
        }
        setProfileUserID(await readProfileUserID(r));
        authCheckPassed = true;
        return authCheckPassed;
      })
      .catch(() => {
        authCheckPassed = false;
        setProfileUserID(0);
        return false;
      });
  }

  chrome.storage.local.get<StoredSettings>(
    [
      'histerURL',
      'histerToken',
      'histerCustomHeaders',
      'indexingEnabled',
      'histerLabel',
      'showIndexedBadge',
      'submitPublicDocuments',
      'histerProfileUserID',
    ],
    (data) => {
      if (!data['histerURL']) {
        chrome.storage.local.set({ histerURL: DEFAULT_SERVER_URL });
      }
      const settings = readConnectionSettings(data);
      url = settings.url;
      accessToken = settings.token;
      customHeaders = settings.headers;
      indexingEnabled = data['indexingEnabled'] !== false;
      showIndexedBadge = data['showIndexedBadge'] === true;
      submitPublicDocuments = data['submitPublicDocuments'] === true;
      profileUserID = Number(data['histerProfileUserID'] ?? 0);
      pageLabel = data['histerLabel'] || '';

      checkAuth(url);

      chrome.tabs.query({ active: true, currentWindow: true }, async (tabs) => {
        if (!tabs?.length) return;
        const tab = tabs[0];
        chrome.action.getBadgeText({ tabId: tab.id! }, (badgeText) => {
          if (badgeText === '!') {
            setErrorMessage('Failed to send page data to server');
          }
        });
        const currentTabURL = tab.url;
        if (currentTabURL) {
          sourceTabURL = currentTabURL;
          tabURL = await getTabPageURL(tab.id!, currentTabURL);
          checkTabSkipRule(tabURL);
          chrome.runtime.sendMessage(
            { action: 'getTabState', tabId: tab.id, url: currentTabURL },
            (resp) => {
              if (resp?.isSensitive) {
                setInfoMessage('Page not indexed: it contains sensitive data');
              }
            },
          );
        }
      });
    },
  );

  async function handleAddSkipRule(type: 'url' | 'domain', deleteMatches: boolean) {
    const { pattern, deleteQuery } = skipRuleRequest(type, deleteMatches);
    try {
      const response = await chrome.runtime.sendMessage({
        action: 'addSkipRule',
        pattern,
        deleteQuery,
      });
      if (response?.ok) {
        await checkTabSkipRule(tabURL);
      } else {
        skipRuleFailed(response);
      }
    } catch (e: any) {
      skipRuleFailed({ error: e.message });
    }
  }

  async function checkTabSkipRule(tabURL: string) {
    try {
      const response = await chrome.runtime.sendMessage({
        action: 'checkSkipRule',
        url: tabURL,
        sourceURL: sourceTabURL,
      });
      isPageSkipped = response?.isSkipped === true;
    } catch (_) {}
  }

  async function save(e: Event) {
    e.preventDefault();

    const verifyURL = serverBaseURL(url);

    const tokenToSave = accessToken.trim();
    try {
      const response = await fetchAPI(verifyURL + 'api/config', {
        customHeaders,
        accessToken: tokenToSave,
      });
      if (response.status !== 200) {
        setErrorMessage(`Server returned status ${response.status}`);
        return;
      }
      if (!(await validServerResponse(response))) return;
      await chrome.storage.local.set({
        histerURL: url,
        histerToken: tokenToSave,
        histerCustomHeaders: $state.snapshot(customHeaders),
        submitPublicDocuments: canSubmitPublicDocuments(),
      });
      accessToken = tokenToSave;
      const authenticated = await checkAuth(url, tokenToSave);
      savedSettingsMessage(tokenToSave, authenticated);

      chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
        if (tabs?.length) {
          chrome.action.setBadgeText({ text: '', tabId: tabs[0].id! });
        }
      });
    } catch (error) {
      setErrorMessage((error as Error).message);
    }
  }

  function toggleIndexing() {
    chrome.storage.local.set({ indexingEnabled: indexingEnabled });
    setSuccessMessage(`Automatic indexing ${indexingEnabled ? 'enabled' : 'disabled'}`);
  }

  function toggleShowIndexedBadge() {
    chrome.storage.local.set({ showIndexedBadge: showIndexedBadge });
  }

  function toggleSubmitPublicDocuments() {
    chrome.storage.local.set({
      submitPublicDocuments: canSubmitPublicDocuments(),
    });
  }

  async function authenticate() {
    const authURL = serverBaseURL(url);
    try {
      const cookieHeader = await syncServerCookies(authURL);
      if (!cookieHeader) {
        setErrorMessage(
          'No cookies found for server URL. Make sure you are logged in to the Hister web app.',
        );
        return;
      }
      if (await checkAuth(url)) {
        setSuccessMessage('Authentication successful');
      } else {
        setErrorMessage(
          'Authentication failed. Make sure you are logged in to the Hister web app.',
        );
      }
    } catch (error) {
      setErrorMessage(cookieErrorMessage(error));
    }
  }

  function reindex() {
    chrome.runtime.sendMessage({ action: 'indexCurrentPage' }, (r) => {
      const error = chrome.runtime.lastError;
      if (r?.status === 'ok' && r.status_code === 201) {
        setSuccessMessage('Index successful');
        return;
      }
      const msg = reindexErrorMessage(r, error);
      setErrorMessage(msg);
    });
  }

  function saveLabel() {
    chrome.storage.local.set({ histerLabel: pageLabel }, () => {
      setSuccessMessage('Label saved');
      //reindex();
    });
  }

  async function applyLabel() {
    const serverURL = url.endsWith('/') ? url.slice(0, -1) : url;
    try {
      const res = await fetchAPI(`${serverURL}/api/label`, {
        body: { url: tabURL, label: pageLabel },
        customHeaders,
      });
      if (res.ok) {
        pageLabel = '';
        setSuccessMessage('Label applied to this page');
      } else {
        setErrorMessage('Failed to apply label');
      }
    } catch (e: unknown) {
      labelFailed(e);
    }
  }

  function toggleSettings() {
    showSettings = !showSettings;
    message = '';
  }
  function cookieErrorMessage(error: unknown) {
    return (error as Error).message ?? 'Failed to read browser cookies.';
  }
  function skipRuleRequest(type: 'url' | 'domain', deleteMatches: boolean) {
    const pattern = type === 'url' ? buildUrlSkipPattern(tabURL) : buildDomainSkipPattern(tabURL);
    const deleteQuery = deleteMatches
      ? type === 'url'
        ? `url:"${tabURL.replaceAll('"', '\\"')}"`
        : `domain:${new URL(tabURL).hostname}`
      : undefined;
    return { pattern, deleteQuery };
  }
  function skipRuleFailed(response: { error?: string } | undefined) {
    setErrorMessage(response?.error ?? 'Failed to add skip rule');
  }
  async function validServerResponse(response: Response) {
    try {
      await response.json();
      return true;
    } catch {
      setErrorMessage('Server response is not valid JSON - probably invalid server URL.');
      return false;
    }
  }
  function savedSettingsMessage(token: string, authenticated: boolean) {
    if (token && !authenticated)
      setErrorMessage('Settings saved, but the access token was rejected.');
    else setSuccessMessage('Settings saved');
  }
  function canSubmitPublicDocuments() {
    return isAuthenticated(profileUserID) && submitPublicDocuments;
  }
  function reindexErrorMessage(
    r: { error?: string; status_code?: number } | undefined,
    error: chrome.runtime.LastError | undefined,
  ) {
    let msg = 'Reindex failed';
    if (error) {
      msg += ': ' + error.message;
    } else if (r?.error) {
      msg += ': ' + r.error;
    }
    msg += reindexStatusMessage(r);
    return msg;
  }
  function reindexStatusMessage(r: { status_code?: number } | undefined) {
    return r?.status_code === 403 ? ': Unauthorized' : '';
  }
  function labelFailed(error: unknown) {
    setErrorMessage((error as Error).message ?? 'Failed to apply label');
  }
</script>

{#snippet connectionSettings()}
  <Card.Root
    class="border-brutal-border gap-0 rounded-none border-0 border-b-[3px] py-0 shadow-none"
  >
    <Card.Content class="space-y-4 p-5">
      <form onsubmit={save} class="space-y-4">
        <SettingsInput label="Server URL" bind:value={url} placeholder="Server URL..." />

        <SettingsInput
          label="Access Token"
          bind:value={accessToken}
          type="password"
          placeholder="Optional access token..."
          description="Authenticate with a global or personal access token. Leave blank to use browser cookies."
        />

        <Button
          type="submit"
          class="bg-hister-coral border-brutal-border font-outfit h-9 w-full border-[3px] text-sm font-bold tracking-wide text-white shadow-[3px_3px_0_var(--brutal-shadow)] transition-all hover:translate-x-px hover:translate-y-px hover:shadow-[1px_1px_0_var(--brutal-shadow)]"
        >
          Save
        </Button>

        <div class="flex items-center justify-between">
          <Label
            for="show-indexed-badge"
            class="font-outfit text-text-brand cursor-pointer text-sm font-bold"
          >
            Show indicator for indexed pages
          </Label>
          <Switch
            id="show-indexed-badge"
            bind:checked={showIndexedBadge}
            onCheckedChange={toggleShowIndexedBadge}
          />
        </div>

        {#if isAuthenticated(profileUserID)}
          <div class="flex items-center justify-between">
            <Label
              for="submit-public-documents"
              class="font-outfit text-text-brand cursor-pointer text-sm font-bold"
            >
              Submit as public documents
            </Label>
            <Switch
              id="submit-public-documents"
              bind:checked={submitPublicDocuments}
              onCheckedChange={toggleSubmitPublicDocuments}
            />
          </div>
        {/if}

        <div class="flex items-center justify-between">
          <span class="font-outfit text-text-brand text-sm font-bold">Theme</span>
          <button
            type="button"
            onclick={toggleMode}
            class="border-brutal-border hover:border-hister-indigo flex cursor-pointer items-center gap-2 rounded border-[3px] bg-transparent px-3 py-1.5 transition-all"
            aria-label="Toggle theme"
          >
            {#if mode.current === 'light'}
              <Sun size={16} />
              <span class="font-outfit text-text-brand text-sm font-bold">Light</span>
            {:else}
              <Moon size={16} />
              <span class="font-outfit text-text-brand text-sm font-bold">Dark</span>
            {/if}
          </button>
        </div>
      </form>
    </Card.Content>
  </Card.Root>
{/snippet}

<ModeWatcher />

<main class="w-80">
  <!-- Header bar -->
  <header
    class="bg-hister-indigo/90 border-brutal-border flex items-center justify-between border-b-[3px] px-5 py-3"
  >
    <a
      href={url}
      onclick={(e) => {
        e.preventDefault();
        chrome.tabs.create({ url });
      }}
      class="font-outfit cursor-pointer text-lg font-black tracking-widest text-white uppercase hover:underline"
      >Hister</a
    >
    <button
      onclick={toggleSettings}
      class="hover:text-hister-coral cursor-pointer border-0 bg-transparent p-0 text-white transition-colors"
      aria-label="Settings"
    >
      <Settings size={20} />
    </button>
  </header>

  {#if showSettings}
    <!-- Settings View -->
    {@render connectionSettings()}
  {:else}
    <!-- Main View -->
    <!-- Automatic Indexing Toggle -->
    <div class="border-brutal-border border-b-[3px] px-5 py-4">
      <div class="flex items-center justify-between">
        <Label for="indexing" class="font-outfit text-text-brand cursor-pointer text-sm font-bold">
          Automatic indexing
        </Label>
        <Switch id="indexing" bind:checked={indexingEnabled} onCheckedChange={toggleIndexing} />
      </div>
    </div>

    <!-- Reindex section -->
    {#if tabURL}
      <div class="border-brutal-border border-b-[3px] px-5 py-4">
        {#if isPageSkipped}
          <p class="text-text-brand mb-3 text-sm">
            Indexing rules block automatic indexing. You can still index this page manually.
          </p>
        {/if}
        <Button
          variant="outline"
          onclick={reindex}
          class="border-brutal-border font-outfit hover:border-hister-indigo h-9 w-full border-[3px] text-sm font-bold tracking-wide transition-all hover:shadow-[3px_3px_0_var(--brutal-shadow)]"
        >
          Index this page now
        </Button>
      </div>

      <!-- Label section -->
      <div class="border-brutal-border border-b-[3px] px-5 py-4">
        <div class="mb-2 flex items-center gap-1">
          <p class="font-outfit text-text-brand text-xs font-bold tracking-widest">Label</p>
          <Tooltip.Provider delayDuration={0}>
            <Tooltip.Root>
              <Tooltip.Trigger>
                <Info size={16} class="cursor-help" />
              </Tooltip.Trigger>
              <Tooltip.Portal>
                <Tooltip.Content class="max-w-52 text-xs">
                  Search with <span class="font-mono">label:yourtext</span> to filter results.<br />
                  Useful for labeling pages, research sessions or differentiating browser profiles.
                </Tooltip.Content>
              </Tooltip.Portal>
            </Tooltip.Root>
          </Tooltip.Provider>
        </div>
        <div class="flex gap-2">
          <Input
            type="text"
            bind:value={pageLabel}
            placeholder="Add label..."
            class="bg-page-bg border-hister-indigo font-fira text-text-brand placeholder:text-text-brand-muted focus-visible:border-hister-coral h-9 flex-1 border-[3px] px-3 text-sm shadow-none transition-colors focus-visible:ring-0"
          />
          <Tooltip.Provider delayDuration={0}>
            <Tooltip.Root>
              <Tooltip.Trigger>
                {#snippet child({ props })}
                  <Button
                    {...props}
                    variant="outline"
                    onclick={saveLabel}
                    aria-label="Save label"
                    class="border-brutal-border font-outfit hover:border-hister-indigo h-9 border-[3px] px-3 text-sm font-bold tracking-wide transition-all hover:shadow-[3px_3px_0_var(--brutal-shadow)]"
                  >
                    <Save size={16} />
                  </Button>
                {/snippet}
              </Tooltip.Trigger>
              <Tooltip.Portal>
                <Tooltip.Content class="max-w-52 text-xs">
                  Save as default label applied to all pages you index from now on with this profile
                </Tooltip.Content>
              </Tooltip.Portal>
            </Tooltip.Root>
          </Tooltip.Provider>
          <Tooltip.Provider delayDuration={0}>
            <Tooltip.Root>
              <Tooltip.Trigger>
                {#snippet child({ props })}
                  <Button
                    {...props}
                    variant="outline"
                    onclick={applyLabel}
                    aria-label="Apply label to this page"
                    class="border-brutal-border font-outfit hover:border-hister-indigo h-9 border-[3px] px-3 text-sm font-bold tracking-wide transition-all hover:shadow-[3px_3px_0_var(--brutal-shadow)]"
                  >
                    <Check size={16} />
                  </Button>
                {/snippet}
              </Tooltip.Trigger>
              <Tooltip.Portal>
                <Tooltip.Content class="max-w-52 text-xs">
                  Apply label to this page only
                </Tooltip.Content>
              </Tooltip.Portal>
            </Tooltip.Root>
          </Tooltip.Provider>
        </div>
      </div>

      <!-- Disable indexing section -->
      {#if tabURL && !isPageSkipped}
        <SkipRuleActions
          urlLabel="This Page"
          onAddSkipRule={handleAddSkipRule}
          class="border-brutal-border border-b-[3px] px-5 py-4"
        />
      {/if}
    {/if}

    <!-- Authenticate section -->
    {#if authCheckPassed === false}
      <div class="border-brutal-border border-b-[3px] px-5 py-4">
        <Button
          variant="outline"
          onclick={authenticate}
          class="border-brutal-border font-outfit hover:border-hister-indigo h-9 w-full border-[3px] text-sm font-bold tracking-wide transition-all hover:shadow-[3px_3px_0_var(--brutal-shadow)]"
        >
          Authenticate with Browser Session
        </Button>
      </div>
    {/if}
  {/if}
  <!-- Status message -->
  {#if message}
    {#key messageKey}
      <div
        transition:slide
        class="font-inter mx-5 my-4 border-l-[4px] px-4 py-3 text-sm {messageType === 'success'
          ? 'border-l-hister-teal bg-hister-teal/10 text-hister-teal'
          : messageType === 'info'
            ? 'border-l-hister-indigo/60 bg-hister-indigo/10 text-hister-indigo'
            : 'border-l-hister-rose bg-hister-rose/10 text-hister-rose'}"
      >
        {message}
      </div>
    {/key}
  {/if}
</main>
