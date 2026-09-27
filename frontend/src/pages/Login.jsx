import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import './auth.css'
import Field from './Field.jsx'

export default function Login() {
  const [form, setForm] = useState({ username: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    document.title = 'Sign in — DevStream'
  }, [])

  const setField = (key) => (e) => {
    setForm((f) => ({ ...f, [key]: e.target.value }))
    setError('')
  }

  const onSubmit = async (e) => {
    e.preventDefault()
    if (!form.username.trim() || !form.password) {
      setError('Enter your username and password.')
      return
    }

    setLoading(true)
    setError('')

    try {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: form.username.trim(),
          password: form.password,
        }),
      })

      if (res.ok) {
        const data = await res.json()
        localStorage.setItem('access_token', data.token)
        localStorage.setItem('refresh_token', data.refresh_token)
        navigate('/profile')
      } else {
        const data = await res.json().catch(() => null)
        setError(data?.error ?? 'Something went wrong. Please try again.')
      }
    } catch {
      setError('Could not reach the server. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <section className="auth-page">
      <div className="auth-card">
        <h1>Welcome back</h1>
        <p className="auth-sub">Sign in to post, reply and follow builders.</p>

        <form onSubmit={onSubmit} noValidate>
          <Field
            id="username"
            label="Username"
            type="text"
            value={form.username}
            onChange={setField('username')}
            autoComplete="username"
          />
          <Field
            id="password"
            label="Password"
            type="password"
            value={form.password}
            onChange={setField('password')}
            autoComplete="current-password"
          />
          <button className="btn btn-primary auth-submit" type="submit" disabled={loading}>
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
          {error && (
            <p className="auth-error" role="alert">
              {error}
            </p>
          )}
        </form>
        <p className="auth-alt">
          Don't have an account? <Link to="/register">Create one</Link>
        </p>
      </div>
    </section>
  )
}
