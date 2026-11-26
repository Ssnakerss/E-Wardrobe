/**
 * Category tree navigation component.
 * Renders an expandable/collapsible tree with action buttons per node.
 */

import { element, on, clearElement } from '../dom/element.js';
import { createLoadingState } from './loading-state.js';
import { errorStateFromApiError } from './error-state.js';
import { fetchCategories, createCategory, updateCategory, deleteCategory } from '../api/categories.js';
import { buildCategoryTree, getCategoryById, getAllDescendantIds, getAvailableParents, toggleNodeExpanded, type CategoryNode, type CategoryFlat } from '../domain/category-tree.js';
import type { View } from '../app/lifecycle.js';
import type { ApiError } from '../api/errors.js';

interface CategoryNavOptions {
  onCategorySelect?: (categoryId: string) => void;
}

export function createCategoryNav(opts: CategoryNavOptions = {}): View {
  let abortController: AbortController | null = null;
  let categories: CategoryFlat[] = [];
  let tree: CategoryNode[] = [];
  let error: ApiError | null = null;
  let loading = false;

  const containerRef: { current: HTMLElement | null } = { current: null };

  async function loadCategories(): Promise<void> {
    if (abortController) {
      abortController.abort();
    }

    abortController = new AbortController();
    loading = true;
    error = null;
    render(containerRef.current);

    try {
      categories = await fetchCategories();
      tree = buildCategoryTree(categories);
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }
      error = err as ApiError;
    } finally {
      loading = false;
      render(containerRef.current);
    }
  }

  async function handleCreate(parentId: string | null): Promise<void> {
    const name = prompt('Enter category name:');
    if (!name || name.trim().length === 0) return;

    try {
      const newCat = await createCategory({
        name: name.trim(),
        parent_id: parentId,
        sort_order: categories.filter((c) => c.parent_id === parentId).length,
      });
      categories.push(newCat);
      tree = buildCategoryTree(categories);
      render(containerRef.current);
    } catch (err) {
      error = err as ApiError;
      render(containerRef.current);
    }
  }

  async function handleRename(nodeId: string): Promise<void> {
    const node = getCategoryById(tree, nodeId);
    if (!node) return;

    const newName = prompt('Enter new name:', node.name);
    if (!newName || newName.trim().length === 0 || newName.trim() === node.name) return;

    try {
      const updated = await updateCategory(nodeId, {
        base_version: node.version,
        fields: { name: newName.trim() },
      });
      const idx = categories.findIndex((c) => c.id === nodeId);
      if (idx !== -1) {
        categories[idx] = updated;
      }
      tree = buildCategoryTree(categories);
      render(containerRef.current);
    } catch (err) {
      error = err as ApiError;
      render(containerRef.current);
    }
  }

  async function handleMove(nodeId: string, newParentId: string | null): Promise<void> {
    const node = getCategoryById(tree, nodeId);
    if (!node) return;

    const descendantIds = getAllDescendantIds(tree, nodeId);
    if (newParentId !== null && descendantIds.includes(newParentId)) {
      alert('Cannot move a category into its own subtree.');
      return;
    }

    try {
      const updated = await updateCategory(nodeId, {
        base_version: node.version,
        fields: { parent_id: newParentId },
      });
      const idx = categories.findIndex((c) => c.id === nodeId);
      if (idx !== -1) {
        categories[idx] = updated;
      }
      tree = buildCategoryTree(categories);
      render(containerRef.current);
    } catch (err) {
      error = err as ApiError;
      render(containerRef.current);
    }
  }

  async function handleDelete(nodeId: string): Promise<void> {
    const node = getCategoryById(tree, nodeId);
    if (!node) return;

    const confirmed = confirm(`Delete "${node.name}"? This cannot be undone.`);
    if (!confirmed) return;

    try {
      await deleteCategory(nodeId, node.version);
      categories = categories.filter((c) => c.id !== nodeId);
      tree = buildCategoryTree(categories);
      render(containerRef.current);
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        return;
      }
      error = err as ApiError;
      render(containerRef.current);
    }
  }

  function handleToggleExpand(nodeId: string): void {
    tree = toggleNodeExpanded(tree, nodeId);
    render(containerRef.current);
  }

  function renderNodeContent(node: CategoryNode): HTMLElement {
    const hasChildren = node.children.length > 0;
    const isExpanded = node.expanded ?? false;
    const isArchived = node.archived_at !== null;

    const actions: (Node | string)[] = [];

    // Toggle expand button
    if (hasChildren) {
      actions.push(
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: 'btn btn--toggle',
            'aria-label': isExpanded ? 'Collapse' : 'Expand',
          },
          children: [isExpanded ? '▾' : '▸'],
        }),
      );
    } else {
      actions.push(element({ tag: 'span', attrs: { className: 'category-spacer' } }));
    }

    // Category name with archived indicator
    const nameParts: (Node | string)[] = [];
    if (isArchived) {
      nameParts.push(
        element({
          tag: 'span',
          attrs: { className: 'category-name category-name--archived' },
          children: ['[Archived] '],
        }),
      );
    }
    nameParts.push(node.name);

    actions.push(
      element({
        tag: 'span',
        attrs: {
          className: 'category-name',
        },
        children: nameParts,
      }),
    );

    // Action buttons (disabled for archived)
    const baseBtnAttrs: Record<string, string | boolean | null> = {
      type: 'button',
      className: '',
      'aria-label': '',
    };
    if (isArchived) {
      baseBtnAttrs.disabled = '';
    }

    actions.push(
      element({ tag: 'button', attrs: { ...baseBtnAttrs, className: 'btn btn--icon btn--rename', 'aria-label': 'Rename' }, children: ['✎'] }),
      element({ tag: 'button', attrs: { ...baseBtnAttrs, className: 'btn btn--icon btn--move', 'aria-label': 'Move' }, children: ['⇄'] }),
      element({ tag: 'button', attrs: { ...baseBtnAttrs, className: 'btn btn--icon btn--delete', 'aria-label': 'Delete' }, children: ['✕'] }),
    );

    const row = element({
      tag: 'div',
      attrs: {
        className: 'category-row',
      },
      children: actions,
    });

    // Attach event handlers
    const toggleBtn = row.querySelector('.btn--toggle');
    if (toggleBtn) {
      on(toggleBtn as HTMLElement, 'click', () => {
        handleToggleExpand(node.id);
      });
    }

    const renameBtn = row.querySelector('.btn--rename');
    if (renameBtn) {
      on(renameBtn as HTMLElement, 'click', () => {
        void handleRename(node.id);
      });
    }

    const moveBtn = row.querySelector('.btn--move');
    if (moveBtn) {
      on(moveBtn as HTMLElement, 'click', () => {
        void handleMovePrompt(node);
      });
    }

    const deleteBtn = row.querySelector('.btn--delete');
    if (deleteBtn) {
      on(deleteBtn as HTMLElement, 'click', () => {
        void handleDelete(node.id);
      });
    }

    // Click on name to select
    const nameSpan = row.querySelector('.category-name');
    if (nameSpan && opts.onCategorySelect) {
      on(nameSpan as HTMLElement, 'click', () => {
        opts.onCategorySelect!(node.id);
      });
    }

    return row;
  }

  function renderTree(nodes: CategoryNode[], depth: number): HTMLElement {
    const ul = element({
      tag: 'ul',
      attrs: { className: 'category-tree' },
      children: nodes.map((node) => renderNode(node, depth)),
    });
    return ul;
  }

  function renderNode(node: CategoryNode, depth: number): HTMLElement {
    const li = element({
      tag: 'li',
      attrs: {
        className: `category-node category-node--depth-${depth}`,
        'data-category-id': node.id,
      },
      children: [
        renderNodeContent(node),
      ],
    });

    if (node.children.length > 0 && node.expanded) {
      li.appendChild(renderTree(node.children, depth + 1));
    }

    return li;
  }

  function handleMovePrompt(node: CategoryNode): void {
    const availableParents = getAvailableParents(tree, node.id);
    if (availableParents.length === 0) {
      alert('No valid parent categories available.');
      return;
    }

    const options = availableParents
      .map((p: CategoryFlat) => `${p.parent_id === null ? '(root)' : p.name}: ${p.name}`)
      .join('\n');

    const choice = prompt(
      `Move "${node.name}" to:\n\n${options}\n\nEnter the category name or "(root)" for root.`,
    );

    if (!choice) return;

    const trimmed = choice.trim();
    let newParentId: string | null = null;

    if (trimmed === '(root)') {
      newParentId = null;
    } else {
      const found = availableParents.find((p: CategoryFlat) => p.name === trimmed);
      if (found) {
        newParentId = found.id;
      } else {
        alert('Category not found. Try again.');
        return;
      }
    }

    void handleMove(node.id, newParentId);
  }

  function render(mainEl: HTMLElement | null): void {
    if (!mainEl) return;
    clearElement(mainEl);

    if (loading && tree.length === 0) {
      mainEl.appendChild(createLoadingState('Loading categories...'));
      return;
    }

    if (error) {
      mainEl.appendChild(errorStateFromApiError(error, 'Retry', () => {
        void loadCategories();
      }));
      return;
    }

    if (tree.length === 0) {
      mainEl.appendChild(
        element({
          tag: 'div',
          attrs: { className: 'empty-state' },
          children: [
            element({
              tag: 'h2',
              className: 'empty-state__title',
              children: ['No categories'],
            }),
            element({
              tag: 'p',
              children: ['Create your first category to organize items.'],
            }),
            element({
              tag: 'button',
              attrs: {
                type: 'button',
                className: 'btn btn--primary',
              },
              children: ['Create category'],
            }),
          ],
        }),
      );
      return;
    }

    const wrapper = element({
      tag: 'div',
      attrs: { className: 'category-nav' },
      children: [
        element({
          tag: 'h1',
          attrs: { className: 'view-title' },
          children: ['Categories'],
        }),
        element({
          tag: 'button',
          attrs: {
            type: 'button',
            className: 'btn btn--primary',
          },
          children: ['+ New category'],
        }),
        renderTree(tree, 0),
      ],
    });

    mainEl.appendChild(wrapper);

    // Attach create button handler
    const createBtn = wrapper.querySelector('button.btn--primary');
    if (createBtn) {
      on(createBtn as HTMLElement, 'click', () => {
        void handleCreate(null);
      });
    }
  }

  return {
    mount(container: HTMLElement): void {
      containerRef.current = container;
      void loadCategories();
    },
    destroy(): void {
      if (abortController) {
        abortController.abort();
        abortController = null;
      }
    },
  };
}
