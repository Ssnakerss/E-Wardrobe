/**
 * Domain functions for items: patch building, price handling, validation.
 */

import type { Item, ItemPatch, ItemFields, ItemInput } from '../api/items.js';

/**
 * Build a minimal ItemPatch from original and draft.
 * Only includes fields that differ from the original.
 * Sends null for cleared nullable fields.
 */
export function buildItemPatch(original: Item, draft: Partial<ItemFields>): ItemPatch {
  const fields: Partial<ItemFields> = {};

  for (const key of Object.keys(draft) as (keyof ItemFields)[]) {
    const originalValue = (original as unknown as Record<string, unknown>)[key];
    const draftValue = draft[key];

    if (originalValue !== draftValue) {
      (fields as Record<string, unknown>)[key] = draftValue;
    }
  }

  return {
    base_version: original.version,
    fields,
  };
}

/**
 * Validate price string: must be integer major currency units.
 * Returns the validated string or null if invalid.
 */
export function parsePrice(value: string): string | null {
  if (!value || value.trim() === '') {
    return null;
  }

  const trimmed = value.trim();

  // Must be a non-negative integer
  if (!/^(0|[1-9][0-9]*)$/.test(trimmed)) {
    return null;
  }

  const num = Number(trimmed);
  if (num < 0 || !Number.isInteger(num)) {
    return null;
  }

  return trimmed;
}

/**
 * Format price with Intl.NumberFormat, using the currency.
 * Returns empty string if price or currency is null.
 */
export function formatPrice(price: string | null, currency: string | null): string {
  if (!price || !currency) {
    return '';
  }

  try {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(Number(price));
  } catch {
    return `${price} ${currency}`;
  }
}

/**
 * Validate item form data.
 * Returns a record of field errors: { [field]: message }.
 */
export function validateItemForm(data: Partial<ItemInput>): Record<string, string> {
  const errors: Record<string, string> = {};

  if (!data.name || data.name.trim() === '') {
    errors.name = 'Name is required';
  } else if (data.name.length > 200) {
    errors.name = 'Name must be 200 characters or less';
  }

  if (!data.category_id || data.category_id.trim() === '') {
    errors.category_id = 'Category is required';
  }

  if (data.purchase_year !== undefined && data.purchase_year !== null) {
    if (data.purchase_year < 1900 || data.purchase_year > 2100) {
      errors.purchase_year = 'Year must be between 1900 and 2100';
    }
  }

  if (data.purchase_price !== undefined && data.purchase_price !== null) {
    const parsed = parsePrice(data.purchase_price);
    if (parsed === null && data.purchase_price.trim() !== '') {
      errors.purchase_price = 'Price must be a positive integer';
    }
  }

  if (data.condition_rating !== undefined && data.condition_rating !== null) {
    if (data.condition_rating < 1 || data.condition_rating > 5) {
      errors.condition_rating = 'Rating must be between 1 and 5';
    }
  }

  if (data.preference_rating !== undefined && data.preference_rating !== null) {
    if (data.preference_rating < 1 || data.preference_rating > 5) {
      errors.preference_rating = 'Rating must be between 1 and 5';
    }
  }

  return errors;
}

/**
 * Serialize items query params for URL.
 * Only includes non-null, non-undefined values.
 */
export function serializeQueryParams(params: {
  q?: string;
  category_id?: string;
  include_descendants?: boolean;
  status?: string;
  season?: string;
  condition_rating?: number;
  preference_rating?: number;
  archived?: boolean;
  sort?: string;
  cursor?: string;
  limit?: number;
}): Record<string, string> {
  const result: Record<string, string> = {};

  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') {
      continue;
    }
    result[key] = String(value);
  }

  return result;
}

/**
 * Deserialize URL search params into items query params.
 */
export function deserializeQueryParams(search: string): Record<string, unknown> {
  const params = new URLSearchParams(search);
  const result: Record<string, unknown> = {};

  for (const [key, value] of params) {
    if (value === '' || value === undefined) continue;

    if (key === 'include_descendants' || key === 'archived') {
      result[key] = value === 'true';
    } else if (key === 'condition_rating' || key === 'preference_rating' || key === 'limit') {
      const num = parseInt(value, 10);
      result[key] = isNaN(num) ? value : num;
    } else {
      result[key] = value;
    }
  }

  return result;
}
