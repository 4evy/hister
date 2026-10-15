import { fieldFrame } from './ascii-field.ts';

export const initialTime = 2;
export const drawingFontSize = 12;

export function drawingSize(width: number, height: number) {
  const scale = Math.min(1, 1440 / Math.max(1, width), 900 / Math.max(1, height));
  return [Math.max(1, Math.round(width * scale)), Math.max(1, Math.round(height * scale))];
}

type GlyphAtlas = {
  width: number;
  height: number;
  paths: Record<string, string>;
};

export function staticField(width: number, height: number, atlas: GlyphAtlas) {
  const cols = Math.floor(width / atlas.width);
  const rows = Math.floor(height / atlas.height);
  const offsetX = Math.floor((width - cols * atlas.width) / 2);
  const offsetY = Math.floor((height - rows * atlas.height) / 2);
  const palette = ['#8080c0', '#c08060', '#60a080', '#c0a060'];
  const names = ['indigo', 'coral', 'teal', 'amber'];
  const cells = fieldFrame(cols, rows, initialTime, palette);
  const definitions = Object.entries(atlas.paths)
    .map(([code, path]) => `<path id="ascii-glyph-${code}" d="${path}"/>`)
    .join('');
  const body = cells
    .map((cell) => {
      const x = offsetX + cell.x * atlas.width;
      const y = offsetY + cell.y * atlas.height;
      const name = names[palette.indexOf(cell.color.slice(0, 7))];
      const opacity = cell.highlighted ? 1 : parseInt(cell.color.slice(7), 16) / 255;
      const background = cell.highlighted
        ? `<rect x="${x}" y="${y}" width="${atlas.width}" height="${atlas.height}" fill="var(--hister-coral)"/>`
        : '';
      const fill = cell.highlighted ? 'var(--brutal-bg)' : `var(--hister-${name})`;
      return `${background}<use href="#ascii-glyph-${cell.glyph.charCodeAt(0)}" x="${x}" y="${y}" fill="${fill}" opacity="${opacity}"/>`;
    })
    .join('');
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}" preserveAspectRatio="none" aria-hidden="true"><defs>${definitions}</defs><rect width="100%" height="100%" fill="var(--brutal-bg)"/>${body}</svg>`;
}
