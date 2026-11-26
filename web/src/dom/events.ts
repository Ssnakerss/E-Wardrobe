/**
 * Event helpers — common event utilities.
 */

/**
 * Check if an event should be ignored for link navigation.
 */
export function shouldIgnoreClick(e: MouseEvent): boolean {
  return !!(e.ctrlKey || e.metaKey || e.altKey || e.shiftKey || e.button !== 0);
}

/**
 * Debounce a function.
 */
export function debounce<T extends (...args: unknown[]) => void>(
  fn: T,
  ms: number,
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout> | null = null;
  return (...args: Parameters<T>) => {
    if (timer !== null) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      fn(...args);
      timer = null;
    }, ms);
  };
}
