/**
 * Items API endpoints.
 */

import { apiGet, apiPost, apiPatch, apiDelete, apiPut } from './client.js';

export type ItemStatus = 'in_use' | 'stored' | 'needs_repair';
export type Season = 'winter' | 'spring' | 'summer' | 'autumn' | 'demi_season' | 'all_season';

export interface ItemFields {
  category_id: string;
  name: string;
  brand: string | null;
  description: string | null;
  purchase_year: number | null;
  purchase_price: string | null;
  price_currency: string | null;
  purchase_place: string | null;
  condition_rating: number | null;
  preference_rating: number | null;
  last_worn_on: string | null;
  storage_location: string | null;
  season: Season | null;
  color: string | null;
  size: string | null;
  status: ItemStatus;
}

export interface Item extends ItemFields {
  id: string;
  version: number;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
  photos: Photo[];
}

export interface ItemInput extends Omit<ItemFields, 'status'> {
  status: ItemStatus;
}

export interface ItemPatch {
  base_version: number;
  fields: Partial<ItemFields>;
}

export interface ItemPage {
  items: Item[];
  next_cursor: string | null;
}

export interface ArchiveRequest {
  base_version: number;
  archived: boolean;
}

export interface Photo {
  id: string;
  item_id: string;
  sort_order: number;
  status: string;
  mime_type: string | null;
  size_bytes: number | null;
  width: number | null;
  height: number | null;
  version: number;
}

export interface StorageLocationSuggestion {
  suggestions: string[];
}

export interface ItemsQueryParams {
  q?: string;
  category_id?: string;
  include_descendants?: boolean;
  status?: ItemStatus;
  season?: Season;
  condition_rating?: number;
  preference_rating?: number;
  archived?: boolean;
  sort?: string;
  cursor?: string;
  limit?: number;
}

export async function fetchItems(params: ItemsQueryParams = {}): Promise<ItemPage> {
  return apiGet<ItemPage>(
    '/api/v1/items',
    params as Record<string, string | number | boolean | null | undefined>,
  );
}

export async function fetchItem(id: string): Promise<Item> {
  return apiGet<Item>(`/api/v1/items/${id}`);
}

export async function createItem(data: ItemInput): Promise<Item> {
  return apiPost<Item>('/api/v1/items', data);
}

export async function updateItem(id: string, patch: ItemPatch): Promise<Item> {
  return apiPatch<Item>(`/api/v1/items/${id}`, patch);
}

export async function archiveItem(id: string, request: ArchiveRequest): Promise<Item> {
  return apiPost<Item>(`/api/v1/items/${id}/archive`, request);
}

export async function deleteItem(id: string, baseVersion: number): Promise<void> {
  await apiDelete(`/api/v1/items/${id}?base_version=${baseVersion}`);
}

export async function fetchStorageLocations(
  q: string,
  limit: number = 10,
): Promise<StorageLocationSuggestion> {
  return apiGet<StorageLocationSuggestion>('/api/v1/storage-locations', { q, limit });
}

/* ==========================================================================
   Photo API endpoints
   ========================================================================== */

export interface PhotoOrderRequest {
  base_version: number;
  photos: { id: string; version: number }[];
}

export async function uploadPhoto(itemId: string, file: File, sortOrder: number): Promise<Photo> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('sort_order', String(sortOrder));

  return apiPost<Photo>(`/api/v1/items/${itemId}/photos`, formData);
}

export async function deletePhoto(photoId: string, baseVersion: number): Promise<void> {
  await apiDelete(`/api/v1/photos/${photoId}?base_version=${baseVersion}`);
}

export async function reorderPhotos(
  itemId: string,
  baseVersion: number,
  orderedPhotos: { id: string; version: number }[],
): Promise<void> {
  await apiPut(`/api/v1/items/${itemId}/photos/order`, {
    base_version: baseVersion,
    photos: orderedPhotos,
  });
}
