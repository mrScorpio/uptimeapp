'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { authErrorMessage } from '../lib/auth-errors'
import { useSession } from '../lib/session'

export default function HomePage() {
  const router = useRouter()
  const session = useSession()
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  useEffect(() => {
    if (session.status === 'anonymous') router.replace('/login')
  }, [router, session.status])

  async function logout() {
    if (pending) return
    setError('')
    setPending(true)
    try {
      await session.logout()
      router.replace('/login')
    } catch (cause) {
      setError(authErrorMessage(cause, 'session'))
    } finally {
      setPending(false)
    }
  }

  if (session.status !== 'authenticated' || !session.user) {
    return <main className="center"><p role="status">Загрузка сессии…</p></main>
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div className="brand">Uptime App</div>
        <details className="profile-menu">
          <summary aria-label={`Профиль пользователя ${session.user.name || session.user.login}`}>
            {session.user.avatar_url ? <img className="profile-avatar-image" src={`${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'}${session.user.avatar_url}`} alt="" /> : <span className="profile-avatar" aria-hidden="true">{(session.user.name || session.user.login).slice(0, 1).toUpperCase()}</span>}
            <span className="profile-name">{session.user.name || session.user.login}</span>
            <span className="profile-chevron" aria-hidden="true">⌄</span>
          </summary>
          <div className="profile-dropdown">
            <p className="profile-dropdown-title">Профиль</p>
            <dl>
              <div><dt>Имя</dt><dd>{session.user.name || session.user.login}</dd></div>
              <div><dt>Логин</dt><dd>{session.user.login}</dd></div>
              <div><dt>Email</dt><dd>{session.user.email}</dd></div>
            </dl>
            <Link className="profile-link" href="/profile">Настроить профиль</Link>
            <button type="button" className="secondary" disabled={pending} onClick={logout}>
              {pending ? 'Выходим…' : 'Выйти из аккаунта'}
            </button>
            {error && <p className="error profile-error" role="alert">{error}</p>}
          </div>
        </details>
      </header>
    </div>
  )
}
