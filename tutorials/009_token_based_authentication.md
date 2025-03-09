# Adding token-based authentication ([#9](https://github.com/kujacorp/checklist/pull/9))
We'll add token-based authentication to `checklist` web application in this tutorial.
This web application consists of a Go backend and TypeScript frontend communicating over HTTP.

## Setup
This tutorial starts at SHA [89d37af](https://github.com/kujacorp/checklist/tree/89d37afffdd3232023c645be2a02357fa981b7c5).
```sh
git checkout 89d37af
```

At this SHA, the app has rudimentary authentication where it can check the user's password.
If the user exists and the password matches, it simply returns a successful status code indicating the log in succeeded.
All API endpoints are also unauthenticated.

By the end of this tutorial, this will be made secure by using token-based authentication using JSON Web Tokens (JWTs).
JWTs give the client a token which they can include in requests to prove that they're authenticated.
We'll also create a middleware for the server to require authentication for API endpoints.

## Background

### JSON Web Tokens

JWTs are a standard token standard used in a lot of modern web-apps.
A lot of older websites used sessions and cookies, but those don't scale as well.

A JWT has three parts:

1. **Header**: Identifies which algorithm is used to generate the signature
2. **Payload**: Contains the "claims" (data) being transferred
3. **Signature**: Ensures the token hasn't been altered

See [jwt.io](https://jwt.io/) for a demo of how they work.

Unlike session-based authentication, JWTs are stateless—the server doesn't need to store session information, making them ideal when your app is hosted on a cluster.

In our implementation, we use JWTs as the secure token used to authenticate requests to the server.

### The Authorization Header and Bearer Tokens

When using JWTs for authentication, the token is typically sent in the HTTP Authorization header using the Bearer scheme.

The format is:

```
Authorization: Bearer <token>
```

For example:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImpvaG4iLCJleHAiOjE2NTIzNzM5MDN9.JkLX3Qp3q8vlwCh9c4UOw-f5NWY9t0V1OrNFd7O8YW4
```

This is done because it's:
- Standard: The `Bearer` format is widely recognized
- Secure: Tokens are only sent over HTTPS

In our implementation:
1. The backend will verify this header in the `authMiddleware`
2. The frontend will automatically include this header in all authenticated requests via the `authFetch` utility

### Go's Struct Embedding

Go lets you take fields from an existing struct and embed them in a new struct.
It calls this feature [Embedding](https://go.dev/doc/effective_go#embedding).
Consider this:

```go
type Address struct {
    Street  string
    City    string
    Country string
}

type Person struct {
    Name    string
    Age     int
    Address // This is an embedded struct
}
```

By embedding `Address`, a `Person` can directly access address fields:

```go
p := Person{Name: "Alice", Age: 30}
p.Street = "123 Main St"  // Direct access to Address.Street
p.City = "New York"       // Direct access to Address.City

// Instead of:
// p.Address.Street = "123 Main St"
```

In our implementation, embedding is used when implementing JWT claims in Go, allowing us to extend standard claims with custom fields while maintaining a clean interface.

## Part 1: Setting Up the Go Backend

### Step 1: Install the JWT Package

First, we need to add the JWT package to our Go application:

```bash
go get github.com/golang-jwt/jwt/v5
```

### Step 2: Define JWT Structures

Let's define the essential JWT structures in our `main.go`:

```go
// Secret key for signing JWT tokens
var jwtKey = []byte("your-secret-key")

// Custom claims struct for JWT, note this uses struct embedding so all
// members of jwt.RegisteredClaims are members of Claims
type Claims struct {
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// Update response to include token
type LoginResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}
```

### Step 3: Implement Token Generation

Add a function to generate tokens for authenticated users:

```go
func generateToken(username string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}
```

### Step 4: Create Authentication Middleware

Implement middleware to protect routes:

```go
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }

        // Remove "Bearer " prefix
        tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
        claims := &Claims{}

        // Parse and validate token
        token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add username to request context for later use
        ctx := context.WithValue(r.Context(), "username", claims.Username)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

### Step 5: Update Login Handler

Modify the login handler to generate and return a token:

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    var storedHash string
    var user User
    err := db.QueryRow(
        "SELECT password_hash, username, created_at FROM users WHERE username = $1",
        req.Username,
    ).Scan(&storedHash, &user.Username, &user.CreatedAt)

    // Validate the password
    if err != nil || bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)) != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate token for the authenticated user
    token, err := generateToken(user.Username)
    if err != nil {
        http.Error(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    // Return token and user information
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(LoginResponse{
        Token: token,
        User:  user,
    })
}
```

### Step 6: Add Token Verification Endpoint

Create an endpoint to verify token validity:

```go
func verifyHandler(w http.ResponseWriter, r *http.Request) {
    // This will be wrapped by authMiddleware which verifies the token
    // If we get here, the token is valid
    w.WriteHeader(http.StatusOK)
}
```

### Step 7: Protect Routes

Apply the middleware to routes that require authentication:

```go
func main() {
    http.HandleFunc("/", authMiddleware(viewCountHandler))
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/verify", authMiddleware(verifyHandler))
    // Other routes...
}
```

## Part 2: Implementing Authentication in React (TypeScript)

### Step 1: Update the Auth Context

First, let's enhance our authentication context to handle tokens:

```tsx
interface AuthState {
  user: User | null
  token: string | null
}

export function AuthProvider({ children }: { children: ReactNode }) {
  // Initialize state from localStorage
  const [state, setState] = useState<AuthState>(() => {
    const token = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user')
    return {
      token,
      user: storedUser ? JSON.parse(storedUser) : null
    }
  })

  // Verify token on initial load
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
  }, [])

  // Rest of the component...
}
```

### Step 2: Update Login Function

Modify the login function to store the token:

```tsx
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
```

### Step 3: Implement Logout Function

Update the logout function to clear the token:

```tsx
const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  setState({ user: null, token: null })
}
```

### Step 4: Create Authenticated Fetch Utility

Add a utility function to make authenticated requests:

```tsx
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
```

### Step 5: Update Auth Context Provider

Make these new functionalities available through the context:

```tsx
return (
  <AuthContext.Provider value={{
    user: state.user,
    isAuthenticated: !!state.token,
    login,
    logout,
    authFetch
  }}>
    {children}
  </AuthContext.Provider>
)
```

### Step 6: Use Authenticated Fetching in Components

Update components to use `authFetch` for API calls:

```tsx
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
        .then(data => setCount(data.count))
        .catch(err => {
          console.error("Failed to fetch count:", err)
          if (err.message === 'Session expired') {
            logout()
          }
        })
    }
  }, [isAuthenticated])

  // Rest of component...
}
```
