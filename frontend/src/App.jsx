import { useEffect, useState } from 'react'
import { Route, Routes, Link, Navigate, useLocation } from 'react-router-dom'
import './App.css'
import Home from './pages/Home.jsx'
import Register from './pages/Register.jsx'
import Login from './pages/Login.jsx'
import Profile from './pages/Profile.jsx'
import PublicProfile from './pages/PublicProfile.jsx'
import Dashboard from './pages/Dashboard.jsx'
import New from './pages/New.jsx'
import NotFound from './NotFound.jsx'
import { IconPulse, IconSun, IconMoon } from './icons.jsx'

function CatchAll() {
  const { pathname } = useLocation()
  const match = pathname.match(/^\/@([^/]+)\/?$/)
  if (match) {
    return <PublicProfile username={decodeURIComponent(match[1])} key={match[1]} />
  }
  return <NotFound />
}

function GuestOnly({ children }) {
  if (localStorage.getItem('access_token')) {
    return <Navigate to="/profile" replace />
  }
  return children
}

function AuthOnly({ children }) {
  if (!localStorage.getItem('access_token')) {
    return <Navigate to="/login" replace />
  }
  return children
}

function ThemeToggle({ theme, onToggle }) {
  const dark = theme === 'dark'
  return (
    <button
      type="button"
      className="theme-toggle"
      onClick={onToggle}
      aria-label={dark ? 'Switch to light theme' : 'Switch to dark theme'}
    >
      {dark ? <IconSun /> : <IconMoon />}
    </button>
  )
}

function Header({ theme, onToggleTheme }) {
  const [me, setMe] = useState(null)
  const location = useLocation()
  const token = localStorage.getItem('access_token')

  useEffect(() => {
    let cancelled = false

    fetch('/api/v1/auth/me', {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
      .then(async (res) => {
        if (cancelled) return
        if (res.ok) {
          setMe(await res.json())
        } else {
          setMe(null)
        }
      })
      .catch(() => {
        if (!cancelled) setMe(null)
      })

    return () => {
      cancelled = true
    }
  }, [location, token])

  return (
    <header className="site-header">
      <div className="container header-inner">
        <Link className="brand" to="/" aria-label="DevStream home">
          <IconPulse />
          <span>devstream</span>
        </Link>
        <nav className="site-nav" aria-label="Primary">
          <a href="/#features">Features</a>
        </nav>
        <div className="header-actions">
          <ThemeToggle theme={theme} onToggle={onToggleTheme} />
          {token && me?.avatar_url ? (
            <Link className="header-avatar" to="/profile" aria-label="Your profile">
              <img src={me.avatar_url} alt="" />
            </Link>
          ) : token ? (
            <Link className="btn btn-primary" to="/profile">Profile</Link>
          ) : (
            <>
              <Link className="btn btn-ghost" to="/login">Sign in</Link>
              <Link className="btn btn-primary" to="/register">Get started</Link>
            </>
          )}
        </div>
      </div>
    </header>
  )
}

function Footer() {
  return (
    <footer className="site-footer">
      <div className="container footer-inner">
        <span>&copy; 2026 DevStream &middot; MIT Licensed</span>
        <nav className="footer-links" aria-label="Footer">
          <a href="https://github.com/coderian/DevStream">Source</a>
          <a href="https://github.com/coderian/DevStream/blob/main/LICENSE">License</a>
        </nav>
      </div>
    </footer>
  )
}

function getInitialTheme() {
  return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light'
}

export default function App() {
  const [theme, setTheme] = useState(getInitialTheme)

  const toggleTheme = () => {
    const next = theme === 'dark' ? 'light' : 'dark'
    setTheme(next)
    document.documentElement.dataset.theme = next
    localStorage.setItem('theme', next)
  }

  return (
    <>
      <a className="skip-link" href="#main">Skip to content</a>
      <Header theme={theme} onToggleTheme={toggleTheme} />
      <main id="main">
        <Routes>
          <Route path="/" element={<GuestOnly><Home /></GuestOnly>} />
          <Route path="/register" element={<GuestOnly><Register /></GuestOnly>} />
          <Route path="/login" element={<GuestOnly><Login /></GuestOnly>} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/dashboard" element={<AuthOnly><Dashboard /></AuthOnly>} />
          <Route path="/new" element={<AuthOnly><New /></AuthOnly>} />
          <Route path="*" element={<CatchAll />} />
        </Routes>
      </main>
      <Footer />
    </>
  )
}
