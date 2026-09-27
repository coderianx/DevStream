import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import './NotFound.css'

export default function NotFound() {
  useEffect(() => {
    document.title = 'Page not found — DevStream'
  }, [])

  return (
    <section className="notfound" aria-labelledby="notfound-title">
      <span className="notfound-code" aria-hidden="true">404</span>
      <h1 id="notfound-title">Page not found</h1>
      <p>
        The page you are looking for does not exist or may have been moved.
      </p>
      <div className="notfound-actions">
        <Link className="btn btn-primary" to="/">Back to home</Link>
        <a className="btn btn-ghost" href="https://github.com/coderian/DevStream">View source</a>
      </div>
    </section>
  )
}
