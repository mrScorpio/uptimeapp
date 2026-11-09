import type { Metadata } from 'next'
import { SessionProvider } from '../lib/session'
import './globals.css'

export const metadata: Metadata = {
  title: 'Uptime App',
  description: 'Мониторинг доступности сервисов',
}

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ru">
      <body>
        <SessionProvider>{children}</SessionProvider>
      </body>
    </html>
  )
}
