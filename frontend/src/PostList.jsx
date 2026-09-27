import Markdown from 'react-markdown'

function PostCard({ post }) {
  const date = new Date(post.created_at).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })

  return (
    <article className="post-card-item">
      <h3 className="post-item-title">{post.title}</h3>
      <time className="post-item-date" dateTime={post.created_at}>
        {date}
      </time>
      <div className="markdown-body post-item-body">
        <Markdown>{post.content}</Markdown>
      </div>
    </article>
  )
}

export default function PostList({ posts, error }) {
  return (
    <section className="profile-posts" aria-label="Posts">
      {error ? (
        <p className="profile-posts-status" role="alert">
          Could not load posts.
        </p>
      ) : posts.length === 0 ? (
        <p className="profile-posts-status">No posts yet.</p>
      ) : (
        posts.map((post) => <PostCard key={post.id} post={post} />)
      )}
    </section>
  )
}
