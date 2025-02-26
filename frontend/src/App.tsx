import { useEffect, useState } from "react"

function App() {
  const [count, setCount] = useState<number>(-1)

  useEffect(() => {
    fetch("/api")
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP error! Status: ${res.status}`)
        return res.json()
      })
      .then((data) => setCount(data.count))
      .catch((err) => console.error("Failed to fetch count:", err))
  }, [])

  return (
    <div>
      <h1>Hello World!</h1>
      <p>I have been seen {count !== -1 ? count : "loading..."} times.</p>
    </div>
  )
}

export default App
