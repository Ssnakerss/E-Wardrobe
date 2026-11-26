/**
 * Route definitions — maps paths to guards and loaders.
 */

import { addRoute } from './router.js';
import type { View } from './lifecycle.js';

/**
 * Register all application routes.
 * Views are loaded lazily to avoid loading unused code.
 */
export function registerRoutes(): void {
  // Public routes
  addRoute(
    '/login',
    () => true,
    () => import('../views/login-view.js').then((m) => new m.LoginView() as View),
  );
  addRoute(
    '/register',
    () => true,
    () => import('../views/register-view.js').then((m) => new m.RegisterView() as View),
  );
  addRoute(
    '/recovery',
    () => true,
    () => import('../views/recovery-view.js').then((m) => new m.RecoveryView() as View),
  );

  // Authenticated routes
  addRoute(
    '/items',
    () => true,
    () => import('../views/items-view.js').then((m) => new m.ItemsView() as View),
  );
  addRoute(
    '/items/new',
    () => true,
    () => import('../views/item-form-view.js').then((m) => new m.ItemFormView() as View),
  );
  addRoute(
    '/items/:itemId',
    () => true,
    () => import('../views/item-details-view.js').then((m) => new m.ItemDetailsView() as View),
  );
  addRoute(
    '/items/:itemId/edit',
    () => true,
    () => import('../views/item-form-view.js').then((m) => new m.ItemFormView() as View),
  );
  addRoute(
    '/categories',
    () => true,
    () => import('../views/categories-view.js').then((m) => new m.CategoriesView() as View),
  );
  addRoute(
    '/archive',
    () => true,
    () => import('../views/archive-view.js').then((m) => new m.ArchiveView() as View),
  );

  // 404 fallback
  addRoute(
    '/*',
    () => true,
    () => import('../views/not-found-view.js').then((m) => new m.NotFoundView() as View),
  );
}
