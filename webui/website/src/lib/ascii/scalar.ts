// TS 7's standard Math object supplies scalar operations, not clamp, lerp, or vectors
export function clamp(value: number, minimum: number, maximum: number): number {
  return Math.max(minimum, Math.min(maximum, value));
}

export function lerp(from: number, to: number, fraction: number): number {
  return from + (to - from) * fraction;
}

// The cubic satisfying f(0)=0, f(1)=1, f'(0)=f'(1)=0 is 3t²-2t³
// Its zero endpoint velocities make the lens leave and arrive at words smoothly
// This is a Hermite polynomial derived from those constraints, not an arXiv formula
export function smoothstep(fraction: number): number {
  return fraction * fraction * (3 - 2 * fraction);
}
