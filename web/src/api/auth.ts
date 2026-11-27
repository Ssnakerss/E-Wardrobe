/**
 * Auth API endpoints.
 */

import { apiPost, apiDelete } from './client.js';
import type { User } from './auth-store.js';

export interface InvitationValidateResponse {
  email: string | null;
  expires_at: string;
}

export interface AuthResponse {
  user: User;
  expires_at: string;
  csrf_token?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
  client: 'web' | 'mobile';
  device_name?: string | null;
}

export interface RegisterRequest {
  invitation_token: string;
  email: string;
  password: string;
  client: 'web' | 'mobile';
  device_name?: string | null;
}

export interface RecoveryValidateRequest {
  recovery_token: string;
}

export interface RecoveryValidateResponse {
  user_id: string;
}

export interface RecoveryCompleteRequest {
  recovery_token: string;
  new_password: string;
}

/**
 * Validate an invitation token.
 */
export async function validateInvitation(token: string): Promise<InvitationValidateResponse> {
  return apiPost<InvitationValidateResponse>('/api/auth/invitations/validate', {
    invitation_token: token,
  });
}

/**
 * Register with an invitation token.
 */
export async function register(data: RegisterRequest): Promise<AuthResponse> {
  return apiPost<AuthResponse>('/api/auth/register', data);
}

/**
 * Login with email and password.
 */
export async function login(data: LoginRequest): Promise<AuthResponse> {
  return apiPost<AuthResponse>('/api/auth/login', data);
}

/**
 * Logout — revokes the session.
 */
export async function logout(): Promise<void> {
  await apiDelete('/api/auth/logout');
}

/**
 * Validate a recovery token.
 */
export async function validateRecovery(token: string): Promise<RecoveryValidateResponse> {
  return apiPost<RecoveryValidateResponse>('/api/auth/recovery/validate', {
    recovery_token: token,
  });
}

/**
 * Complete password reset with a recovery token.
 */
export async function completeRecovery(data: RecoveryCompleteRequest): Promise<void> {
  await apiPost('/api/auth/recovery/complete', data);
}
