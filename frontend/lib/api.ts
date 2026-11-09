import type { AuthResponse, RefreshResponse, User } from '../types/auth'

const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

export class ApiError extends Error {
  constructor(message: string, readonly status: number) {
    super(message)
    this.name = 'ApiError'
  }
}

async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}${path}`, {
      ...options,
      credentials: 'include',
      cache: 'no-store',
      headers: {
        ...(options.body && !(options.body instanceof FormData) ? { 'Content-Type': 'application/json' } : {}),
        ...options.headers,
      },
    })
  } catch {
    throw new ApiError('Не удалось связаться с сервером. Повторите попытку позже.', 0)
  }

  if (!response.ok) {
    const body = await response.json().catch(() => null) as { error?: string } | null
    throw new ApiError(body?.error ?? 'Не удалось выполнить запрос.', response.status)
  }

  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export const authApi = {
  register(email: string, login: string, password: string) {
    return apiRequest<AuthResponse>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, login, password }),
    })
  },
  login(identifier: string, password: string) {
    return apiRequest<AuthResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ identifier, password }),
    })
  },
  refresh() {
    return apiRequest<RefreshResponse>('/api/v1/auth/refresh', { method: 'POST' })
  },
  logout() {
    return apiRequest<void>('/api/v1/auth/logout', { method: 'POST' })
  },
}

export const usersApi = {
  me(accessToken: string) {
    return apiRequest<User>('/api/v1/users/me', {
      headers: { Authorization: `Bearer ${accessToken}` },
    })
  },
  updateName(accessToken: string, name: string) {
    return apiRequest<User>('/api/v1/users/me', {
      method: 'PATCH',
      headers: { Authorization: `Bearer ${accessToken}` },
      body: JSON.stringify({ name }),
    })
  },
  updateAvatar(accessToken: string, file: File) {
    const body = new FormData()
    body.append('avatar', file)
    return apiRequest<User>('/api/v1/users/me/avatar', {
      method: 'POST',
      headers: { Authorization: `Bearer ${accessToken}` },
      body,
    })
  },
}
