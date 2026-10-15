<script lang="ts">
  import type { ComponentProps } from 'svelte';
  import PreviewPanel from './PreviewPanel.svelte';
  import { replacePreviewHistory, withSkipUrl } from '#lib/preview.js';

  type Props = Pick<
    ComponentProps<typeof PreviewPanel>,
    'url' | 'documentId' | 'hintTitle' | 'onclose' | 'onfullscreentoggle'
  > & {
    viewingVersion: number | null;
    skipUrl: { value: boolean };
  };

  let { viewingVersion = $bindable(), skipUrl, ...props }: Props = $props();

  function changeVersion(id: number | null) {
    viewingVersion = id;
    withSkipUrl(skipUrl, () =>
      replacePreviewHistory(props.url, props.hintTitle ?? '', id, props.documentId),
    );
  }
</script>

<PreviewPanel
  {...props}
  fullscreen
  initialViewingVersionId={viewingVersion}
  onviewingversionchange={changeVersion}
/>
