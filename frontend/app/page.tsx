'use client'

import { FormEvent, useCallback, useEffect, useState } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { authErrorMessage } from '../lib/auth-errors'
import { useSession } from '../lib/session'
import { monitorsApi, type Monitor } from '../lib/api'

export default function HomePage() {
  const router = useRouter()
  const session = useSession()
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)
  const [monitorPending, setMonitorPending] = useState(false)
  const [monitorError, setMonitorError] = useState('')
  const [monitors, setMonitors] = useState<Monitor[]>([])
  const [monitorFormOpen, setMonitorFormOpen] = useState(false)

  const loadMonitors = useCallback(async () => {
    setMonitors(await monitorsApi.list())
  }, [])

  useEffect(() => {
    if (session.status === 'anonymous') router.replace('/login')
  }, [router, session.status])

  useEffect(() => {
    if (session.status !== 'authenticated') return
    let active = true
    async function load() {
      try {
        const result = await monitorsApi.list()
        if (active) setMonitors(result)
      } catch (cause) {
        if (active) setMonitorError(authErrorMessage(cause, 'session'))
      }
    }
    void load()
    return () => { active = false }
  }, [session.status])

  async function createMonitor(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (monitorPending) return
    setMonitorError('')
    setMonitorPending(true)
    const formElement = event.currentTarget
    const form = new FormData(formElement)
    try {
      await monitorsApi.create(String(form.get('url')), Number(form.get('interval')))
      formElement.reset()
      setMonitorFormOpen(false)
      await loadMonitors()
    } catch (cause) {
      setMonitorError(authErrorMessage(cause, 'session'))
    } finally {
      setMonitorPending(false)
    }
  }

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
      <main className="monitor-page">
        <div className="monitor-heading">
          <div><h1>Мониторинг сайтов</h1><p className="muted">Добавьте сайт, чтобы создать точку мониторинга.</p></div>
          <button type="button" className="add-monitor-link" onClick={() => { setMonitorError(''); setMonitorFormOpen(true) }}>Добавить сайт</button>
        </div>
        <section className="monitor-list" aria-label="Точки мониторинга">
          {monitors.length === 0 ? <p className="empty-monitors">Пока нет добавленных сайтов</p> : monitors.map((monitor) => (
            <article className="monitor-card" key={monitor.id}>
              <span className="monitor-indicator" aria-hidden="true" />
              <div className="monitor-info"><a href={monitor.url} target="_blank" rel="noreferrer">{monitor.url}</a><span>Опрос каждые {monitor.interval_seconds} сек.</span></div>
              <span className="monitor-status">Ожидает проверки</span>
            </article>
          ))}
        </section>
      </main>
      {monitorFormOpen && (
        <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget && !monitorPending) setMonitorFormOpen(false) }}>
          <section className="card monitor-form-card" role="dialog" aria-modal="true" aria-labelledby="monitor-form-title" onKeyDown={(event) => { if (event.key === 'Escape' && !monitorPending) setMonitorFormOpen(false) }}>
            <h2 id="monitor-form-title">Добавить сайт</h2>
            <p className="muted">Создайте точку мониторинга для сайта.</p>
            <form onSubmit={createMonitor}>
              <label htmlFor="monitor-url">URL сайта</label>
              <input id="monitor-url" name="url" type="url" placeholder="https://example.com" autoFocus required />
              <label htmlFor="monitor-interval">Частота опроса (секунды)</label>
              <input id="monitor-interval" name="interval" type="number" min="1" max="86400" defaultValue="60" required />
              {monitorError && <p className="error" role="alert">{monitorError}</p>}
              <div className="monitor-form-actions">
                <button type="button" className="secondary" disabled={monitorPending} onClick={() => setMonitorFormOpen(false)}>Отмена</button>
                <button type="submit" disabled={monitorPending}>{monitorPending ? 'Создаём…' : 'Создать'}</button>
              </div>
            </form>
          </section>
        </div>
      )}
    </div>
  )
}
