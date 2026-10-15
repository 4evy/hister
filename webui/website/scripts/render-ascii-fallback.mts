import { writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { openSync } from 'fontkit';
import { drawingFontSize, staticField } from '../src/lib/ascii-svg.ts';

const font = openSync(
  fileURLToPath(new URL('../src/lib/assets/drawing-font/garden-mono.woff', import.meta.url)),
);
if (!('glyphForCodePoint' in font)) throw new Error('Expected a single drawing font');
const scale = drawingFontSize / font.unitsPerEm;
const glyphs = font.characterSet.map((code) => font.glyphForCodePoint(code));
const width = Math.ceil(Math.max(...glyphs.map((glyph) => glyph.advanceWidth * scale)));
const height = Math.ceil(
  Math.max(...glyphs.map((glyph) => (glyph.bbox.maxY - glyph.bbox.minY) * scale)),
);
const baseline = Math.round(height / 2 - drawingFontSize / 2) + font.ascent * scale;
const codes = [183, ...Array.from({ length: 255 }, (_, i) => 0x2801 + i)];
const paths = Object.fromEntries(
  codes.map((code) => {
    if (!font.hasGlyphForCodePoint(code)) throw new Error(`Missing glyph ${code}`);
    const glyph = font.glyphForCodePoint(code);
    const x = (width - glyph.advanceWidth * scale) / 2;
    return [code, glyph.path.transform(scale, 0, 0, -scale, x, baseline).toSVG()];
  }),
);
const atlas = { width, height, paths };
await writeFile(
  new URL('../src/lib/assets/ascii-glyphs.json', import.meta.url),
  `${JSON.stringify(atlas, null, 2)}\n`,
);
await writeFile(
  new URL('../src/lib/assets/ascii-fallback.svg', import.meta.url),
  `${staticField(1440, 900, atlas)}\n`,
);
