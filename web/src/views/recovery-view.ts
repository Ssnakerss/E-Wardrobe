/**
 * Recovery view — password reset via recovery token.
 */

import { element, on } from '../dom/element.js';
import { navigate } from '../app/router.js';
import { validateRecovery, completeRecovery } from '../api/auth.js';
import { createToast } from '../components/toast.js';

let isLoading = false;

export class RecoveryView {
  private container: HTMLElement | null = null;
  private cleanup: (() => void)[] = [];
  private recoveryToken: string | null = null;

  async mount(container: HTMLElement): Promise<void> {
    this.container = container;
    this.recoveryToken = this.extractToken();
    this.render();

    document.title = 'Recover Account — E-Wardrobe';

    if (this.recoveryToken) {
      await this.validateRecovery();
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
    const token = params.get('token');
    if (token) {
      const url = new URL(window.location.href);
      url.searchParams.delete('token');
      window.history.replaceState({}, '', url.toString());
    }
    return token;
  }

  private async validateRecovery(): Promise<void> {
    const container = this.container;
    if (!container || !this.recoveryToken) return;

    const statusContainer = container.querySelector<HTMLDivElement>('#recovery-status');
    const errorContainer = container.querySelector<HTMLDivElement>('#recovery-error');

    if (!statusContainer || !errorContainer) return;

    try {
      await validateRecovery(this.recoveryToken);
      statusContainer.innerHTML = '';
      statusContainer.appendChild(
        element({
          tag: 'p',
          attrs: { class: 'form-success' },
          children: ['Valid recovery link. Set your new password below.'],
        }),
      );
    } catch (error: unknown) {
      if (error instanceof Error && 'kind' in error) {
        const e = error as Error & { kind: string; message: string; httpStatus: number };
        if (e.httpStatus === 410) {
          statusContainer.innerHTML = '';
          statusContainer.appendChild(
            element({
              tag: 'p',
              attrs: { class: 'form-error-text' },
              children: ['This recovery link is no longer valid.'],
            }),
          );
        } else {
          errorContainer.textContent = e.message || 'Invalid recovery link.';
        }
      }
    }
  }

  private render(): void {
    const container = this.container;
    if (!container) return;

    container.innerHTML = '';

    const mainHeading = element({
      tag: 'h1',
      attrs: { class: 'view-title' },
      children: ['Reset Password'],
    });

    container.appendChild(mainHeading);

    const statusContainer = element({
      tag: 'div',
      attrs: { id: 'recovery-status', class: 'form-status' },
    });

    container.appendChild(statusContainer);

    const form = element({
      tag: 'form',
      attrs: {
        id: 'recovery-form',
        noValidate: true,
      },
      children: [
        // New password
        element({
          tag: 'div',
          attrs: { class: 'form-group' },
          children: [
            element({
              tag: 'label',
              attrs: { for: 'recovery-password' },
              children: ['New Password'],
            }),
            element({
              tag: 'input',
              attrs: {
                type: 'password',
                id: 'recovery-password',
                name: 'new_password',
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

        // Confirm password
        element({
          tag: 'div',
          attrs: { class: 'form-group' },
          children: [
            element({
              tag: 'label',
              attrs: { for: 'recovery-confirm' },
              children: ['Confirm Password'],
            }),
            element({
              tag: 'input',
              attrs: {
                type: 'password',
                id: 'recovery-confirm',
                name: 'confirm_password',
                required: true,
                minlength: '12',
                maxlength: '128',
                autocomplete: 'new-password',
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
                id: 'recovery-submit',
              },
              children: ['Reset Password'],
            }),
          ],
        }),

        // Error container
        element({
          tag: 'div',
          attrs: {
            id: 'recovery-error',
            className: 'form-error',
            role: 'alert',
          },
        }),
      ],
    });

    container.appendChild(form);

    // Form submission handler
    const submitHandler = async (e: Event) => {
      e.preventDefault();
      if (isLoading) return;

      const formEl = container.querySelector<HTMLFormElement>('#recovery-form');
      if (!formEl) return;

      if (!formEl.checkValidity()) {
        formEl.reportValidity();
        return;
      }

      const passwordInput = formEl.querySelector<HTMLInputElement>('input[name="new_password"]');
      const confirmInput = formEl.querySelector<HTMLInputElement>('input[name="confirm_password"]');
      const errorContainer = container.querySelector<HTMLDivElement>('#recovery-error');

      if (!passwordInput || !confirmInput || !errorContainer) return;

      if (passwordInput.value !== confirmInput.value) {
        errorContainer.textContent = 'Passwords do not match.';
        return;
      }

      if (!this.recoveryToken) {
        errorContainer.textContent = 'No recovery token provided.';
        return;
      }

      isLoading = true;
      errorContainer.innerHTML = '';

      const submitBtn = formEl.querySelector<HTMLButtonElement>('#recovery-submit');
      const originalText = submitBtn?.textContent;
      if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Resetting...';
      }

      try {
        await completeRecovery({
          recovery_token: this.recoveryToken,
          new_password: passwordInput.value,
        });

        createToast({
          type: 'success',
          message: 'Password reset successfully. Please log in.',
        });

        navigate('/login');
      } catch (error: unknown) {
        if (error instanceof Error && 'kind' in error) {
          const apiError = error as { kind: string; message: string };
          errorContainer.textContent = apiError.message || 'Password reset failed.';
        } else {
          errorContainer.textContent = 'Network error. Please try again.';
        }
      } finally {
        isLoading = false;
        if (submitBtn) {
          submitBtn.disabled = false;
          submitBtn.textContent = originalText ?? 'Reset Password';
        }
      }
    };

    this.cleanup.push(on(form, 'submit', submitHandler));
  }
}
