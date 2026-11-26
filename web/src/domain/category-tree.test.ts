/**
 * Tests for category tree domain functions.
 */

import { describe, it, expect } from 'vitest';
import {
  buildCategoryTree,
  getCategoryById,
  getAllDescendantIds,
  getAvailableParents,
  toggleNodeExpanded,
  type CategoryFlat,
} from './category-tree.js';

// Test fixtures
const flatCategories: CategoryFlat[] = [
  {
    id: '1',
    name: 'Clothing',
    parent_id: null,
    sort_order: 1,
    version: 1,
    archived_at: null,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: '2',
    name: 'Shoes',
    parent_id: null,
    sort_order: 2,
    version: 1,
    archived_at: null,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: '3',
    name: 'T-shirts',
    parent_id: '1',
    sort_order: 1,
    version: 1,
    archived_at: null,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: '4',
    name: 'Jeans',
    parent_id: '1',
    sort_order: 2,
    version: 1,
    archived_at: null,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: '5',
    name: 'Sneakers',
    parent_id: '2',
    sort_order: 1,
    version: 1,
    archived_at: null,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
  {
    id: '6',
    name: 'Vintage T-shirts',
    parent_id: '3',
    sort_order: 1,
    version: 1,
    archived_at: null,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-01T00:00:00Z',
  },
];

describe('buildCategoryTree', () => {
  it('builds correct tree from flat list', () => {
    const tree = buildCategoryTree(flatCategories);

    expect(tree).toHaveLength(2); // Clothing, Shoes (roots)
    const clothing = tree[0]!;
    expect(clothing.id).toBe('1');
    expect(clothing.name).toBe('Clothing');
    expect(clothing.children).toHaveLength(2); // T-shirts, Jeans
    const tshirts = clothing.children[0]!;
    expect(tshirts.id).toBe('3'); // T-shirts (sort_order: 1)
    const jeans = clothing.children[1]!;
    expect(jeans.id).toBe('4'); // Jeans (sort_order: 2)
    expect(tshirts.children).toHaveLength(1); // Vintage T-shirts
    expect(tshirts.children[0]!.id).toBe('6');
    const shoes = tree[1]!;
    expect(shoes.id).toBe('2');
    expect(shoes.children).toHaveLength(1); // Sneakers
    expect(shoes.children[0]!.id).toBe('5');
  });

  it('handles empty list', () => {
    const tree = buildCategoryTree([]);
    expect(tree).toEqual([]);
  });

  it('handles single category', () => {
    const single: CategoryFlat[] = [
      {
        id: '1',
        name: 'Root',
        parent_id: null,
        sort_order: 1,
        version: 1,
        archived_at: null,
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z',
      },
    ];
    const tree = buildCategoryTree(single);
    expect(tree).toHaveLength(1);
    expect(tree[0]!.id).toBe('1');
    expect(tree[0]!.children).toHaveLength(0);
  });

  it('handles missing parents gracefully', () => {
    const withMissingParent: CategoryFlat[] = [
      {
        id: '1',
        name: 'Child',
        parent_id: '999', // parent doesn't exist
        sort_order: 1,
        version: 1,
        archived_at: null,
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z',
      },
    ];
    const tree = buildCategoryTree(withMissingParent);
    expect(tree).toHaveLength(1);
    expect(tree[0]!.id).toBe('1');
    expect(tree[0]!.children).toHaveLength(0);
  });

  it('sorts nodes by sort_order', () => {
    const unsorted: CategoryFlat[] = [
      {
        id: '1',
        name: 'B',
        parent_id: null,
        sort_order: 2,
        version: 1,
        archived_at: null,
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z',
      },
      {
        id: '2',
        name: 'A',
        parent_id: null,
        sort_order: 1,
        version: 1,
        archived_at: null,
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-01T00:00:00Z',
      },
    ];
    const tree = buildCategoryTree(unsorted);
    expect(tree[0]!.name).toBe('A');
    expect(tree[1]!.name).toBe('B');
  });

  it('marks nodes as not expanded by default', () => {
    const tree = buildCategoryTree(flatCategories);
    const clothing = tree[0]!;
    expect(clothing.expanded).toBe(false);
    expect(clothing.children[0]!.expanded).toBe(false);
  });
});

describe('getCategoryById', () => {
  it('finds node by id at root level', () => {
    const tree = buildCategoryTree(flatCategories);
    const found = getCategoryById(tree, '1');
    expect(found).not.toBeNull();
    expect(found!.name).toBe('Clothing');
  });

  it('finds node by id at nested level', () => {
    const tree = buildCategoryTree(flatCategories);
    const found = getCategoryById(tree, '6');
    expect(found).not.toBeNull();
    expect(found!.name).toBe('Vintage T-shirts');
  });

  it('returns null for non-existent id', () => {
    const tree = buildCategoryTree(flatCategories);
    const found = getCategoryById(tree, 'nonexistent');
    expect(found).toBeNull();
  });

  it('returns null for empty tree', () => {
    const found = getCategoryById([], 'any');
    expect(found).toBeNull();
  });
});

describe('getAllDescendantIds', () => {
  it('returns all descendants recursively', () => {
    const tree = buildCategoryTree(flatCategories);
    const ids = getAllDescendantIds(tree, '1'); // Clothing
    expect(ids).toEqual(['3', '6', '4']); // T-shirts, Vintage T-shirts, Jeans (depth-first)
  });

  it('returns empty array for leaf node', () => {
    const tree = buildCategoryTree(flatCategories);
    const ids = getAllDescendantIds(tree, '6'); // Vintage T-shirts (leaf)
    expect(ids).toEqual([]);
  });

  it('returns empty array for non-existent node', () => {
    const tree = buildCategoryTree(flatCategories);
    const ids = getAllDescendantIds(tree, 'nonexistent');
    expect(ids).toEqual([]);
  });
});

describe('getAvailableParents', () => {
  it('excludes self and descendants', () => {
    const tree = buildCategoryTree(flatCategories);
    const parents = getAvailableParents(tree, '1'); // Clothing

    // Should not include Clothing itself or any descendants
    const ids = parents.map((p) => p.id);
    expect(ids).not.toContain('1'); // self
    expect(ids).not.toContain('3'); // T-shirts (child)
    expect(ids).not.toContain('4'); // Jeans (child)
    expect(ids).not.toContain('6'); // Vintage T-shirts (grandchild)
    expect(ids).toContain('2'); // Shoes (sibling, available)
  });

  it('excludes self and descendants for nested node', () => {
    const tree = buildCategoryTree(flatCategories);
    const parents = getAvailableParents(tree, '3'); // T-shirts

    const ids = parents.map((p) => p.id);
    expect(ids).not.toContain('3'); // self
    expect(ids).not.toContain('6'); // Vintage T-shirts (child)
    expect(ids).toContain('1'); // Clothing (parent)
    expect(ids).toContain('2'); // Shoes (sibling)
  });

  it('returns all roots for leaf node', () => {
    const tree = buildCategoryTree(flatCategories);
    const parents = getAvailableParents(tree, '6'); // Vintage T-shirts (leaf)

    const ids = parents.map((p) => p.id);
    expect(ids).toContain('1');
    expect(ids).toContain('2');
  });
});

describe('toggleNodeExpanded', () => {
  it('toggles expanded state', () => {
    const tree = buildCategoryTree(flatCategories);
    expect(tree[0]!.expanded).toBe(false);

    const toggled = toggleNodeExpanded(tree, '1');
    expect(toggled[0]!.expanded).toBe(true);

    const toggledBack = toggleNodeExpanded(toggled, '1');
    expect(toggledBack[0]!.expanded).toBe(false);
  });

  it('does not affect other nodes', () => {
    const tree = buildCategoryTree(flatCategories);
    const toggled = toggleNodeExpanded(tree, '1');

    // Tree[1] (Shoes) should be unchanged
    expect(toggled[1]!.id).toBe('2');
    expect(toggled[1]!.expanded).toBe(false);
  });

  it('toggles nested node', () => {
    const tree = buildCategoryTree(flatCategories);
    const toggled = toggleNodeExpanded(tree, '3');

    const node = getCategoryById(toggled, '3');
    expect(node!.expanded).toBe(true);
  });
});
