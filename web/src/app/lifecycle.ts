/**
 * View lifecycle interface for managing view mounting/unmounting.
 */

export interface View {
  /**
   * Called when the view is mounted to a container.
   */
  mount(container: HTMLElement): Promise<void> | void;

  /**
   * Called when the view is unmounted.
   * Should clean up event listeners, abort pending requests, etc.
   */
  destroy(): void;
}
