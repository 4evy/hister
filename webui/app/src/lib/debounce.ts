// SPDX-License-Identifier: AGPL-3.0-or-later

/** Runs a callback after activity stops, with explicit cancellation and flushing. */
export function createDebouncer(callback: () => void, delayMs: number) {
  let timer: ReturnType<typeof setTimeout> | undefined;

  function cancel() {
    clearTimeout(timer);
    timer = undefined;
  }

  function flush() {
    if (timer === undefined) return;
    cancel();
    callback();
  }

  function schedule() {
    cancel();
    timer = setTimeout(flush, delayMs);
  }

  return { schedule, cancel, flush };
}
