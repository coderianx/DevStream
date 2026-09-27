import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import './auth.css'
import { IconCamera, IconCalendar, IconMail } from '../icons.jsx'
import PostList from '../PostList.jsx'

const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp']
const MAX_AVATAR_SIZE = 2 * 1024 * 1024
const MAX_BANNER_SIZE = 4 * 1024 * 1024

export default function Profile() {
  const [user, setUser] = useState(null)
  const [status, setStatus] = useState('loading')
  const [posts, setPosts] = useState([])
  const [postsError, setPostsError] = useState(false)
  const [uploading, setUploading] = useState('')
  const [uploadError, setUploadError] = useState('')
  const avatarInputRef = useRef(null)
  const bannerInputRef = useRef(null)
  const navigate = useNavigate()

  useEffect(() => {
    document.title = 'Profile — DevStream'
  }, [])

  const clearTokens = () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  const fetchProfile = () => {
    const token = localStorage.getItem('access_token')
    if (!token) {
      navigate('/login', { replace: true })
      return
    }

    fetch('/api/v1/auth/me', {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(async (res) => {
        if (res.status === 401) {
          clearTokens()
          navigate('/login', { replace: true })
          return
        }
        if (!res.ok) throw new Error()
        const me = await res.json()
        setUser(me)
        setStatus('ready')

        const postsRes = await fetch(`/api/v1/users/${me.username}/posts`)
        if (postsRes.ok) {
          const data = await postsRes.json()
          setPosts(data.posts)
          setPostsError(false)
        } else {
          setPostsError(true)
        }
      })
      .catch(() => setStatus('error'))
  }

  useEffect(fetchProfile, [navigate])

  const retry = () => {
    setStatus('loading')
    fetchProfile()
  }

  const onFileChange = (kind) => async (e) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    setUploadError('')

    if (!ALLOWED_TYPES.includes(file.type)) {
      setUploadError('Only JPEG, PNG and WebP images are allowed.')
      return
    }
    const maxSize = kind === 'banner' ? MAX_BANNER_SIZE : MAX_AVATAR_SIZE
    if (file.size > maxSize) {
      setUploadError(kind === 'banner'
        ? 'Banner must be smaller than 4 MB.'
        : 'Image must be smaller than 2 MB.')
      return
    }

    setUploading(kind)
    try {
      const token = localStorage.getItem('access_token')
      const formData = new FormData()
      formData.append(kind, file)
      const res = await fetch(`/api/v1/users/${kind}`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
        body: formData,
      })

      if (res.status === 401) {
        clearTokens()
        navigate('/login', { replace: true })
        return
      }
      const data = await res.json().catch(() => null)
      if (!res.ok) {
        setUploadError(data?.error ?? 'Upload failed. Please try again.')
        return
      }
      const urlKey = `${kind}_url`
      setUser((u) => (u ? { ...u, [urlKey]: data[urlKey] } : u))
    } catch {
      setUploadError('Could not reach the server. Please try again.')
    } finally {
      setUploading('')
    }
  }

  const signOut = async () => {
    const refreshToken = localStorage.getItem('refresh_token')
    if (refreshToken) {
      try {
        await fetch('/api/v1/auth/logout', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        })
      } catch {
        clearTokens()
        navigate('/login')
        return
      }
    }
    clearTokens()
    navigate('/login')
  }

  if (status === 'loading') {
    return (
      <section className="profile-page">
        <p className="profile-loading" role="status">Loading…</p>
      </section>
    )
  }

  if (status === 'error' || !user) {
    return (
      <section className="profile-page">
        <div className="profile-card">
          <p className="profile-error" role="alert">
            Could not load your profile.
          </p>
          <button className="btn btn-ghost profile-signout" type="button" onClick={retry}>
            Try again
          </button>
        </div>
      </section>
    )
  }

  const initial = user.username.charAt(0).toUpperCase()
  const memberSince = new Date(user.created_at).toLocaleDateString('en-US', {
    month: 'long',
    year: 'numeric',
  })

  return (
    <section className="profile-page">
      <article className="profile-hero">
        <button
          className="profile-banner profile-banner-btn"
          type="button"
          onClick={() => bannerInputRef.current?.click()}
          disabled={uploading === 'banner'}
          aria-label="Upload banner image"
        >
          {user.banner_url && <img src={user.banner_url} alt="" />}
          <span className="banner-overlay" aria-hidden="true">
            <IconCamera />
          </span>
        </button>
        <input
          ref={bannerInputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="visually-hidden"
          onChange={onFileChange('banner')}
        />

        <div className="profile-head">
          <button
            className="profile-avatar profile-avatar-btn profile-avatar-lg"
            type="button"
            onClick={() => avatarInputRef.current?.click()}
            disabled={uploading === 'avatar'}
            aria-label="Upload profile photo"
          >
            {user.avatar_url ? (
              <img src={user.avatar_url} alt={`${user.username}'s avatar`} />
            ) : (
              initial
            )}
            <span className="avatar-overlay" aria-hidden="true">
              <IconCamera />
            </span>
          </button>
          <input
            ref={avatarInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="visually-hidden"
            onChange={onFileChange('avatar')}
          />

          <div className="profile-head-info">
            <h1 className="profile-name">{user.username}</h1>
            <p className="profile-handle">@{user.username}</p>
          </div>

          <div className="profile-actions">
            <button className="btn btn-ghost" type="button" onClick={signOut}>
              Sign out
            </button>
          </div>
        </div>

        <ul className="profile-stats" aria-label="Profile statistics">
          <li><strong>{user.posts_count}</strong> Posts</li>
          <li><strong>{user.followers_count}</strong> Followers</li>
          <li><strong>{user.following_count}</strong> Following</li>
        </ul>

        <dl className="profile-meta">
          <div className="profile-meta-row">
            <IconMail />
            <dt className="visually-hidden">Email</dt>
            <dd>{user.email}</dd>
          </div>
          <div className="profile-meta-row">
            <IconCalendar />
            <dt className="visually-hidden">Member since</dt>
            <dd>Joined {memberSince}</dd>
          </div>
        </dl>

        {uploadError && (
          <p className="profile-upload-error" role="alert">
            {uploadError}
          </p>
        )}
      </article>

      <PostList posts={posts} error={postsError} />
    </section>
  )
}
