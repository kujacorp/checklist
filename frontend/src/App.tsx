import { useEffect, useState } from "react"
import { useAuth } from "./contexts/AuthContext"
import { Login } from "./components/Login"

function App() {
  const { isAuthenticated, user, logout, authFetch } = useAuth()
  const [count, setCount] = useState<number>(0)

  useEffect(() => {
    if (isAuthenticated) {
      authFetch("/api")
        .then((res) => {
          if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`)
          return res.json()
        })
        .then((data) => setCount(data.count))
        .then(data => setCount(data.count))
                .catch(err => {
                  console.error("Failed to fetch count:", err)
                  if (err.message === 'Session expired') {
                    logout()
                  }
                })
            }
  }, [isAuthenticated])

  if (!isAuthenticated) {
    return <Login />
  }

  return (
    <div>
      <h1>Hello {user?.username}!</h1>
      <p>I have been seen {count !== 0 ? count : "loading..."} times.</p>
      <button onClick={logout}>Logout</button>
    </div>
  )
}

export default App
