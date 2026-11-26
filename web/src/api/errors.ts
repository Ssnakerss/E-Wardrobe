/**
 * API error types and normalization.
 */

export type ApiErrorKind =
  | 'network'
  | 'unauthorized'
  | 'forbidden'
  | 'not_found'
  | 'version_conflict'
  | 'gone'
  | 'payload_too_large'
  | 'validation'
  | 'rate_limited'
  | 'server'
  | 'invalid_response';

export interface ApiErrorDetails {
  [field: string]: string[];
}

export interface ApiError extends Error {
  kind: ApiErrorKind;
  httpStatus: number;
  code: string;
  message: string;
  details: ApiErrorDetails | null;
  retryAfter: number | null;
  currentVersion: number | null;
}

/**
 * Map HTTP status and API error code to a normalized ApiError.
 */
export function normalizeApiError(
  response: Response | null,
  body: unknown | null,
  networkError: boolean,
): ApiError {
  if (networkError && !response) {
    return {
      kind: 'network',
      httpStatus: 0,
      code: 'network_error',
      message: 'Network error. Check your connection.',
      details: null,
      retryAfter: null,
      currentVersion: null,
      name: 'ApiError',
    };
  }

  if (!response) {
    return {
      kind: 'server',
      httpStatus: 0,
      code: 'unknown',
      message: 'No response from server.',
      details: null,
      retryAfter: null,
      currentVersion: null,
      name: 'ApiError',
    };
  }

  // Parse error body
  let errorCode = '';
  let errorMessage = '';
  let details: ApiErrorDetails | null = null;
  let currentVersion: number | null = null;

  if (body && typeof body === 'object' && 'error' in body) {
    const errorObj = (body as Record<string, unknown>).error;
    if (errorObj && typeof errorObj === 'object' && errorObj !== null) {
      errorCode = ((errorObj as Record<string, unknown>).code as string) ?? '';
      errorMessage =
        ((errorObj as Record<string, unknown>).message as string) ?? response.statusText;
      const errorDetails = (errorObj as Record<string, unknown>).details as
        Record<string, unknown> | undefined;
      if (errorDetails) {
        details = Object.fromEntries(
          Object.entries(errorDetails).map(([k, v]) => [
            k,
            Array.isArray(v) ? (v as string[]) : [String(v)],
          ]),
        ) as ApiErrorDetails;
      }
    } else {
      errorMessage = errorObj as string;
    }
  } else {
    errorMessage = response.statusText || 'Unknown error';
  }

  // Determine kind from status
  let kind: ApiErrorKind;
  switch (response.status) {
    case 0:
    case 8: // Firefox network error status
      kind = 'network';
      break;
    case 401:
      kind = 'unauthorized';
      break;
    case 403:
      kind = 'forbidden';
      break;
    case 404:
      kind = 'not_found';
      break;
    case 409:
      kind = 'version_conflict';
      if (body && typeof body === 'object') {
        const b = body as Record<string, unknown>;
        if (typeof b.current_version === 'number') {
          currentVersion = b.current_version;
        }
      }
      break;
    case 410:
      kind = 'gone';
      break;
    case 413:
      kind = 'payload_too_large';
      break;
    case 422:
      kind = 'validation';
      break;
    case 429:
      kind = 'rate_limited';
      break;
    default:
      if (response.status >= 500) {
        kind = 'server';
      } else {
        kind = 'validation';
      }
  }

  const retryAfter = response.headers.get('Retry-After');
  const retryAfterMs = retryAfter ? parseInt(retryAfter, 10) * 1000 : null;

  return {
    kind,
    httpStatus: response.status,
    code: errorCode,
    message: errorMessage,
    details,
    retryAfter: retryAfterMs,
    currentVersion,
    name: 'ApiError',
  };
}

/**
 * Validate minimal response structure for critical endpoints.
 * Returns null if valid, or an ApiError if invalid.
 */
export function validateResponse(
  response: Response,
  expectedType: 'object' | 'empty',
): ApiError | null {
  const contentType = response.headers.get('Content-Type') ?? '';

  if (expectedType === 'empty') {
    return null;
  }

  if (!contentType.includes('application/json')) {
    return {
      kind: 'invalid_response',
      httpStatus: response.status,
      code: 'invalid_response',
      message: 'Unexpected response content type.',
      details: null,
      retryAfter: null,
      currentVersion: null,
      name: 'ApiError',
    };
  }

  return null;
}
