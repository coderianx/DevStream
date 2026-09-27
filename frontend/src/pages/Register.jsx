import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import './auth.css'
import Field from './Field.jsx'

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export default function Register() {
  const [form, setForm] = useState({ username: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    document.title = 'Create account — DevStream'
  }, [])

  const setField = (key) => (e) => {
    setForm((f) => ({ ...f, [key]: e.target.value }))
    setError('')
  }

  const validate = () => {
    if (form.username.trim().length < 3) {
      return 'Username must be at least 3 characters.'
    }
    if (!EMAIL_RE.test(form.email)) {
      return 'Enter a valid email address.'
    }
    if (form.password.length < 8) {
      return 'Password must be at least 8 characters.'
    }
    return ''
  }

  const onSubmit = async (e) => {
    e.preventDefault()
    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }

    setLoading(true)
    setError('')

    try {
      const res = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: form.username.trim(),
          email: form.email.trim().toLowerCase(),
          password: form.password,
        }),
      })

      if (res.status === 201) {
        setSuccess(true)
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
        <h1>Create your account</h1>
        <p className="auth-sub">
          Join DevStream, follow builders and make your first post.
        </p>

        {success ? (
          <div className="auth-success" role="status">
            Account created. You can sign in now.
          </div>
        ) : (
          <>
            <form onSubmit={onSubmit} noValidate>
              <Field
                id="username"
                label="Username"
                type="text"
                value={form.username}
                onChange={setField('username')}
                autoComplete="username"
                hint="Visible on your posts."
              />
              <Field
                id="email"
                label="Email"
                type="email"
                value={form.email}
                onChange={setField('email')}
                autoComplete="email"
              />
              <Field
                id="password"
                label="Password"
                type="password"
                value={form.password}
                onChange={setField('password')}
                autoComplete="new-password"
                hint="At least 8 characters."
              />
              <button className="btn btn-primary auth-submit" type="submit" disabled={loading}>
                {loading ? 'Creating…' : 'Create account'}
              </button>
              {error && (
                <p className="auth-error" role="alert">
                  {error}
                </p>
              )}
            </form>
            <p className="auth-alt">
              Already have an account? <Link to="/login">Sign in</Link>
            </p>
          </>
        )}
      </div>
    </section>
  )
}
