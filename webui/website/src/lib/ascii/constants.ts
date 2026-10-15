export const PAGE = {
  sheets: [0, 1, 2],
  segments: 28,
  baseBend: 0.75,
  turnBend: 0.6,
  idleBend: 0.13,
  bendFrequency: 0.23,
  halfWidth: 0.7,
  height: 1.65,
  depthSpacing: 0.16,
  yaw: -0.48,
  idleYaw: 0.045,
  yawFrequency: 0.17,
  hingeOffset: 0.65,
  stackX: 0.2,
  stackY: 0.12,
} as const;

export const VIEW = {
  distance: 3.8,
  tiltCosine: 0.96,
  tiltSine: 0.28,
  roll: 0.16,
  widthScale: 0.32,
  heightScale: 0.73,
  centerX: 0.77,
  centerY: 0.53,
  cellAspect: 0.6,
} as const;

export const BRAILLE = {
  width: 2,
  height: 4,
  bits: [1, 8, 2, 16, 4, 32, 64, 128],
  baseCodePoint: 0x2800,
  masks: 256,
} as const;

export const LENS = {
  legSeconds: 3.6,
  restFraction: 0.16,
  radius: 0.27,
  magnification: 1.7,
  rimHalfWidth: 0.035,
  handleEnd: 2.05,
  handleHalfWidth: 0.095,
  boundsPadding: 4,
  boundsExtent: 2.2,
} as const;

export const TEXTURE = {
  size: 192,
  rows: 65,
  columns: 85,
  columnFrequency: 2.4,
  lineLength: 0.58,
  lineVariation: 0.3,
  rowPhase: 13,
  sheetPhase: 7,
  lineSpacing: 3,
  wordPeriod: 11,
  wordGap: 2,
  matchHalfWidth: 0.03,
  ribbonU: 0.8,
  ribbonSpacing: 0.09,
  ribbonHalfWidth: 0.028,
  ribbonBottom: 0.18,
  ribbonNotch: 0.152,
  border: 0.008,
  marginU: 0.075,
  marginV: 0.065,
  headingBottom: 0.15,
  titleBottom: 0.105,
  titleWidth: 0.67,
  titleLineSpacing: 2,
  ruleTop: 0.137,
} as const;

export const INK = {
  blank: 0,
  text: 1,
  match: 2,
  ribbon: 3,
} as const;
export const MATERIAL = {
  page: 0,
  highlight: 1,
  ribbon: 2,
  lens: 3,
} as const;
export const SHEET_MATERIALS = [MATERIAL.page, MATERIAL.ribbon, MATERIAL.lens] as const;
export const SHADING = {
  normalSweep: 1.5,
  lightAngle: 0.8,
  ambient: 1,
  diffuse: 2,
  lens: 5,
} as const;
export const ALPHAS = [24, 72, 108, 150, 204, 255] as const;

// Signed double area is measured in squared braille-dot units
export const MIN_TRIANGLE_AREA = 0.0001;
export const SAMPLE_CENTER = 0.5;
export const PAGE_TRANSITIONS = PAGE.sheets.length - 1;
