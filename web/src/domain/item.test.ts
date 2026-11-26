/**
 * Unit tests for item domain functions.
 */

import { describe, it, expect } from 'vitest';
import {
  buildItemPatch,
  parsePrice,
  formatPrice,
  validateItemForm,
  serializeQueryParams,
} from './item.js';
import type { Item, ItemFields } from '../api/items.js';

function createItem(overrides: Partial<Item> = {}): Item {
  return {
    id: 'test-id',
    category_id: 'cat-1',
    name: 'Test Item',
    brand: 'Test Brand',
    description: 'Test description',
    purchase_year: 2024,
    purchase_price: '12990',
    price_currency: 'RUB',
    purchase_place: 'Store',
    condition_rating: 4,
    preference_rating: 3,
    last_worn_on: '2024-01-01',
    storage_location: 'Closet',
    season: 'winter',
    color: 'black',
    size: 'M',
    status: 'in_use',
    version: 1,
    archived_at: null,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    photos: [],
    ...overrides,
  };
}

describe('buildItemPatch', () => {
  it('builds patch with changed fields only', () => {
    const original = createItem();
    const draft: Partial<ItemFields> = {
      name: 'Updated Name',
      brand: null,
    };

    const patch = buildItemPatch(original, draft);

    expect(patch.base_version).toBe(1);
    expect(patch.fields).toEqual({
      name: 'Updated Name',
      brand: null,
    });
  });

  it('returns empty fields when nothing changed', () => {
    const original = createItem();
    const patch = buildItemPatch(original, {});

    expect(patch.base_version).toBe(1);
    expect(Object.keys(patch.fields)).toHaveLength(0);
  });

  it('includes null for cleared fields', () => {
    const original = createItem({ brand: 'Test Brand' });
    const patch = buildItemPatch(original, { brand: null });

    expect(patch.fields.brand).toBeNull();
  });
});

describe('parsePrice', () => {
  it('validates integer prices', () => {
    expect(parsePrice('12990')).toBe('12990');
    expect(parsePrice('0')).toBe('0');
    expect(parsePrice('100')).toBe('100');
  });

  it('rejects non-integer values', () => {
    expect(parsePrice('12.99')).toBeNull();
    expect(parsePrice('-100')).toBeNull();
    expect(parsePrice('abc')).toBeNull();
    expect(parsePrice('')).toBeNull();
    expect(parsePrice('  ')).toBeNull();
  });

  it('trims whitespace', () => {
    expect(parsePrice('  12990  ')).toBe('12990');
  });
});

describe('formatPrice', () => {
  it('formats price with currency', () => {
    const result = formatPrice('12990', 'RUB');
    expect(result).toMatch(/12[\s ]990/);
    expect(result).toContain('₽');
  });

  it('returns empty string for null price', () => {
    expect(formatPrice(null, 'RUB')).toBe('');
    expect(formatPrice('100', null)).toBe('');
  });
});

describe('validateItemForm', () => {
  it('validates required name', () => {
    const errors = validateItemForm({});
    expect(errors.name).toBe('Name is required');
  });

  it('validates required category', () => {
    const errors = validateItemForm({ name: 'Test' });
    expect(errors.category_id).toBe('Category is required');
  });

  it('validates year range', () => {
    const errors = validateItemForm({ name: 'Test', category_id: 'cat-1', purchase_year: 1800 });
    expect(errors.purchase_year).toBe('Year must be between 1900 and 2100');
  });

  it('validates rating range', () => {
    const errors = validateItemForm({ name: 'Test', category_id: 'cat-1', condition_rating: 6 });
    expect(errors.condition_rating).toBe('Rating must be between 1 and 5');
  });

  it('passes valid form', () => {
    const errors = validateItemForm({
      name: 'Valid Item',
      category_id: 'cat-1',
      status: 'in_use',
    });
    expect(Object.keys(errors)).toHaveLength(0);
  });
});

describe('serializeQueryParams', () => {
  it('serializes non-empty params', () => {
    const result = serializeQueryParams({
      q: 'shirt',
      status: 'in_use',
      limit: 20,
    });

    expect(result).toEqual({
      q: 'shirt',
      status: 'in_use',
      limit: '20',
    });
  });

  it('excludes empty/null/undefined params', () => {
    const result = serializeQueryParams({
      q: '',
      cursor: undefined as unknown as string,
      limit: null as unknown as number,
      sort: 'name_asc',
    });

    expect(result).toEqual({ sort: 'name_asc' });
  });
});
