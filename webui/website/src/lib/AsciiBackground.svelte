<script lang="ts">
  import { onMount } from 'svelte';
  import type { Textmodifier } from 'textmode.js';
  import fontSource from '#lib/assets/drawing-font/garden-mono.woff?url';
  import fallback from '#lib/assets/ascii-fallback.svg?raw';
  import atlas from '#lib/assets/ascii-glyphs.json';
  import { drawingFontSize, drawingSize, initialTime, staticField } from '#lib/ascii-svg.js';
  import { fieldFrame } from '#lib/ascii-field.js';

  let host: HTMLDivElement;
  let renderer: Textmodifier | undefined;
  let initialized = false;
  let ready = $state(false);
  let paused = $state(false);
  let reduced = $state(false);
  let small = $state(false);
  let scrolled = $state(false);
  let stageTime = initialTime;
  let staticImage = $state(fallback);
  let pageTurn = 0;

  function shouldAnimate() {
    return !paused && !reduced && !small && !document.hidden;
  }

  function syncMotion() {
    if (!renderer || !initialized) return;
    if (shouldAnimate()) renderer.loop();
    else renderer.noLoop();
  }

  onMount(() => {
    let disposed = false;
    let starting = false;
    let generation = 0;
    let failed = false;
    let dimensions = '';
    let background = '';
    let colors: string[] = [];
    const preference = matchMedia('(prefers-reduced-motion: reduce)');
    const narrow = matchMedia('(max-width: 800px)');
    const size = () => drawingSize(host.clientWidth, host.clientHeight);
    const scrollChanged = () => {
      scrolled = window.scrollY > window.innerHeight * 0.55;
      if (paused || reduced) return;
      const distance = document.documentElement.scrollHeight - window.innerHeight;
      pageTurn = Math.max(0, Math.min(1, window.scrollY / Math.max(1, distance)));
    };
    const readPalette = () => {
      const style = getComputedStyle(host);
      background = style.getPropertyValue('--brutal-bg').trim();
      colors = ['indigo', 'coral', 'teal', 'amber'].map((name) =>
        style.getPropertyValue(`--hister-${name}`).trim(),
      );
      if (initialized) renderer?.redraw();
    };
    const stop = () => {
      generation++;
      starting = false;
      dimensions = '';
      if (initialized) renderer?.destroy();
      else renderer?.noLoop();
      initialized = false;
      renderer = undefined;
      ready = false;
      if (!disposed) {
        const [width, height] = size();
        staticImage = staticField(width, height, atlas);
      }
    };
    const contextLost = (event: Event) => {
      event.preventDefault();
      failed = true;
      stop();
    };
    function requireWebGL() {
      // Probe before loading the renderer so disabled WebGL keeps the static SVG
      const canvas = document.createElement('canvas');
      const gl = canvas.getContext('webgl2');
      if (!gl) throw new Error('WebGL unavailable');
      gl.getExtension('WEBGL_lose_context')?.loseContext();
    }
    function failedStart(current: number) {
      if (current !== generation) return;
      failed = true;
      stop();
    }
    function finishStart(current: number) {
      if (current === generation) starting = false;
    }
    function canStart() {
      return ![disposed, starting, renderer, small, reduced, failed].some(Boolean);
    }
    async function start() {
      if (!canStart()) return;
      starting = true;
      const current = ++generation;
      try {
        requireWebGL();
        const { textmode } = await import('textmode.js');
        if (current !== generation) return;
        const [width, height] = size();
        const t = textmode.create({
          width,
          height,
          pixelDensity: 1,
          fontSize: drawingFontSize,
          fontSource,
          frameRate: 24,
          loadingScreen: { transition: 'none' },
        });
        renderer = t;
        dimensions = `${width}:${height}`;
        t.canvas.addEventListener(
          'webglcontextlost',
          (event) => {
            if (current === generation) contextLost(event);
          },
          { once: true },
        );
        host.append(t.canvas);
        t.draw(() => {
          if (current !== generation) return;
          t.background(background);
          t.cellColor(background);
          const grid = t.grid!;
          const cells = fieldFrame(grid.cols, grid.rows, stageTime, colors, pageTurn);
          cells.forEach((cell) => {
            t.push();
            t.translate(cell.x - grid.cols / 2 + 0.5, cell.y - grid.rows / 2 + 0.5);
            if (cell.highlighted) t.cellColor(colors[1]);
            t.char(cell.glyph);
            t.charColor(cell.highlighted ? background : cell.color);
            t.rect(1, 1);
            t.pop();
          });
          ready = true;
          if (shouldAnimate()) stageTime += 1 / 24;
        });
        void t.setup(() => {
          if (current !== generation) {
            t.destroy();
            return;
          }
          initialized = true;
          const [width, height] = size();
          dimensions = `${width}:${height}`;
          t.resizeCanvas(width, height);
          syncMotion();
          t.redraw();
        });
      } catch {
        failedStart(current);
      } finally {
        finishStart(current);
      }
    }
    const environmentChanged = () => {
      small = narrow.matches;
      reduced = preference.matches;
      if (small || reduced) stop();
      else void start();
    };
    readPalette();
    scrollChanged();
    environmentChanged();
    const themeObserver = new MutationObserver(readPalette);
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    });
    function updateStaticImage(width: number, height: number) {
      if (!ready) staticImage = staticField(width, height, atlas);
    }
    let resizeFrame = 0;
    const resize = new ResizeObserver(() => {
      cancelAnimationFrame(resizeFrame);
      resizeFrame = requestAnimationFrame(() => {
        const [width, height] = size();
        updateStaticImage(width, height);
        if (!renderer || !initialized) return;
        const next = `${width}:${height}`;
        if (next === dimensions) return;
        dimensions = next;
        renderer.resizeCanvas(width, height);
        renderer.redraw();
      });
    });
    resize.observe(host);
    window.addEventListener('scroll', scrollChanged, { passive: true });
    preference.addEventListener('change', environmentChanged);
    narrow.addEventListener('change', environmentChanged);
    document.addEventListener('visibilitychange', syncMotion);
    return () => {
      disposed = true;
      stop();
      window.removeEventListener('scroll', scrollChanged);
      preference.removeEventListener('change', environmentChanged);
      narrow.removeEventListener('change', environmentChanged);
      document.removeEventListener('visibilitychange', syncMotion);
      themeObserver.disconnect();
      resize.disconnect();
      cancelAnimationFrame(resizeFrame);
    };
  });
</script>

<div class="canvas-host" bind:this={host} class:ready class:scrolled aria-hidden="true">
  {#if !ready}{@html staticImage}{/if}
</div>
{#if ready && !reduced && !small}
  <button
    aria-label={paused ? 'Play ASCII background' : 'Pause ASCII background'}
    aria-pressed={paused}
    onclick={() => {
      paused = !paused;
      syncMotion();
    }}
  >
    {paused ? 'Play background' : 'Pause background'}
  </button>
{/if}

<style>
  .canvas-host {
    position: fixed;
    inset: 0;
    z-index: -1;
    pointer-events: none;
    overflow: hidden;
    opacity: 0.72;
    transition: opacity 400ms;
    mask-image: linear-gradient(90deg, #0001, #0002 45%, #0009 65%, black);
  }
  .canvas-host.scrolled {
    opacity: 0.22;
  }
  @media (prefers-reduced-motion: reduce) {
    .canvas-host {
      transition: none;
    }
  }
  .canvas-host :global(canvas) {
    position: absolute;
    inset: 0;
    width: 100% !important;
    height: 100% !important;
    opacity: 0;
  }
  .canvas-host.ready :global(canvas) {
    opacity: 1;
  }
  .canvas-host :global(svg) {
    width: 100%;
    height: 100%;
  }
  button {
    position: fixed;
    z-index: 20;
    right: 48px;
    bottom: 12px;
    pointer-events: auto;
    cursor: pointer;
    padding: 8px;
    font:
      11px 'Fira Code',
      monospace;
    color: var(--text-secondary);
    background: var(--brutal-bg);
  }
  @media (max-width: 800px) {
    .canvas-host,
    button {
      display: none;
    }
  }
</style>
