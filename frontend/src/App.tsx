import { useEffect, useState } from "react"
import { useAuth } from "./contexts/AuthContext"
import { Login } from "./components/Login"
import { SignUp } from "./components/SignUp"

function App() {
  const { isAuthenticated, user, logout, authFetch } = useAuth()
  const [count, setCount] = useState<number>(0)
  const [showSignUp, setShowSignUp] = useState(false)

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
  }, [isAuthenticated, authFetch, logout])

  if (!isAuthenticated) {
    return (
      <div>
        {showSignUp ? (
          <>
            <SignUp />
            <p>
              Already have an account?{" "}
              <button onClick={() => setShowSignUp(false)}>Log in</button>
            </p>
          </>
        ) : (
          <>
            <Login />
            <p>
              Don't have an account?{" "}
              <button onClick={() => setShowSignUp(true)}>Sign up</button>
            </p>
          </>
        )}
      </div>
    )
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
