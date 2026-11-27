/**
 * Redirect view — navigates to a target path on mount.
 */

import type { View } from '../app/lifecycle.js';
import { navigate } from '../app/router.js';

export class RedirectView implements View {
  private targetPath: string;

  constructor(targetPath: string) {
    this.targetPath = targetPath;
  }

  mount(container: HTMLElement): void {
    container.innerHTML = '';
    navigate(this.targetPath);
  }

  destroy(): void {}
}
