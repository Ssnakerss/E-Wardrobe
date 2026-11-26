/**
 * Registration view with invitation token validation.
 */

import { element, on } from '../dom/element.js';
import { navigate } from '../app/router.js';
import { validateInvitation, register } from '../api/auth.js';
import { setAuthenticated } from '../api/auth-store.js';
import { createToast } from '../components/toast.js';

let isLoading = false;

export class RegisterView {
  private container: HTMLElement | null = null;
  private cleanup: (() => void)[] = [];
  private invitationToken: string | null = null;

  async mount(container: HTMLElement): Promise<void> {
    this.container = container;
    this.invitationToken = this.extractToken();
    this.render();
    document.title = 'Register - E-Wardrobe';
    if (this.invitationToken) {
      await this.validateInvitation();
    }
  }

  destroy(): void {
    for (const fn of this.cleanup) {
      fn();
    }
    this.cleanup = [];
    this.container = null;
    isLoading = false;
  }

  private extractToken(): string | null {
    const params = new URLSearchParams(window.location.search);
    const token = params.get('invite');
    if (token) {
      const url = new URL(window.location.href);
      url.searchParams.delete('invite');
      window.history.replaceState({}, '', url.toString());
    }
    return token;
  }

  private async validateInvitation(): Promise<void> {
    const container = this.container;
    if (!container || !this.invitationToken) return;

    const errorContainer = container.querySelector<HTMLDivElement>('#register-error');
    const statusContainer = container.querySelector<HTMLDivElement>('#register-status');
    if (!errorContainer || !statusContainer) return;

    try {
      const result = await validateInvitation(this.invitationToken);
      statusContainer.innerHTML = '';
      statusContainer.appendChild(
        element({
          tag: 'p',
          attrs: { class: 'form-success' },
          children: ['Valid invitation'],
        }),
      );

      if (result.email) {
        const emailInput = container.querySelector<HTMLInputElement>('input[name="email"]');
        if (emailInput) {
          emailInput.value = result.email;
          emailInput.disabled = true;
        }
      }
    } catch (error: unknown) {
      if (error instanceof Error && 'kind' in error) {
        const e = error as Error & { kind: string; message: string; httpStatus: number };
        if (e.httpStatus === 410) {
          statusContainer.innerHTML = '';
          statusContainer.appendChild(
            element({
              tag: 'p',
              attrs: { class: 'form-error-text' },
              children: ['This invitation is no longer valid.'],
            }),
          );
        } else {
          errorContainer.textContent = e.message || 'Invalid invitation.';
        }
      }
    }
  }

  private render(): void {
    const container = this.container;
    if (!container) return;

    container.innerHTML = '';

    container.appendChild(
      element({
        tag: 'h1',
        attrs: { class: 'view-title' },
        children: ['Create Account'],
      }),
    );

    const statusContainer = element({
      tag: 'div',
      attrs: { id: 'register-status', class: 'form-status' },
    });
    container.appendChild(statusContainer);

    const form = element({
      tag: 'form',
      attrs: { id: 'register-form', noValidate: true },
      children: [
        element({
          tag: 'div',
          attrs: { class: 'form-group' },
          children: [
            element({ tag: 'label', attrs: { for: 'register-email' }, children: ['Email'] }),
            element({
              tag: 'input',
              attrs: {
                type: 'email',
                id: 'register-email',
                name: 'email',
                required: true,
                autocomplete: 'email',
              },
            }),
          ],
        }),
        element({
          tag: 'div',
          attrs: { class: 'form-group' },
          children: [
            element({ tag: 'label', attrs: { for: 'register-password' }, children: ['Password'] }),
            element({
              tag: 'input',
              attrs: {
                type: 'password',
                id: 'register-password',
                name: 'password',
                required: true,
                minlength: '12',
                maxlength: '128',
                autocomplete: 'new-password',
              },
            }),
            element({
              tag: 'small',
              attrs: { class: 'form-hint' },
              children: ['12-128 characters'],
            }),
          ],
        }),
        element({
          tag: 'div',
          attrs: { class: 'form-group' },
          children: [
            element({
              tag: 'label',
              attrs: { for: 'register-confirm' },
              children: ['Confirm Password'],
            }),
            element({
              tag: 'input',
              attrs: {
                type: 'password',
                id: 'register-confirm',
                name: 'confirm_password',
                required: true,
                minlength: '12',
                maxlength: '128',
                autocomplete: 'new-password',
              },
            }),
          ],
        }),
        element({
          tag: 'div',
          attrs: { class: 'form-actions' },
          children: [
            element({
              tag: 'button',
              attrs: { type: 'submit', className: 'btn btn--primary', id: 'register-submit' },
              children: ['Create Account'],
            }),
          ],
        }),
        element({
          tag: 'div',
          attrs: { id: 'register-error', className: 'form-error', role: 'alert' },
        }),
      ],
    });

    container.appendChild(form);

    const submitHandler = async (e: Event) => {
      e.preventDefault();
      if (isLoading) return;

      const formEl = container.querySelector<HTMLFormElement>('#register-form');
      if (!formEl) return;

      if (!formEl.checkValidity()) {
        formEl.reportValidity();
        return;
      }

      const emailInput = formEl.querySelector<HTMLInputElement>('input[name="email"]');
      const passwordInput = formEl.querySelector<HTMLInputElement>('input[name="password"]');
      const confirmInput = formEl.querySelector<HTMLInputElement>('input[name="confirm_password"]');
      const errorContainer = container.querySelector<HTMLDivElement>('#register-error');

      if (!emailInput || !passwordInput || !confirmInput || !errorContainer) return;

      if (passwordInput.value !== confirmInput.value) {
        errorContainer.textContent = 'Passwords do not match.';
        return;
      }

      if (!this.invitationToken) {
        errorContainer.textContent = 'No invitation token provided.';
        return;
      }

      const email = emailInput.value.trim().toLowerCase();
      const password = passwordInput.value;

      isLoading = true;
      errorContainer.innerHTML = '';

      const submitBtn = formEl.querySelector<HTMLButtonElement>('#register-submit');
      const originalText = submitBtn?.textContent;
      if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Creating account...';
      }

      try {
        const response = await register({
          invitation_token: this.invitationToken,
          email,
          password,
          client: 'web',
        });

        setAuthenticated(response.user);
        createToast({ type: 'success', message: 'Account created successfully' });
        navigate('/items');
      } catch (error: unknown) {
        if (error instanceof Error && 'kind' in error) {
          const e = error as Error & { message: string };
          errorContainer.textContent = e.message || 'Registration failed.';
        } else {
          errorContainer.textContent = 'Network error. Please try again.';
        }
      } finally {
        isLoading = false;
        if (submitBtn) {
          submitBtn.disabled = false;
          submitBtn.textContent = originalText ?? 'Create Account';
        }
      }
    };

    this.cleanup.push(on(form, 'submit', submitHandler));
  }
}
