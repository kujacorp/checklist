import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'

export function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await login(username, password)
    } catch (err) {
      console.log(err)
      setError('Login failed')
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <div>
        <label>
          Username:
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
        </label>
      </div>
      <div>
        <label>
          Password:
          <input
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </label>
      </div>
      <button type="submit">Login</button>
    </form>
  )
}
