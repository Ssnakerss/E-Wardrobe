/**
 * Reusable category selector dropdown.
 * Used in item forms to select a category.
 * Loads options on first focus/click.
 */

import { element, on } from '../dom/element.js';
import { fetchCategories } from '../api/categories.js';
import { buildCategoryTree, type CategoryNode } from '../domain/category-tree.js';

interface CategorySelectOptions {
  value?: string | null;
  onChange?: (categoryId: string | null) => void;
}

export function createCategorySelect(opts: CategorySelectOptions = {}): HTMLElement {
  const select = element({
    tag: 'select',
    attrs: {
      className: 'form-control',
    },
    children: [
      element({
        tag: 'option',
        attrs: { value: '' },
        children: ['— None —'],
      }),
    ],
  }) as HTMLSelectElement;

  let loaded = false;

  async function loadOptions(): Promise<void> {
    if (loaded) return;

    try {
      const categories = await fetchCategories();
      const tree = buildCategoryTree(categories);
      appendTreeOptions(tree, select, 0);
      loaded = true;
    } catch {
      // Silently fail — select stays empty
    }
  }

  function appendTreeOptions(
    nodes: CategoryNode[],
    parent: HTMLSelectElement,
    depth: number,
  ): void {
    for (const node of nodes) {
      const indent = '—'.repeat(depth);
      const label = depth > 0 ? `${indent} ${node.name}` : node.name;

      const option = element({
        tag: 'option',
        attrs: {
          value: node.id,
        },
        children: [label],
      }) as HTMLOptionElement;

      parent.appendChild(option);

      if (node.children.length > 0) {
        appendTreeOptions(node.children, parent, depth + 1);
      }
    }
  }

  // Set initial value
  if (opts.value) {
    select.value = opts.value;
  }

  // Load options on first focus
  on(select, 'focus', () => {
    void loadOptions();
  });

  // Also load on click in case focus doesn't trigger
  on(select, 'click', () => {
    void loadOptions();
  });

  // Handle change
  const onChange = opts.onChange;
  if (onChange) {
    on(select, 'change', () => {
      const val = select.value;
      onChange(val === '' ? null : val);
    });
  }

  return select;
}
