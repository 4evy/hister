import { BRAILLE, PAGE, PAGE_TRANSITIONS, VIEW } from './constants.ts';
import { clamp } from './scalar.ts';
import type { Frame, Sheet, TexturePoint, Vertex } from './types.ts';

// Pérez-Díaz & Shen, Determination and (re)parametrization of rational
// developable surfaces, §2, eq (2.1), p 3, and Lemma 2.1's constant ruling
// https://arxiv.org/pdf/1305.2463#page=3
// We use the cylindrical form C(u) + vD with a trigonometric circular directrix,
// not the paper's rational parametrization algorithm
// radius * angular span = (halfWidth / bend) * (2 * bend) = 2 * halfWidth:
// changing bend preserves the sheet's arc length instead of stretching its text
export function project({ u, v }: TexturePoint, sheet: Sheet, frame: Frame): Vertex {
  const { time, cols, rows, scroll } = frame;
  const turn = clamp(scroll * PAGE_TRANSITIONS - sheet, 0, 1);
  const bend =
    PAGE.baseBend +
    Math.sin(turn * Math.PI) * PAGE.turnBend +
    PAGE.idleBend * Math.sin(time * PAGE.bendFrequency + sheet);
  const angle = (u - 0.5) * bend * 2;
  const radius = PAGE.halfWidth / bend;
  const x = radius * Math.sin(angle);
  const z = radius * (1 - Math.cos(angle)) + -sheet * PAGE.depthSpacing;
  const y = (v - 0.5) * PAGE.height;
  // The hinge affects screen x only: this is the original stylized placement,
  // not a rigid 3D rotation about that hinge
  const yaw = PAGE.yaw - turn * Math.PI + PAGE.idleYaw * Math.sin(time * PAGE.yawFrequency);
  const rx =
    (x + PAGE.hingeOffset) * Math.cos(yaw) +
    z * Math.sin(yaw) -
    PAGE.hingeOffset * Math.cos(PAGE.yaw);
  const rz = z * Math.cos(yaw) - x * Math.sin(yaw);
  const ry = y * VIEW.tiltCosine - rz * VIEW.tiltSine;
  // Perspective divide: q = 1/w, so closer vertices have larger q
  // These tilt coefficients have unit squared sum and form a rotation
  const q = 1 / (VIEW.distance - rz * VIEW.tiltCosine - y * VIEW.tiltSine);
  const scale = frameScale(cols, rows);
  return {
    x:
      (cols * VIEW.centerX +
        ((rx * Math.cos(VIEW.roll) + ry * Math.sin(VIEW.roll)) * VIEW.distance * q +
          sheet * PAGE.stackX) *
          scale) *
      BRAILLE.width,
    y:
      (rows * VIEW.centerY +
        ((ry * Math.cos(VIEW.roll) - rx * Math.sin(VIEW.roll)) * VIEW.distance * q -
          sheet * PAGE.stackY) *
          scale *
          VIEW.cellAspect) *
      BRAILLE.height,
    q,
    u,
    v,
  };
}

export function frameScale(cols: number, rows: number): number {
  return Math.min(cols * VIEW.widthScale, rows * VIEW.heightScale);
}
