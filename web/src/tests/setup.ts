/**
 * Vitest globals setup.
 * Extends globalThis with test utilities if needed.
 */

// Globals (describe, it, expect) are enabled via vitest.config.ts globals: true

// jsdom does not have URL.createObjectURL / URL.revokeObjectURL
// Add minimal polyfill for blob URL support in tests
if (typeof URL.createObjectURL !== 'function') {
  let _blobUrlCounter = 0;
  URL.createObjectURL = function (_blob: unknown): string {
    return `blob:test-${++_blobUrlCounter}`;
  };
}

if (typeof URL.revokeObjectURL !== 'function') {
  URL.revokeObjectURL = function (_url: string): void {
    // no-op
  };
}

if (typeof URL.revokeObjectURL !== 'function') {
  URL.revokeObjectURL = function (_url: string): void {
    // no-op
  };
}
