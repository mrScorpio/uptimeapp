'use client'

import { useEffect, useState, type FormEvent } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { authErrorMessage } from '../../lib/auth-errors'
import { useSession } from '../../lib/session'

export default function ProfilePage() {
  const router = useRouter()
  const session = useSession()
  const [nameDraft, setNameDraft] = useState<{ userId: number; value: string } | null>(null)
  const [error, setError] = useState('')
  const [saved, setSaved] = useState(false)
  const [pending, setPending] = useState(false)
  const [avatarFile, setAvatarFile] = useState<File | null>(null)
  const [avatarPreview, setAvatarPreview] = useState('')
  const name = session.user && nameDraft?.userId === session.user.id
    ? nameDraft.value
    : session.user?.name || session.user?.login || ''

  useEffect(() => {
    if (session.status === 'anonymous') router.replace('/login')
  }, [router, session.status])

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending) return
    setError('')
    setSaved(false)
    setPending(true)
    try {
      await session.updateName(name.trim())
      if (avatarFile) await session.updateAvatar(avatarFile)
      if (session.user) setNameDraft({ userId: session.user.id, value: name.trim() })
      setSaved(true)
    } catch (cause) {
      setError(authErrorMessage(cause, 'session'))
    } finally {
      setPending(false)
    }
  }

  if (session.status !== 'authenticated' || !session.user) {
    return <main className="center"><p role="status">Загрузка сессии…</p></main>
  }
  const user = session.user

  return (
    <main className="center profile-page">
      <section className="card profile-card" aria-labelledby="profile-title">
        <Link className="back-link" href="/">← На главную</Link>
        <div className="brand">Uptime App</div>
        <h1 id="profile-title">Профиль пользователя</h1>
        <div className="avatar-editor">
          {avatarPreview || user.avatar_url ? <img className="avatar-preview" src={avatarPreview || `${process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'}${user.avatar_url}`} alt="Аватар профиля" /> : <span className="profile-avatar">{(name || user.login).slice(0, 1).toUpperCase()}</span>}
          <label htmlFor="profile-avatar">Загрузить аватар</label>
          <input id="profile-avatar" type="file" accept="image/png,image/jpeg,image/webp" disabled={pending} onChange={(event) => {
            const file = event.target.files?.[0] ?? null
            setAvatarFile(file)
            setAvatarPreview(file ? URL.createObjectURL(file) : '')
            setSaved(false)
          }} />
          <p className="hint">PNG, JPEG или WebP, до 5 МБ. Изменение вступит в силу после сохранения.</p>
        </div>
        <dl className="profile-details">
          <div><dt>Логин</dt><dd>{user.login}</dd></div>
          <div><dt>Email</dt><dd>{user.email}</dd></div>
        </dl>
        <form onSubmit={submit}>
          <label htmlFor="profile-name">Имя</label>
          <input
            id="profile-name"
            name="name"
            type="text"
            autoComplete="name"
            value={name}
            onChange={(event) => {
              setNameDraft({ userId: user.id, value: event.target.value })
              setSaved(false)
            }}
            required
            maxLength={100}
            disabled={pending}
          />
          {error && <p className="error" role="alert">{error}</p>}
          {saved && <p className="success" role="status">Имя сохранено.</p>}
          <button type="submit" disabled={pending || !name.trim()}>
            {pending ? 'Сохраняем…' : 'Сохранить изменения'}
          </button>
        </form>
      </section>
    </main>
  )
}
