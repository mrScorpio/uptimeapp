import { ApiError } from './api'

export function authErrorMessage(error: unknown, action: 'register' | 'login' | 'session'): string {
  if (error instanceof ApiError) {
    if (action === 'register' && error.status === 409) return 'Этот email или логин уже занят.'
    if (action === 'login' && error.status === 401) return 'Неверный email, логин или пароль.'
    if (error.status === 400) return 'Проверьте введённые данные.'
    return error.message
  }
  return 'Произошла ошибка. Повторите попытку.'
}
