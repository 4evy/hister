import { untrack } from 'svelte';
import { pushPreviewHistory, replacePreviewHistory, withSkipUrl } from './preview.js';

interface PreviewHistoryState {
  type?: string;
  id?: string;
  documentId?: string;
  title?: string;
  versionId?: number | null;
}

export class PreviewState {
  url = $state('');
  documentId = $state('');
  title = $state('');
  version = $state<number | null>(null);
  fullscreen = $state(false);
  panelOpen = $state(true);

  constructor(
    private storageKey: string,
    panelOpen = true,
  ) {
    this.panelOpen = panelOpen;
  }

  open(
    url: string,
    title: string,
    documentId: string,
    desktop: boolean,
    skipUrl: { value: boolean },
  ) {
    this.url = url;
    this.title = title;
    this.documentId = documentId;
    this.version = null;
    if (desktop) {
      if (!this.panelOpen) {
        this.panelOpen = true;
        localStorage.setItem(this.storageKey, 'true');
      }
      return;
    }
    this.enterFullscreen(skipUrl);
  }

  restore(state: PreviewHistoryState | null): boolean {
    if (state?.type !== 'preview') return false;
    this.restoreDocument(state);
    this.panelOpen = true;
    this.fullscreen = true;
    this.version = state.versionId ?? null;
    return true;
  }

  private restoreDocument(state: PreviewHistoryState) {
    this.url = state.id || '';
    this.documentId = state.documentId || '';
    this.title = state.title || '';
  }

  visibleOnDesktop(desktop: boolean): boolean {
    return desktop && (this.panelOpen || this.fullscreen);
  }

  syncDocument(url: string, title: string, documentId: string, skipUrl: { value: boolean }) {
    if (url === untrack(() => this.url) && documentId === untrack(() => this.documentId)) return;
    this.version = null;
    this.documentId = documentId;
    this.title = title;
    this.url = url;
    if (this.fullscreen)
      withSkipUrl(skipUrl, () => replacePreviewHistory(url, title, null, documentId));
  }

  clearDocument() {
    this.url = '';
    this.documentId = '';
    this.title = '';
  }

  toggle(
    desktop: boolean,
    skipUrl: { value: boolean },
    restoreHistory: () => void,
    openSelected: () => void,
  ) {
    if (desktop) this.toggleDesktop(skipUrl, restoreHistory, openSelected);
    else if (this.fullscreen) this.close(skipUrl, restoreHistory);
    else openSelected();
  }

  private toggleDesktop(
    skipUrl: { value: boolean },
    restoreHistory: () => void,
    openSelected: () => void,
  ) {
    if (this.fullscreen) this.exitFullscreen(skipUrl, restoreHistory);
    else if (this.panelOpen) this.enterFullscreen(skipUrl);
    else openSelected();
  }

  enterFullscreen(skipUrl: { value: boolean }) {
    this.fullscreen = true;
    withSkipUrl(skipUrl, () =>
      pushPreviewHistory(this.url, this.title, this.version, this.documentId),
    );
  }

  exitFullscreen(skipUrl: { value: boolean }, restoreHistory: () => void) {
    this.fullscreen = false;
    withSkipUrl(skipUrl, restoreHistory);
  }

  closePanel() {
    this.panelOpen = false;
    localStorage.setItem(this.storageKey, 'false');
  }

  close(skipUrl: { value: boolean }, restoreHistory: () => void) {
    this.closePanel();
    this.exitFullscreen(skipUrl, restoreHistory);
  }
}
