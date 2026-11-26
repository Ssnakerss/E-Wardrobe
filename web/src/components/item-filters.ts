/**
 * Item filters component — search, category, sort, filters.
 */

import { element, on } from '../dom/element.js';
import type { ItemsQueryParams } from '../api/items.js';

interface ItemFiltersProps {
  value: ItemsQueryParams;
  onChange: (params: ItemsQueryParams) => void;
  categories: { id: string; name: string }[];
}

function selectedAttr(value: string | undefined, optionValue: string): string | null {
  return value === optionValue ? '' : null;
}

export function createItemFilters(props: ItemFiltersProps): HTMLElement {
  const container = element({
    tag: 'div',
    className: 'item-filters',
  });

  // Search input
  const searchInput = element({
    tag: 'input',
    attrs: {
      type: 'search',
      className: 'item-filters__search',
      placeholder: 'Search items...',
      value: props.value.q ?? '',
    },
  });

  // Debounced search handler
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  on(searchInput, 'input', (e: Event) => {
    const value = (e.target as HTMLInputElement).value;
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
      props.onChange({ ...props.value, q: value || undefined, cursor: undefined });
      searchTimer = null;
    }, 300);
  });

  // Category select
  const catOptions = [
    element({ tag: 'option', attrs: { value: '' }, children: ['All categories'] }),
  ];
  for (const cat of props.categories) {
    catOptions.push(
      element({
        tag: 'option',
        attrs: { value: cat.id, selected: selectedAttr(props.value.category_id, cat.id) },
        children: [cat.name],
      }),
    );
  }

  const categorySelect = element({
    tag: 'select',
    className: 'item-filters__select',
    children: catOptions,
  });

  on(categorySelect, 'change', () => {
    const val = (categorySelect as HTMLSelectElement).value;
    props.onChange({ ...props.value, category_id: val || undefined, cursor: undefined });
  });

  // Status select
  const statusOptions = [
    element({ tag: 'option', attrs: { value: '' }, children: ['All statuses'] }),
    element({
      tag: 'option',
      attrs: { value: 'in_use', selected: selectedAttr(props.value.status, 'in_use') },
      children: ['In use'],
    }),
    element({
      tag: 'option',
      attrs: { value: 'stored', selected: selectedAttr(props.value.status, 'stored') },
      children: ['Stored'],
    }),
    element({
      tag: 'option',
      attrs: { value: 'needs_repair', selected: selectedAttr(props.value.status, 'needs_repair') },
      children: ['Needs repair'],
    }),
  ];

  const statusSelect = element({
    tag: 'select',
    className: 'item-filters__select',
    children: statusOptions,
  });

  on(statusSelect, 'change', () => {
    const val = (statusSelect as HTMLSelectElement).value;
    props.onChange({
      ...props.value,
      status: (val as ItemsQueryParams['status']) || undefined,
      cursor: undefined,
    });
  });

  // Sort select
  const sortOptions = [
    element({
      tag: 'option',
      attrs: { value: 'updated_desc', selected: selectedAttr(props.value.sort, 'updated_desc') },
      children: ['Last updated'],
    }),
    element({
      tag: 'option',
      attrs: { value: 'created_desc', selected: selectedAttr(props.value.sort, 'created_desc') },
      children: ['Date added'],
    }),
    element({
      tag: 'option',
      attrs: { value: 'name_asc', selected: selectedAttr(props.value.sort, 'name_asc') },
      children: ['Name A-Z'],
    }),
    element({
      tag: 'option',
      attrs: {
        value: 'purchase_year_desc',
        selected: selectedAttr(props.value.sort, 'purchase_year_desc'),
      },
      children: ['Year'],
    }),
  ];

  const sortSelect = element({
    tag: 'select',
    className: 'item-filters__select',
    children: sortOptions,
  });

  on(sortSelect, 'change', () => {
    const val = (sortSelect as HTMLSelectElement).value;
    props.onChange({ ...props.value, sort: val || undefined });
  });

  // Archived toggle
  const archivedCheckbox = element({
    tag: 'input',
    attrs: {
      type: 'checkbox',
      className: 'item-filters__checkbox',
      id: 'filter-archived',
      checked: props.value.archived ? '' : null,
    },
  });

  on(archivedCheckbox, 'change', () => {
    props.onChange({ ...props.value, archived: (archivedCheckbox as HTMLInputElement).checked });
  });

  // Reset button
  const resetBtn = element({
    tag: 'button',
    attrs: { type: 'button', className: 'btn btn--sm', 'data-action': 'reset' },
    children: ['Reset'],
  });

  on(resetBtn, 'click', () => {
    props.onChange({ sort: 'updated_desc' });
  });

  // Layout
  container.append(
    searchInput,
    categorySelect,
    statusSelect,
    sortSelect,
    archivedCheckbox,
    resetBtn,
  );

  return container;
}
