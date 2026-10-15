import {
  INK,
  MATERIAL,
  MIN_TRIANGLE_AREA,
  SAMPLE_CENTER,
  SHADING,
  SHEET_MATERIALS,
  TEXTURE,
} from './constants.ts';
import { pages } from './texture.ts';
import type { Barycentrics, Scanline, Sheet, Surface, Triangle, Vertex } from './types.ts';

function sample(
  surface: Surface,
  index: number,
  a: Vertex,
  b: Vertex,
  c: Vertex,
  weights: Barycentrics,
  sheet: Sheet,
) {
  const [wa, wb, wc] = weights;
  // Brown, Barycentric Coordinates as Interpolants, eq (4), p 4
  // https://arxiv.org/pdf/1308.1279#page=4
  // Interpolate u/w and 1/w, then divide: linear screen-space UVs distort text
  const q = wa * a.q + wb * b.q + wc * c.q;
  if (q <= surface.depth[index]) return;
  const u = (wa * a.u * a.q + wb * b.u * b.q + wc * c.u * c.q) / q;
  const v = (wa * a.v * a.q + wb * b.v * b.q + wc * c.v * c.q) / q;
  surface.depth[index] = q;
  // Art-directed cosine shading uses a fixed bend, not the animated surface normal
  const normalAngle = (u - 0.5) * SHADING.normalSweep;
  surface.shade[index] = Math.round(
    SHADING.ambient + SHADING.diffuse * Math.max(0, Math.cos(normalAngle - SHADING.lightAngle)),
  );
  const tx = Math.min(TEXTURE.size - 1, Math.max(0, Math.floor(u * TEXTURE.size)));
  const ty = Math.min(TEXTURE.size - 1, Math.max(0, Math.floor(v * TEXTURE.size)));
  const ink = pages[sheet][ty * TEXTURE.size + tx];
  surface.ink[index] = ink;
  surface.material[index] = ink === INK.ribbon ? MATERIAL.ribbon : SHEET_MATERIALS[sheet];
}

// Brown, Barycentric Coordinates as Interpolants, eqs (1)-(2), pp 2-3
// https://arxiv.org/pdf/1308.1279#page=3
// Signed double-area ratios sum to one; negative weights are outside the triangle
// The common factor of 1/2 cancels, and either vertex winding works
export function triangle(surface: Surface, a: Vertex, b: Vertex, c: Vertex, sheet: Sheet) {
  const area = (b.y - c.y) * (a.x - c.x) + (c.x - b.x) * (a.y - c.y);
  if (Math.abs(area) < MIN_TRIANGLE_AREA) return;
  const left = Math.max(0, Math.floor(Math.min(a.x, b.x, c.x)));
  const right = Math.min(surface.width - 1, Math.ceil(Math.max(a.x, b.x, c.x)));
  const top = Math.max(0, Math.floor(Math.min(a.y, b.y, c.y)));
  const bottom = Math.min(surface.height - 1, Math.ceil(Math.max(a.y, b.y, c.y)));
  for (let y = top; y <= bottom; y++) {
    scanline(surface, [a, b, c], { left, right, y, area }, sheet);
  }
}

function scanline(surface: Surface, vertices: Triangle, bounds: Scanline, sheet: Sheet) {
  const [a, b, c] = vertices;
  const { left, right, y, area } = bounds;
  for (let x = left; x <= right; x++) {
    const wa =
      ((b.y - c.y) * (x + SAMPLE_CENTER - c.x) + (c.x - b.x) * (y + SAMPLE_CENTER - c.y)) / area;
    const wb =
      ((c.y - a.y) * (x + SAMPLE_CENTER - c.x) + (a.x - c.x) * (y + SAMPLE_CENTER - c.y)) / area;
    const wc = 1 - wa - wb;
    if (Math.min(wa, wb, wc) < 0) continue;
    sample(surface, y * surface.width + x, a, b, c, [wa, wb, wc], sheet);
  }
}
