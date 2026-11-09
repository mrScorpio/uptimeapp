export type User = {
  id: number
  email: string
  login: string
  name: string
  avatar_url: string
  created_at: string
}

export type AuthResponse = {
  access_token: string
  token_type: 'Bearer'
  user: User
}

export type RefreshResponse = {
  access_token: string
  token_type: 'Bearer'
}
