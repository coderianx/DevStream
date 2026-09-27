import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import Markdown from 'react-markdown'
import './auth.css'
import './New.css'
import Field from './Field.jsx'

const clearTokens = () => {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

export default function New() {
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [mode, setMode] = useState('write')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    document.title = 'New post — DevStream'
  }, [])

  const onSubmit = async (e) => {
    e.preventDefault()

    const t = title.trim()
    if (!t) {
      setError('Title is required.')
      return
    }
    if (t.length > 255) {
      setError('Title is too long.')
      return
    }
    const body = content.trim()
    if (!body) {
      setError('Content is required.')
      return
    }
    if (body.length > 10000) {
      setError('Content is too long.')
      return
    }

    setLoading(true)
    setError('')
    try {
      const res = await fetch('/api/v1/posts', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('access_token')}`,
        },
        body: JSON.stringify({ title: t, content: body }),
      })

      if (res.status === 401) {
        clearTokens()
        navigate('/login', { replace: true })
        return
      }
      if (res.status === 201) {
        navigate('/dashboard')
        return
      }
      const data = await res.json().catch(() => null)
      setError(data?.error ?? 'Something went wrong. Please try again.')
    } catch {
      setError('Could not reach the server. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <section className="new-post-page">
      <form className="new-post-card" onSubmit={onSubmit} noValidate>
        <h1>New post</h1>

        <Field
          id="title"
          label="Title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          hint="Max 255 characters."
        />

        <div className="new-post-content">
          <div className="mode-tabs" role="tablist" aria-label="Editor mode">
            <button
              type="button"
              role="tab"
              aria-selected={mode === 'write'}
              className="mode-tab"
              onClick={() => setMode('write')}
            >
              Write
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={mode === 'preview'}
              className="mode-tab"
              onClick={() => setMode('preview')}
            >
              Preview
            </button>
          </div>

          {mode === 'write' ? (
            <textarea
              aria-label="Post content"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={10}
              placeholder="Write in markdown…"
            />
          ) : (
            <div className="markdown-preview markdown-body">
              {content.trim() ? (
                <Markdown>{content}</Markdown>
              ) : (
                <p className="preview-empty">Nothing to preview yet.</p>
              )}
            </div>
          )}
        </div>
        <p className="field-hint new-post-hint">
          Markdown supported: **bold**, `code`, lists, links.
        </p>

        <button className="btn btn-primary new-post-submit" type="submit" disabled={loading}>
          {loading ? 'Publishing…' : 'Publish'}
        </button>
        {error && (
          <p className="auth-error" role="alert">
            {error}
          </p>
        )}
      </form>
    </section>
  )
}
