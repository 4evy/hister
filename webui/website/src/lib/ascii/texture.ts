import { INK, PAGE, TEXTURE } from './constants.ts';

// Sparse word positions are shared by the page texture and the lens itinerary
export const matchWords = [
  { u: 0.24, row: 24 },
  { u: 0.66, row: 39 },
  { u: 0.3, row: 51 },
] as const;

// A deterministic sinusoidal pattern varies line lengths without per-frame randomness
function bodyInk(u: number, v: number, sheet: number) {
  const row = Math.floor(v * TEXTURE.rows);
  const column = Math.floor(u * TEXTURE.columns);
  const columnU = ((u - TEXTURE.marginU) * TEXTURE.columnFrequency) % 1;
  const lineEnd =
    TEXTURE.lineLength +
    TEXTURE.lineVariation * Math.sin(row * TEXTURE.rowPhase + sheet * TEXTURE.sheetPhase) ** 2;
  if (
    [
      columnU > lineEnd,
      row % TEXTURE.lineSpacing !== 0,
      (column + row * TEXTURE.lineSpacing) % TEXTURE.wordPeriod < TEXTURE.wordGap,
    ].some(Boolean)
  )
    return INK.blank;
  return matchWords.some(
    (word) => word.row === row && Math.abs(u - word.u) < TEXTURE.matchHalfWidth,
  )
    ? INK.match
    : INK.text;
}

function pageInk(u: number, v: number, sheet: number) {
  const ribbon = Math.abs(u - (TEXTURE.ribbonU - sheet * TEXTURE.ribbonSpacing));
  if (
    ribbon < TEXTURE.ribbonHalfWidth &&
    v < TEXTURE.ribbonBottom &&
    v < TEXTURE.ribbonNotch + ribbon
  )
    return INK.ribbon;
  return pageContentInk(u, v, sheet);
}

function pageContentInk(u: number, v: number, sheet: number) {
  if (Math.min(u, v, 1 - u, 1 - v) < TEXTURE.border) return INK.text;
  if (
    [
      u < TEXTURE.marginU,
      u > 1 - TEXTURE.marginU,
      v < TEXTURE.marginV,
      v > 1 - TEXTURE.marginV,
    ].some(Boolean)
  )
    return INK.blank;
  if (v < TEXTURE.headingBottom) return headingInk(u, v);
  return bodyInk(u, v, sheet);
}

function headingInk(u: number, v: number) {
  if (v < TEXTURE.titleBottom)
    return Number(
      u < TEXTURE.titleWidth && Math.floor(v * TEXTURE.rows) % TEXTURE.titleLineSpacing === 0,
    );
  return Number(v > TEXTURE.ruleTop);
}

const textureSize = TEXTURE.size;
export const pages = Array.from({ length: PAGE.sheets.length }, (_, sheet) =>
  Uint8Array.from({ length: textureSize * textureSize }, (_, i) =>
    pageInk((i % textureSize) / textureSize, Math.floor(i / textureSize) / textureSize, sheet),
  ),
);
