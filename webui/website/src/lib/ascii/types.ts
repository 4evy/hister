import type { PAGE } from './constants.ts';

export type Cell = {
  x: number;
  y: number;
  glyph: string;
  color: string;
  highlighted?: boolean;
};
export type Vertex = ScreenPoint & TexturePoint & { readonly q: number };
export type Surface = {
  width: number;
  height: number;
  depth: Float32Array;
  ink: Uint8Array;
  material: Uint8Array;
  shade: Uint8Array;
};

// Screen coordinates are measured in braille dots, texture coordinates in page fractions
export type ScreenPoint = Readonly<{ x: number; y: number }>;
export type TexturePoint = Readonly<{ u: number; v: number }>;
// Time is elapsed animation seconds; scroll is the clamped fraction of the page-turn sequence
export type Frame = Readonly<{ cols: number; rows: number; time: number; scroll: number }>;
export type Triangle = readonly [a: Vertex, b: Vertex, c: Vertex];
export type Barycentrics = readonly [a: number, b: number, c: number];
export type Scanline = Readonly<{ left: number; right: number; y: number; area: number }>;
export type Lens = Readonly<{ cx: number; cy: number; rx: number; ry: number }>;
export type Sheet = (typeof PAGE.sheets)[number];
