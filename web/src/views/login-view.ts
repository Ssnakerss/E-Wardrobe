/**
 * Login view — email + password form.
 */

import { element, on } from '../dom/element.js';
import { navigate } from '../app/router.js';
import { login } from '../api/auth.js';
import { setAuthenticated } from '../api/auth-store.js';
import { createToast } from '../components/toast.js';

let isLoading = false;

export class LoginView {
  private container: HTMLElement | null = null;
  private cleanup: (() => void)[] = [];

  async mount(container: HTMLElement): Promise<void> {
    this.container = container;
    this.render();

    // Set document title
    document.title = 'Login — E-Wardrobe';

    // Move focus to form
    setTimeout(() => {
      const emailInput = container.querySelector<HTMLInputElement>('input[type="email"]');
      emailInput?.focus();
    }, 100);
  }

  destroy(): void {
    for (const fn of this.cleanup) {
      fn();
    }
    this.cleanup = [];
    this.container = null;
    isLoading = false;
  }

  private render(): void {
    const container = this.container;
    if (!container) return;

    container.innerHTML = '';

    const mainHeading = element({
      tag: 'h1',
      attrs: { class: 'view-title' },
      children: ['Login'],
    });

    container.appendChild(mainHeading);

    const form = element({
      tag: 'form',
      attrs: {
        id: 'login-form',
        noValidate: true,
      },
      children: [
        // Email field
        element({
          tag: 'div',
          attrs: { class: 'form-group' },
          children: [
            element({
              tag: 'label',
              attrs: { for: 'login-email' },
              children: ['Email'],
            }),
            element({
              tag: 'input',
              attrs: {
                type: 'email',
                id: 'login-email',
                name: 'email',
                required: true,
                autocomplete: 'email',
              },
            }),
          ],
        }),

        // Password field
        element({
          tag: 'div',
          attrs: { class: 'form-group' },
          children: [
            element({
              tag: 'label',
              attrs: { for: 'login-password' },
              children: ['Password'],
            }),
            element({
              tag: 'input',
              attrs: {
                type: 'password',
                id: 'login-password',
                name: 'password',
                required: true,
                autocomplete: 'current-password',
              },
            }),
          ],
        }),

        // Submit button
        element({
          tag: 'div',
          attrs: { class: 'form-actions' },
          children: [
            element({
              tag: 'button',
              attrs: {
                type: 'submit',
                className: 'btn btn--primary',
                id: 'login-submit',
              },
              children: ['Login'],
            }),
          ],
        }),

        // Error container
        element({
          tag: 'div',
          attrs: {
            id: 'login-error',
            className: 'form-error',
            role: 'alert',
          },
        }),

        // Recovery link
        element({
          tag: 'p',
          attrs: { class: 'form-help' },
          children: [
            'Forgot access? ',
            element({
              tag: 'a',
              attrs: { href: '/recovery' },
              children: ['Recover account'],
            }),
          ],
        }),
      ],
    });

    container.appendChild(form);

    // Form submission handler
    const submitHandler = async (e: Event) => {
      e.preventDefault();
      if (isLoading) return;

      const formEl = container.querySelector<HTMLFormElement>('#login-form');
      if (!formEl) return;

      // HTML5 validation
      if (!formEl.checkValidity()) {
        formEl.reportValidity();
        return;
      }

      const emailInput = formEl.querySelector<HTMLInputElement>('input[name="email"]');
      const passwordInput = formEl.querySelector<HTMLInputElement>('input[name="password"]');
      const errorContainer = container.querySelector<HTMLDivElement>('#login-error');

      if (!emailInput || !passwordInput || !errorContainer) return;

      // Normalize email (lowercase)
      const email = emailInput.value.trim().toLowerCase();
      const password = passwordInput.value;

      isLoading = true;
      errorContainer.innerHTML = '';

      // Show loading
      const submitBtn = formEl.querySelector<HTMLButtonElement>('#login-submit');
      const originalText = submitBtn?.textContent;
      if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Logging in...';
      }

      try {
        const response = await login({
          email,
          password,
          client: 'web',
        });

        setAuthenticated(response.user);
        createToast({
          type: 'success',
          message: 'Logged in successfully',
        });

        // Check for return_url in current query
        const params = new URLSearchParams(window.location.search);
        const returnUrl = params.get('return_url');
        const redirectPath = returnUrl && isSafeUrl(returnUrl) ? returnUrl : '/items';
        navigate(redirectPath);
      } catch (error: unknown) {
        if (error instanceof Error && 'kind' in error) {
          const apiError = error as { kind: string; message: string };
          if (apiError.kind === 'rate_limited') {
            errorContainer.textContent = 'Too many attempts. Please wait a moment and try again.';
          } else {
            errorContainer.textContent =
              apiError.message || 'Login failed. Please check your credentials.';
          }
        } else {
          errorContainer.textContent = 'Network error. Please check your connection.';
        }
      } finally {
        isLoading = false;
        if (submitBtn) {
          submitBtn.disabled = false;
          submitBtn.textContent = originalText ?? 'Login';
        }
      }
    };

    this.cleanup.push(on(form, 'submit', submitHandler));
  }
}

function isSafeUrl(url: string): boolean {
  try {
    const parsed = new URL(url, window.location.origin);
    return parsed.origin === window.location.origin;
  } catch {
    return false;
  }
}
