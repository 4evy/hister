import {
  BRAILLE,
  INK,
  LENS,
  MATERIAL,
  PAGE,
  PAGE_TRANSITIONS,
  SAMPLE_CENTER,
  SHADING,
  TEXTURE,
  VIEW,
} from './constants.ts';
import { frameScale, project } from './projection.ts';
import { clamp, lerp, smoothstep } from './scalar.ts';
import { matchWords } from './texture.ts';
import type { Frame, Lens, Surface } from './types.ts';

function wordPosition(word: (typeof matchWords)[number], frame: Frame) {
  const { scroll } = frame;
  const sheet = PAGE.sheets[Math.floor(scroll * PAGE_TRANSITIONS)];
  const fraction = scroll * PAGE_TRANSITIONS - sheet;
  const blend = smoothstep(fraction);
  const texture = { u: word.u, v: (word.row + SAMPLE_CENTER) / TEXTURE.rows };
  const a = project(texture, sheet, frame);
  const b = project(texture, PAGE.sheets[Math.min(PAGE_TRANSITIONS, sheet + 1)], frame);
  return { x: lerp(a.x, b.x, blend), y: lerp(a.y, b.y, blend) };
}

function lensPosition(cols: number, rows: number, time: number, scroll: number) {
  const leg = time / LENS.legSeconds;
  const index = Math.floor(leg) % matchWords.length;
  // Rest briefly on each match before accelerating across the page
  const fraction = Math.max(0, ((leg % 1) - LENS.restFraction) / (1 - LENS.restFraction));
  // Ease in and out at words, with the fastest travel between them
  const travel = smoothstep(fraction);
  const from = matchWords[index];
  const to = matchWords[(index + 1) % matchWords.length];
  const a = wordPosition(from, { cols, rows, time, scroll });
  const b = wordPosition(to, { cols, rows, time, scroll });
  return { x: lerp(a.x, b.x, travel), y: lerp(a.y, b.y, travel) };
}

// Focus+context keeps the document visible around an enlarged region
// Li, Zhao & Qin, A New 3D Geometric Approach to Focus+Context Lens Effect Simulation,
// §3-4, pp 2-3: https://arxiv.org/pdf/1308.0375#page=3
// Design inspiration only: their mesh lifting and optimization are not implemented
// Our inverse affine map is source = center + (destination - center) / magnification
export function magnify(
  surface: Surface,
  cols: number,
  rows: number,
  time: number,
  scroll: number,
) {
  const scale = frameScale(cols, rows);
  const position = lensPosition(cols, rows, time, scroll);
  const rx = scale * LENS.radius * BRAILLE.width;
  const ry = scale * LENS.radius * VIEW.cellAspect * BRAILLE.height;
  const cx = clamp(position.x, rx, surface.width - rx);
  const cy = clamp(position.y, ry, surface.height - ry);
  const source = {
    ink: surface.ink.slice(),
    material: surface.material.slice(),
    shade: surface.shade.slice(),
  };
  for (
    let y = Math.max(0, Math.floor(cy - ry - LENS.boundsPadding));
    y < Math.min(surface.height, cy + ry * LENS.boundsExtent);
    y++
  ) {
    for (
      let x = Math.max(0, Math.floor(cx - rx - LENS.boundsPadding));
      x < Math.min(surface.width, cx + rx * LENS.boundsExtent);
      x++
    ) {
      lensPixel(surface, source, x, y, { cx, cy, rx, ry });
    }
  }
}

function lensPixel(
  surface: Surface,
  source: Pick<Surface, 'ink' | 'material' | 'shade'>,
  x: number,
  y: number,
  lens: Lens,
) {
  const { cx, cy, rx, ry } = lens;
  // Normalize the ellipse to a unit circle to account for non-square glyph cells
  const dx = (x - cx) / rx;
  const dy = (y - cy) / ry;
  const distance = Math.hypot(dx, dy);
  const index = y * surface.width + x;
  // Rotate by 45 degrees into the handle's local axes
  // Division by sqrt(2) keeps the axes unit length
  const along = (dx + dy) / Math.SQRT2;
  const across = Math.abs(dx - dy) / Math.SQRT2;
  // Intersection of three half-planes: negative exactly inside the handle rectangle
  const handleDistance = Math.max(1 - along, along - LENS.handleEnd, across - LENS.handleHalfWidth);
  if (Math.min(Math.abs(distance - 1) - LENS.rimHalfWidth, handleDistance) < 0) {
    surface.ink[index] = INK.text;
    surface.material[index] = MATERIAL.lens;
    surface.shade[index] = SHADING.lens;
  } else if (distance < 1 - LENS.rimHalfWidth) {
    magnifiedPixel(surface, source, index, x, y, cx, cy);
  }
}

function magnifiedPixel(
  surface: Surface,
  source: Pick<Surface, 'ink' | 'material' | 'shade'>,
  index: number,
  x: number,
  y: number,
  cx: number,
  cy: number,
) {
  // Sampling between the visible pixel and lens center stays inside the surface
  const sx = Math.floor(cx + (x - cx) / LENS.magnification);
  const sy = Math.floor(cy + (y - cy) / LENS.magnification);
  const origin = sy * surface.width + sx;
  surface.ink[index] = source.ink[origin];
  surface.material[index] =
    source.ink[origin] === INK.match ? MATERIAL.highlight : source.material[origin];
  surface.shade[index] = source.shade[origin] ? SHADING.lens : 0;
}
