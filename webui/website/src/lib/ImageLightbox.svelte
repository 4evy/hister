<script lang="ts">
  import X from '@lucide/svelte/icons/x';
  import { modal } from '#lib/modal.js';

  let { contentKey }: { contentKey?: string } = $props();

  let openSrc = $state<string | null>(null);
  let openAlt = $state('');

  function close() {
    openSrc = null;
  }

  $effect(() => {
    void contentKey;
    close();

    const images = Array.from(document.querySelectorAll<HTMLImageElement>('.content img')).filter(
      (img) => !img.closest('a, button'),
    );
    const previews = images.map((img) => {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'image-preview-trigger';
      button.setAttribute('aria-label', `Open image preview: ${img.alt || 'image'}`);
      button.setAttribute('aria-haspopup', 'dialog');
      img.replaceWith(button);
      button.append(img);
      button.onclick = () => {
        openSrc = img.src;
        openAlt = img.alt;
      };
      return { img, button };
    });

    return () => {
      for (const { img, button } of previews) {
        button.replaceWith(img);
      }
    };
  });
</script>

{#if openSrc}
  <dialog
    use:modal
    onclose={close}
    aria-label="Image preview"
    class="m-auto max-h-[calc(100dvh-2rem)] max-w-[calc(100dvw-2rem)] overflow-visible border-0 bg-transparent p-0 md:max-h-[calc(100dvh-5rem)] md:max-w-[calc(100dvw-5rem)]"
  >
    <img
      src={openSrc}
      alt={openAlt}
      class="block max-h-[calc(100dvh-2rem)] max-w-[calc(100dvw-2rem)] border-[3px] border-white object-contain shadow-[8px_8px_0_rgba(255,255,255,0.15)] md:max-h-[calc(100dvh-5rem)] md:max-w-[calc(100dvw-5rem)]"
    />
    <button
      onclick={close}
      aria-label="Close image preview"
      class="absolute top-2 right-2 flex size-11 cursor-pointer items-center justify-center border-[3px] border-white bg-white text-black shadow-[3px_3px_0_rgba(255,255,255,0.2)] transition-colors hover:bg-gray-100"
    >
      <X size={20} />
    </button>
  </dialog>
{/if}

<style>
  dialog::backdrop {
    background: rgb(0 0 0 / 80%);
  }
</style>
