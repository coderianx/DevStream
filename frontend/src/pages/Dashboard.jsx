import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import './Dashboard.css'
import { IconSearch } from '../icons.jsx'

const clearTokens = () => {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

export default function Dashboard() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [status, setStatus] = useState('idle')
  const navigate = useNavigate()

  useEffect(() => {
    document.title = 'Dashboard — DevStream'
  }, [])

  useEffect(() => {
    const q = query.trim()
    let cancelled = false

    const handle = setTimeout(async () => {
      if (!q) {
        setResults([])
        setStatus('idle')
        return
      }

      setStatus('loading')
      try {
        const res = await fetch(
          `/api/v1/users/search?q=${encodeURIComponent(q)}`,
          { headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` } },
        )
        if (cancelled) return
        if (res.status === 401) {
          clearTokens()
          navigate('/login', { replace: true })
          return
        }
        if (!res.ok) throw new Error()
        const data = await res.json()
        setResults(data.users)
        setStatus('ready')
      } catch {
        if (!cancelled) setStatus('error')
      }
    }, 300)

    return () => {
      cancelled = true
      clearTimeout(handle)
    }
  }, [query, navigate])

  return (
    <section className="dashboard-page">
      <div className="container dashboard-inner">
        <div className="dashboard-header">
          <h1>Dashboard</h1>
          <Link className="btn btn-primary" to="/new">+ Create New Post</Link>
        </div>
        <p className="dashboard-sub">Find builders to follow.</p>

        <div className="search-box">
          <IconSearch />
          <input
            type="search"
            placeholder="Search users"
            aria-label="Search users"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </div>

        <ul className="user-list" aria-live="polite">
          {status === 'loading' && (
            <li className="user-list-status">Searching…</li>
          )}
          {status === 'idle' && (
            <li className="user-list-status">Start typing to find builders.</li>
          )}
          {status === 'error' && (
            <li className="user-list-status" role="alert">
              Could not search. Try again.
            </li>
          )}
          {status === 'ready' && results.length === 0 && (
            <li className="user-list-status">No users found.</li>
          )}
          {results.map((u) => (
            <li key={u.username}>
              <Link className="user-item" to={`/@${u.username}`}>
                <span className="user-item-avatar" aria-hidden="true">
                  {u.avatar_url ? (
                    <img src={u.avatar_url} alt="" />
                  ) : (
                    u.username.charAt(0).toUpperCase()
                  )}
                </span>
                <span className="user-item-name">{u.username}</span>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </section>
  )
}
