import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import api, { messageFromError } from '../services/api'

export default function Dashboard() {
  const [polls, setPolls] = useState([]); const [error, setError] = useState(''); const [loading, setLoading] = useState(true)
  useEffect(() => { api.get('/polls').then((r) => setPolls(r.data.polls || [])).catch((e) => setError(messageFromError(e, 'Unable to load polls'))).finally(() => setLoading(false)) }, [])
  const remove = async (id) => { if (!window.confirm('Delete this poll?')) return; try { await api.delete(`/polls/${id}`); setPolls((items) => items.filter((item) => item.id !== id)) } catch (err) { setError(messageFromError(err, 'Unable to delete poll')) } }
  return <main className="content"><div className="page-heading"><div><p className="eyebrow">YOUR WORKSPACE</p><h1>Polls that move at the speed of the room.</h1></div><Link className="button button-primary" to="/create-poll">＋ New poll</Link></div>{error && <p className="form-error">{error}</p>}{loading ? <p className="loading">Loading your polls...</p> : polls.length === 0 ? <section className="empty-state"><span className="empty-icon">＋</span><h2>Your first question is waiting.</h2><p>Create a poll, share the link, and watch responses arrive live.</p><Link className="button button-primary" to="/create-poll">Create a poll</Link></section> : <div className="poll-grid">{polls.map((item) => <article className="poll-card" key={item.id}><div className="card-top"><span className="status-badge">{item.status}</span><button className="icon-button" title="Delete poll" onClick={() => remove(item.id)}>×</button></div><h2>{item.question}</h2><p className="muted">{item.options.length} options · {item.options.reduce((sum, option) => sum + option.voteCount, 0)} votes</p><div className="card-actions"><Link to={`/polls/${item.id}`}>Manage</Link><Link to={`/poll/${item.id}`}>Open public view ↗</Link></div></article>)}</div>}</main>
}
