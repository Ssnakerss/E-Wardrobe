/**
 * DOM element helper — creates elements with attributes and children safely.
 * Uses textContent for all text to prevent XSS.
 */

interface ElementOptions {
  tag: string;
  attrs?: Record<string, string | boolean | null>;
  children?: (Node | string | null | undefined)[];
  className?: string;
}

export function element(opts: ElementOptions): HTMLElement {
  const el = document.createElement(opts.tag);

  if (opts.className) {
    el.className = opts.className;
  }

  if (opts.attrs) {
    for (const [key, value] of Object.entries(opts.attrs)) {
      if (value === null || value === false) continue;
      if (key === 'disabled' && value === true) {
        el.setAttribute('disabled', '');
      } else if (key === 'hidden' && value === true) {
        el.hidden = true;
      } else if (key.startsWith('on')) {
        // Skip event handlers — use on() instead
        continue;
      } else {
        el.setAttribute(key, String(value));
      }
    }
  }

  if (opts.children) {
    for (const child of opts.children) {
      if (child === null || child === undefined) continue;
      if (typeof child === 'string') {
        el.appendChild(document.createTextNode(child));
      } else {
        el.appendChild(child);
      }
    }
  }

  return el;
}

/**
 * Attach event listener and return a cleanup function.
 */
export function on<T extends EventTarget>(
  target: T,
  event: string,
  handler: (e: Event) => void,
  options?: boolean | AddEventListenerOptions,
): () => void {
  target.addEventListener(event, handler as EventListener, options);
  return () => {
    target.removeEventListener(event, handler as EventListener, options);
  };
}

/**
 * Attach event listener with delegation to a selector on the target.
 */
export function onDelegated<T extends EventTarget>(
  target: T,
  event: string,
  selector: string,
  handler: (e: Event, target: Element) => void,
): () => void {
  const wrapped = (e: Event) => {
    const closest = (e.target as HTMLElement)?.closest(selector);
    if (closest) {
      handler(e, closest);
    }
  };
  target.addEventListener(event, wrapped as EventListener);
  return () => {
    target.removeEventListener(event, wrapped as EventListener);
  };
}

/**
 * Set text content safely.
 */
export function setText(el: HTMLElement | Text, text: string | null | undefined): void {
  if (text === null || text === undefined) {
    if (el instanceof Text) {
      el.replaceWith(document.createTextNode(''));
    } else {
      el.textContent = '';
    }
  } else {
    if (el instanceof Text) {
      el.textContent = text;
    } else {
      el.textContent = text;
    }
  }
}

/**
 * Remove element from DOM and optionally release it.
 */
export function removeElement(el: HTMLElement): void {
  el.remove();
}

/**
 * Clear all children of an element.
 */
export function clearElement(el: HTMLElement): void {
  while (el.firstChild) {
    el.removeChild(el.firstChild);
  }
}
