import type { Cell, Surface } from './ascii/types.ts';
import { ALPHAS, BRAILLE, MATERIAL, PAGE } from './ascii/constants.ts';
import { clamp } from './ascii/scalar.ts';
import { project } from './ascii/projection.ts';
import { triangle } from './ascii/raster.ts';
import { magnify } from './ascii/lens.ts';

const dotGlyphs = Array.from({ length: BRAILLE.masks }, (_, mask) =>
  String.fromCharCode(BRAILLE.baseCodePoint + mask),
);
dotGlyphs[0] = '·';

function cellDots(surface: Surface, x: number, y: number) {
  let mask = 0;
  let material = 0;
  let shade = 0;
  for (let dot = 0; dot < BRAILLE.bits.length; dot++) {
    const index =
      (y * BRAILLE.height + Math.floor(dot / BRAILLE.width)) * surface.width +
      x * BRAILLE.width +
      (dot % BRAILLE.width);
    shade = Math.max(shade, surface.shade[index]);
    if (!surface.ink[index]) continue;
    mask |= BRAILLE.bits[dot];
    if (material !== MATERIAL.highlight) material = surface.material[index];
  }
  return { mask, material, shade };
}

function resolveCell(surface: Surface, x: number, y: number, colors: string[]): Cell | undefined {
  const { mask, material, shade } = cellDots(surface, x, y);
  if (!shade) return;
  return {
    x,
    y,
    glyph: dotGlyphs[mask],
    color: colors[material * ALPHAS.length + shade * Number(mask > 0)],
    highlighted: material === MATERIAL.highlight,
  };
}

function resolve(surface: Surface, cols: number, rows: number, colors: string[]) {
  const cells: Cell[] = [];
  for (let y = 0; y < rows; y++) {
    for (let x = 0; x < cols; x++) {
      const cell = resolveCell(surface, x, y, colors);
      if (cell) cells.push(cell);
    }
  }
  return cells;
}

export function fieldFrame(cols: number, rows: number, time: number, colors: string[], scroll = 0) {
  scroll = clamp(scroll, 0, 1);
  const width = cols * BRAILLE.width;
  const height = rows * BRAILLE.height;
  const surface: Surface = {
    width,
    height,
    depth: new Float32Array(width * height),
    ink: new Uint8Array(width * height),
    material: new Uint8Array(width * height),
    shade: new Uint8Array(width * height),
  };
  const frame = { cols, rows, time, scroll };
  for (const sheet of PAGE.sheets) {
    for (let segment = 0; segment < PAGE.segments; segment++) {
      const a = project({ u: segment / PAGE.segments, v: 0 }, sheet, frame);
      const b = project({ u: (segment + 1) / PAGE.segments, v: 0 }, sheet, frame);
      const c = project({ u: segment / PAGE.segments, v: 1 }, sheet, frame);
      const d = project({ u: (segment + 1) / PAGE.segments, v: 1 }, sheet, frame);
      triangle(surface, a, b, c, sheet);
      triangle(surface, b, d, c, sheet);
    }
  }
  magnify(surface, cols, rows, time, scroll);
  const tones = colors.flatMap((color) =>
    ALPHAS.map((alpha) => `${color}${alpha.toString(16).padStart(2, '0')}`),
  );
  return resolve(surface, cols, rows, tones);
}
