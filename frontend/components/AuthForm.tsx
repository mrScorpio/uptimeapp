'use client'

import { useEffect, useState, type FormEvent } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { authErrorMessage } from '../lib/auth-errors'
import { useSession } from '../lib/session'

type Mode = 'login' | 'register'

export function AuthForm({ mode }: { mode: Mode }) {
  const router = useRouter()
  const session = useSession()
  const [email, setEmail] = useState('')
  const [login, setLogin] = useState('')
  const [identifier, setIdentifier] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  useEffect(() => {
    if (session.status === 'authenticated') router.replace('/')
  }, [router, session.status])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending) return
    setError('')
    setPending(true)
    try {
      if (mode === 'register') {
        await session.register(email.trim(), login.trim(), password)
      } else {
        await session.login(identifier.trim(), password)
      }
      router.replace('/')
    } catch (cause) {
      setError(authErrorMessage(cause, mode))
    } finally {
      setPending(false)
    }
  }

  if (session.status === 'loading' || session.status === 'authenticated') {
    return <main className="center"><p role="status">Загрузка сессии…</p></main>
  }

  const registering = mode === 'register'
  return (
    <main className="center">
      <section className="card auth-card" aria-labelledby="auth-title">
        <div className="brand">Uptime App</div>
        <h1 id="auth-title">{registering ? 'Создать аккаунт' : 'Войти в аккаунт'}</h1>
        <p className="muted">
          {registering ? 'Зарегистрируйтесь, чтобы начать работу.' : 'Введите email или логин и пароль.'}
        </p>
        <form onSubmit={submit}>
          {registering ? (
            <>
              <label htmlFor="email">Email</label>
              <input id="email" name="email" type="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} required maxLength={254} disabled={pending} />
              <label htmlFor="login">Логин</label>
              <input id="login" name="login" type="text" autoComplete="username" value={login} onChange={(event) => setLogin(event.target.value)} required minLength={3} maxLength={32} pattern="[A-Za-z0-9_-]+" title="3–32 латинские буквы, цифры, _ или -" disabled={pending} />
            </>
          ) : (
            <>
              <label htmlFor="identifier">Email или логин</label>
              <input id="identifier" name="identifier" type="text" autoComplete="username" value={identifier} onChange={(event) => setIdentifier(event.target.value)} required disabled={pending} />
            </>
          )}
          <label htmlFor="password">Пароль</label>
          <input id="password" name="password" type="password" autoComplete={registering ? 'new-password' : 'current-password'} value={password} onChange={(event) => setPassword(event.target.value)} required minLength={registering ? 8 : undefined} maxLength={registering ? 72 : undefined} disabled={pending} />
          {registering && <p className="hint">Минимум 8 символов, максимум 72 байта.</p>}
          {error && <p className="error" role="alert">{error}</p>}
          <button type="submit" disabled={pending}>{pending ? 'Подождите…' : registering ? 'Создать аккаунт' : 'Войти'}</button>
        </form>
        <p className="switch-link">
          {registering ? 'Уже есть аккаунт? ' : 'Нет аккаунта? '}
          <Link href={registering ? '/login' : '/register'}>{registering ? 'Войти' : 'Зарегистрироваться'}</Link>
        </p>
      </section>
    </main>
  )
}
