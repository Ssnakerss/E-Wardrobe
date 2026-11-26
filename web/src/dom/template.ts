/**
 * Template helpers — render static HTML from <template> elements.
 */

/**
 * Clone a <template> by selector and return the content fragment.
 */
export function cloneTemplate(selector: string): DocumentFragment {
  const tpl = document.querySelector<HTMLTemplateElement>(selector);
  if (!tpl) {
    throw new Error(`Template not found: ${selector}`);
  }
  return tpl.content.cloneNode(true) as DocumentFragment;
}

/**
 * Render a template with interpolated values.
 * Values are always set via textContent to prevent XSS.
 */
export function renderTemplate(
  selector: string,
  values: Record<string, string | undefined>,
): HTMLElement {
  const frag = cloneTemplate(selector);
  const container = frag.firstElementChild as HTMLElement | null;
  if (!container) {
    throw new Error(`Template has no root element: ${selector}`);
  }

  for (const [key, value] of Object.entries(values)) {
    if (value === undefined) continue;
    const placeholder = container.querySelector(`[data-${key}]`);
    if (placeholder) {
      placeholder.textContent = value;
    }
  }

  return container;
}
