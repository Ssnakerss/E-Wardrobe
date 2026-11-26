/**
 * UI store — manages UI state like view mode, dialogs, toasts.
 */

export type ListViewMode = 'cards' | 'table';

export interface UiState {
  listMode: ListViewMode;
  toast: ToastMessage | null;
}

export interface ToastMessage {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
}

let state: UiState = {
  listMode: (loadListMode() as ListViewMode) || 'cards',
  toast: null,
};

const listeners: Set<(state: UiState) => void> = new Set();

function loadListMode(): string | null {
  try {
    return localStorage.getItem('wardrobe:list_mode');
  } catch {
    return null;
  }
}

function saveListMode(mode: ListViewMode): void {
  try {
    localStorage.setItem('wardrobe:list_mode', mode);
  } catch {
    // Storage unavailable — ignore
  }
}

export function getUiState(): UiState {
  return state;
}

export function subscribeUi(listener: (state: UiState) => void): () => void {
  listeners.add(listener);
  listener(state);
  return () => listeners.delete(listener);
}

function setState(patch: Partial<UiState>): void {
  state = { ...state, ...patch };
  for (const listener of listeners) {
    listener(state);
  }
}

/**
 * Set the list view mode.
 */
export function setListMode(mode: ListViewMode): void {
  saveListMode(mode);
  setState({ listMode: mode });
}

/**
 * Show a toast message.
 */
export function showToast(message: Omit<ToastMessage, 'id'>): string {
  const id = `toast-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  setState({ toast: { ...message, id } });
  return id;
}

/**
 * Dismiss a toast message.
 */
export function dismissToast(id: string): void {
  if (state.toast?.id === id) {
    setState({ toast: null });
  }
}
