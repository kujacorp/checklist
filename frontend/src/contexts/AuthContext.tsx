import { createContext, useContext, useEffect, useState, ReactNode } from 'react'

interface User {
  username: string
  created_at: string
}

interface AuthState {
  user: User | null
  token: string | null
}

interface AuthContextType {
  user: User | null
  login: (username: string, password: string) => Promise<void>
  signup: (username: string, password: string) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
  authFetch: (url: string, options?: RequestInit) => Promise<Response>
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(() => {
    const token = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user')
    return {
      token,
      user: storedUser ? JSON.parse(storedUser) : null
    }
  })

  useEffect(() => {
    if (state.token) {
      fetch('/api/verify', {
        headers: {
          'Authorization': `Bearer ${state.token}`
        }
      }).catch(() => {
        // If token is invalid, logout
        logout()
      })
    }
  }, [state.token])

  const login = async (username: string, password: string) => {
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    })

    if (!response.ok) {
      throw new Error('Login failed')
    }

    const { token, user } = await response.json()
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    setState({ user, token })
  }

  const signup = async (username: string, password: string) => {
    const response = await fetch('/api/signup', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    })

    if (!response.ok) {
      throw new Error('Signup failed')
    }

    const { token, user } = await response.json()
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    setState({ user, token })
  }

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setState({ user: null, token: null })
  }

  const authFetch = async (url: string, options: RequestInit = {}) => {
    if (!state.token) throw new Error('No token available')

    const response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${state.token}`
      }
    })

    if (response.status === 401) {
      logout()
      throw new Error('Session expired')
    }

    return response
  }

  return (
    <AuthContext.Provider value={{
      user: state.user,
      isAuthenticated: !!state.token,
      login,
      signup,
      logout,
      authFetch
    }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
