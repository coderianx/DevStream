import { Link } from 'react-router-dom'
import { IconCode, IconUsers, IconTerminal, IconReply, IconRepeat, IconHeart } from '../icons.jsx'

function Hero() {
  return (
    <section className="container hero">
      <span className="badge">Now in beta</span>
      <h1>
        Microblogging, built <em>for developers</em>.
      </h1>
      <p className="hero-sub">
        Post code snippets, follow builders and keep the conversation technical.
        DevStream strips social media down to what developers actually need.
      </p>
      <div className="hero-actions">
        <Link className="btn btn-primary" to="/register">Create account</Link>
        <a className="btn btn-ghost" href="#features">Learn more</a>
      </div>

      <figure className="post-card" aria-label="Example post on DevStream">
        <div className="post-header">
          <span className="avatar" aria-hidden="true">CD</span>
          <div className="post-author">
            <span className="post-name">coderian</span>
            <span className="post-handle">@coderian &middot; 2h</span>
          </div>
        </div>
        <blockquote className="post-body">
          Just wired per-IP rate limiting into the gateway &mdash;{' '}
          <code>30 req/min</code>, Redis-backed. Brute force, meet{' '}
          <code>429</code>. #golang
        </blockquote>
        <div className="post-meta">
          <span><IconReply /> 12</span>
          <span><IconRepeat /> 34</span>
          <span><IconHeart /> 128</span>
        </div>
      </figure>
    </section>
  )
}

function Features() {
  return (
    <section id="features" className="section">
      <div className="container">
        <div className="section-header">
          <span className="eyebrow">Features</span>
          <h2>Less feed, more signal</h2>
        </div>
        <div className="feature-grid">
          <article className="feature">
            <span className="feature-icon" aria-hidden="true"><IconCode /></span>
            <h3>Code-first posts</h3>
            <p>Syntax-friendly snippets and inline code blocks are first-class citizens, not an afterthought.</p>
          </article>
          <article className="feature">
            <span className="feature-icon" aria-hidden="true"><IconUsers /></span>
            <h3>Follow builders</h3>
            <p>Plain follows, chronological feed. No engagement algorithms deciding what you see.</p>
          </article>
          <article className="feature">
            <span className="feature-icon" aria-hidden="true"><IconTerminal /></span>
            <h3>Notifications that matter</h3>
            <p>Replies, reposts and follows from people you chose. No algorithmic noise, ever.</p>
          </article>
        </div>
      </div>
    </section>
  )
}

function Cta() {
  return (
    <section className="section cta">
      <div className="container">
        <h2>Start posting in minutes</h2>
        <p>Create an account, follow a few builders and make your first post today.</p>
        <Link className="btn btn-primary" to="/register">Create account</Link>
      </div>
    </section>
  )
}

export default function Home() {
  return (
    <>
      <Hero />
      <Features />
      <Cta />
    </>
  )
}
