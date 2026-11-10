import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import type { AuthResponse, RefreshResponse, User } from '../types/auth'

export type Monitor = { id: number; url: string; interval_seconds: number; created_at: string }

const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

type RequestConfig = InternalAxiosRequestConfig & { skipAuth?: boolean }
type AuthTokenProvider = () => Promise<string>
let authTokenProvider: AuthTokenProvider | null = null

export class ApiError extends Error {
  constructor(message: string, readonly status: number) {
    super(message)
    this.name = 'ApiError'
  }
}

export const apiClient = axios.create({
  baseURL: apiBaseUrl,
  withCredentials: true,
  headers: { 'Content-Type': 'application/json' },
})

export function setAuthTokenProvider(provider: AuthTokenProvider) {
  authTokenProvider = provider
}

apiClient.interceptors.request.use(async (config) => {
  const requestConfig = config as RequestConfig
  if (requestConfig.skipAuth) return config
  if (!authTokenProvider) throw new ApiError('Сессия не инициализирована.', 401)
  const token = await authTokenProvider()
  config.headers.set('Authorization', `Bearer ${token}`)
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (cause: unknown) => {
    if (cause instanceof AxiosError) {
      const message = cause.response?.data?.error ?? (cause.request
        ? 'Не удалось связаться с сервером. Повторите попытку позже.'
        : 'Не удалось выполнить запрос.')
      return Promise.reject(new ApiError(message, cause.response?.status ?? 0))
    }
    return Promise.reject(cause)
  },
)

export const authApi = {
  async register(email: string, login: string, password: string) {
    const { data } = await apiClient.post<AuthResponse>('/api/v1/auth/register', { email, login, password }, { skipAuth: true } as RequestConfig)
    return data
  },
  async login(identifier: string, password: string) {
    const { data } = await apiClient.post<AuthResponse>('/api/v1/auth/login', { identifier, password }, { skipAuth: true } as RequestConfig)
    return data
  },
  async refresh() {
    const { data } = await apiClient.post<RefreshResponse>('/api/v1/auth/refresh', undefined, { skipAuth: true } as RequestConfig)
    return data
  },
  async logout() {
    await apiClient.post<void>('/api/v1/auth/logout', undefined, { skipAuth: true } as RequestConfig)
  },
}

export const usersApi = {
  async me() {
    const { data } = await apiClient.get<User>('/api/v1/users/me')
    return data
  },
  async updateName(name: string) {
    const { data } = await apiClient.patch<User>('/api/v1/users/me', { name })
    return data
  },
  async updateAvatar(file: File) {
    const body = new FormData()
    body.append('avatar', file)
    const { data } = await apiClient.post<User>('/api/v1/users/me/avatar', body, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return data
  },
}

export const monitorsApi = {
  async list() {
    const { data } = await apiClient.get<Monitor[]>('/api/v1/monitors')
    return data
  },
  async create(url: string, intervalSeconds: number) {
    const { data } = await apiClient.post<Monitor>('/api/v1/monitors', { url, interval_seconds: intervalSeconds })
    return data
  },
}
