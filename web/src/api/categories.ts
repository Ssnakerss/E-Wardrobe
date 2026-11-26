/**
 * Category API endpoints.
 */

import { apiGet, apiPost, apiPatch, apiFetch } from './client.js';
import { config } from '../app/config.js';
import type { CategoryFlat } from '../domain/category-tree.js';

export interface CategoryResponse {
  categories: CategoryFlat[];
}

export interface CreateCategoryRequest {
  name: string;
  parent_id?: string | null;
  sort_order?: number;
}

export interface UpdateCategoryRequest {
  base_version: number;
  fields?: {
    name?: string;
    parent_id?: string | null;
    sort_order?: number;
  };
}

/**
 * Fetch all categories.
 */
export async function fetchCategories(): Promise<CategoryFlat[]> {
  const response = await apiGet<CategoryResponse>('/api/v1/categories');
  return response.categories;
}

/**
 * Create a new category.
 */
export async function createCategory(data: CreateCategoryRequest): Promise<CategoryFlat> {
  return apiPost<CategoryFlat>('/api/v1/categories', data);
}

/**
 * Update a category with version-based optimistic locking.
 */
export async function updateCategory(
  id: string,
  data: UpdateCategoryRequest,
): Promise<CategoryFlat> {
  return apiPatch<CategoryFlat>(`/api/v1/categories/${id}`, data);
}

/**
 * Delete a category with version-based optimistic locking.
 */
export async function deleteCategory(id: string, baseVersion: number): Promise<void> {
  const url = new URL(`/api/v1/categories/${id}`, config.apiBase);
  url.searchParams.append('base_version', String(baseVersion));
  await apiFetch('DELETE', url.toString());
}

/**
 * Archive a category (soft delete via update).
 */
export async function archiveCategory(id: string, baseVersion: number): Promise<CategoryFlat> {
  return apiPatch<CategoryFlat>(`/api/v1/categories/${id}`, {
    base_version: baseVersion,
    fields: { archived_at: new Date().toISOString() },
  });
}
