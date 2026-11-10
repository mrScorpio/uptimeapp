'use client'

import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react'
import { ApiError, authApi, setAuthTokenProvider, usersApi } from './api'
import type { User } from '../types/auth'

type SessionState = {
  status: 'loading' | 'authenticated' | 'anonymous'
  user: User | null
}

type SessionContextValue = SessionState & {
  register(email: string, login: string, password: string): Promise<void>
  login(identifier: string, password: string): Promise<void>
  logout(): Promise<void>
  getAccessToken(): Promise<string>
  updateName(name: string): Promise<void>
  updateAvatar(file: File): Promise<void>
}

const SessionContext = createContext<SessionContextValue | null>(null)

function tokenExpiresSoon(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/'))) as { exp?: number }
    return !payload.exp || payload.exp * 1000 <= Date.now() + 30_000
  } catch {
    return true
  }
}

export function SessionProvider({ children }: { children: React.ReactNode }) {
  const [session, setSession] = useState<SessionState>({ status: 'loading', user: null })
  const accessToken = useRef<string | null>(null)
  const refreshPromise = useRef<Promise<string> | null>(null)
  const generation = useRef(0)

  const clear = useCallback(() => {
    generation.current += 1
    accessToken.current = null
    refreshPromise.current = null
    setSession({ status: 'anonymous', user: null })
  }, [])

  const refresh = useCallback((): Promise<string> => {
    if (refreshPromise.current) return refreshPromise.current
    const startedAt = generation.current
    const pending = authApi.refresh().then(({ access_token }) => {
      if (generation.current !== startedAt) throw new Error('Session changed')
      accessToken.current = access_token
      return access_token
    }).finally(() => {
      if (refreshPromise.current === pending) refreshPromise.current = null
    })
    refreshPromise.current = pending
    return pending
  }, [])

  const getAccessToken = useCallback(async (): Promise<string> => {
    const token = accessToken.current
    if (token && !tokenExpiresSoon(token)) return token
    try {
      return await refresh()
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) clear()
      throw error
    }
  }, [clear, refresh])

  useEffect(() => {
    setAuthTokenProvider(getAccessToken)
  }, [getAccessToken])

  const loadUser = useCallback(async (): Promise<User> => {
    try {
      return await usersApi.me()
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        await refresh()
        return usersApi.me()
      }
      throw error
    }
  }, [refresh])

  useEffect(() => {
    let active = true
    const startedAt = generation.current
    refresh().then(loadUser).then((user) => {
      if (active && generation.current === startedAt) setSession({ status: 'authenticated', user })
    }).catch((error: unknown) => {
      if (!active || generation.current !== startedAt) return
      if (error instanceof ApiError && error.status === 401) {
        clear()
      } else {
        setSession({ status: 'anonymous', user: null })
      }
    })
    return () => { active = false }
  }, [clear, loadUser, refresh])

  const register = useCallback(async (email: string, login: string, password: string) => {
    const response = await authApi.register(email, login, password)
    generation.current += 1
    accessToken.current = response.access_token
    refreshPromise.current = null
    setSession({ status: 'authenticated', user: response.user })
  }, [])

  const login = useCallback(async (identifier: string, password: string) => {
    const response = await authApi.login(identifier, password)
    generation.current += 1
    accessToken.current = response.access_token
    refreshPromise.current = null
    setSession({ status: 'authenticated', user: response.user })
  }, [])

  const logout = useCallback(async () => {
    await authApi.logout()
    clear()
  }, [clear])

  const updateName = useCallback(async (name: string) => {
    const user = await usersApi.updateName(name)
    setSession({ status: 'authenticated', user })
  }, [])

  const updateAvatar = useCallback(async (file: File) => {
    const user = await usersApi.updateAvatar(file)
    setSession({ status: 'authenticated', user })
  }, [])

  return (
    <SessionContext.Provider value={{ ...session, register, login, logout, getAccessToken, updateName, updateAvatar }}>
      {children}
    </SessionContext.Provider>
  )
}

export function useSession() {
  const session = useContext(SessionContext)
  if (!session) throw new Error('useSession must be used inside SessionProvider')
  return session
}
