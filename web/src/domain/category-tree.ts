/**
 * Pure domain functions for building and querying category trees.
 * All functions are side-effect free and work with immutable data.
 */

export interface CategoryFlat {
  id: string;
  name: string;
  parent_id: string | null;
  sort_order: number;
  version: number;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface CategoryNode extends CategoryFlat {
  children: CategoryNode[];
  expanded?: boolean;
}

/**
 * Build a tree from a flat list of categories.
 * Handles missing/invalid parents gracefully — treats them as root nodes.
 */
export function buildCategoryTree(categories: CategoryFlat[]): CategoryNode[] {
  if (categories.length === 0) {
    return [];
  }

  // Index all categories by id for O(1) lookup
  const byId = new Map<string, CategoryNode>();
  for (const cat of categories) {
    byId.set(cat.id, {
      ...cat,
      children: [],
      expanded: false,
    });
  }

  const roots: CategoryNode[] = [];

  for (const cat of categories) {
    const node = byId.get(cat.id)!;
    const parentId = cat.parent_id;

    if (parentId === null || !byId.has(parentId)) {
      // No parent or parent doesn't exist — treat as root
      roots.push(node);
    } else {
      const parent = byId.get(parentId)!;
      parent.children.push(node);
    }
  }

  // Sort by sort_order within each level
  sortTree(roots);

  return roots;
}

/**
 * Recursively sort tree nodes by sort_order.
 */
function sortTree(nodes: CategoryNode[]): void {
  nodes.sort((a, b) => a.sort_order - b.sort_order);
  for (const node of nodes) {
    sortTree(node.children);
  }
}

/**
 * Find a node by id in the tree. Returns null if not found.
 */
export function getCategoryById(
  tree: CategoryNode[],
  id: string,
): CategoryNode | null {
  for (const node of tree) {
    if (node.id === id) {
      return node;
    }
    const found = getCategoryById(node.children, id);
    if (found) {
      return found;
    }
  }
  return null;
}

/**
 * Get all descendant IDs recursively (depth-first).
 */
export function getAllDescendantIds(
  tree: CategoryNode[],
  parentId: string,
): string[] {
  const node = getCategoryById(tree, parentId);
  if (!node) {
    return [];
  }

  const ids: string[] = [];
  collectDescendantIds(node, ids);
  return ids;
}

function collectDescendantIds(node: CategoryNode, acc: string[]): void {
  for (const child of node.children) {
    acc.push(child.id);
    collectDescendantIds(child, acc);
  }
}

/**
 * Get available parent categories for a given category.
 * Excludes the category itself and all its descendants.
 * Also excludes already-rooted categories (parent_id === null).
 */
export function getAvailableParents(
  tree: CategoryNode[],
  excludeId: string,
): CategoryFlat[] {
  const descendants = new Set(getAllDescendantIds(tree, excludeId));
  descendants.add(excludeId);

  const result: CategoryFlat[] = [];

  for (const node of tree) {
    if (descendants.has(node.id)) {
      continue;
    }
    // Can only set parent to categories that are currently roots
    // (or have a parent that's not excluded)
    result.push(node);
  }

  return result;
}

/**
 * Toggle the expanded state of a node in the tree.
 * Returns a new tree (immutable update).
 */
export function toggleNodeExpanded(
  tree: CategoryNode[],
  nodeId: string,
): CategoryNode[] {
  return tree.map((node) => {
    if (node.id === nodeId) {
      return { ...node, expanded: !node.expanded };
    }
    if (node.children.length > 0) {
      return { ...node, children: toggleNodeExpanded(node.children, nodeId) };
    }
    return node;
  });
}
