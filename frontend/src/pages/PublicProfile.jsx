import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import './auth.css'
import { IconCalendar } from '../icons.jsx'
import PostList from '../PostList.jsx'
import NotFound from '../NotFound.jsx'

const clearTokens = () => {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

export default function PublicProfile({ username }) {
  const [user, setUser] = useState(null)
  const [status, setStatus] = useState('loading')
  const [posts, setPosts] = useState([])
  const [postsError, setPostsError] = useState(false)
  const [followers, setFollowers] = useState(0)
  const [followState, setFollowState] = useState('guest')
  const [followBusy, setFollowBusy] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    document.title = `@${username} — DevStream`
  }, [username])

  useEffect(() => {
    let cancelled = false

    fetch(`/api/v1/users/${encodeURIComponent(username)}`)
      .then(async (res) => {
        if (res.status === 404) {
          if (!cancelled) setStatus('notfound')
          return
        }
        if (!res.ok) throw new Error()
        const data = await res.json()
        if (cancelled) return
        setUser(data)
        setFollowers(data.followers_count)
        setStatus('ready')

        const token = localStorage.getItem('access_token')
        if (!token) {
          setFollowState('guest')
          return
        }
        const followRes = await fetch(
          `/api/v1/users/${encodeURIComponent(username)}/follow`,
          { headers: { Authorization: `Bearer ${token}` } },
        )
        if (followRes.status === 401) {
          clearTokens()
          navigate('/login', { replace: true })
          return
        }
        if (followRes.ok) {
          const followData = await followRes.json()
          setFollowState(followData.following ? 'following' : 'not')
        }

        const postsRes = await fetch(
          `/api/v1/users/${encodeURIComponent(username)}/posts`,
        )
        if (cancelled) return
        if (postsRes.ok) {
          const postsData = await postsRes.json()
          setPosts(postsData.posts)
          setPostsError(false)
        } else {
          setPostsError(true)
        }
      })
      .catch(() => {
        if (!cancelled) setStatus('error')
      })

    return () => {
      cancelled = true
    }
  }, [username, navigate])

  const toggleFollow = async () => {
    if (followBusy) return
    setFollowBusy(true)

    const wasFollowing = followState === 'following'
    try {
      const res = await fetch(
        `/api/v1/users/${encodeURIComponent(username)}/follow`,
        {
          method: wasFollowing ? 'DELETE' : 'POST',
          headers: { Authorization: `Bearer ${localStorage.getItem('access_token')}` },
        },
      )

      if (res.status === 401) {
        clearTokens()
        navigate('/login', { replace: true })
        return
      }
      if (res.ok) {
        setFollowState(wasFollowing ? 'not' : 'following')
        setFollowers((f) => (wasFollowing ? f - 1 : f + 1))
      }
    } catch {
      setFollowState(wasFollowing ? 'following' : 'not')
    } finally {
      setFollowBusy(false)
    }
  }

  if (status === 'notfound') {
    return <NotFound />
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
            Could not load this profile.
          </p>
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
        <div className="profile-banner">
          {user.banner_url && <img src={user.banner_url} alt="" />}
        </div>

        <div className="profile-head">
          <span className="profile-avatar profile-avatar-lg" aria-hidden="true">
            {user.avatar_url ? (
              <img src={user.avatar_url} alt="" />
            ) : (
              initial
            )}
          </span>

          <div className="profile-head-info">
            <h1 className="profile-name">{user.username}</h1>
            <p className="profile-handle">@{user.username}</p>
          </div>

          <div className="profile-actions">
            {followState === 'guest' ? (
              <Link className="btn btn-primary" to="/login">Follow</Link>
            ) : (
              <button
                className={followState === 'following' ? 'btn btn-ghost' : 'btn btn-primary'}
                type="button"
                onClick={toggleFollow}
                disabled={followBusy}
              >
                {followState === 'following' ? 'Following' : 'Follow'}
              </button>
            )}
          </div>
        </div>

        <ul className="profile-stats" aria-label="Profile statistics">
          <li><strong>{user.posts_count}</strong> Posts</li>
          <li><strong>{followers}</strong> Followers</li>
          <li><strong>{user.following_count}</strong> Following</li>
        </ul>

        <dl className="profile-meta">
          <div className="profile-meta-row">
            <IconCalendar />
            <dt className="visually-hidden">Member since</dt>
            <dd>Joined {memberSince}</dd>
          </div>
        </dl>
      </article>

      <PostList posts={posts} error={postsError} />
    </section>
  )
}
