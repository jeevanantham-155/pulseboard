import { useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { messageFromError } from '../services/api'
import ThemeToggle from '../components/ThemeToggle'

export default function Signup() {
  const { user, signup } = useAuth(); const navigate = useNavigate(); const [form, setForm] = useState({ name: '', email: '', password: '' }); const [error, setError] = useState(''); const [loading, setLoading] = useState(false)
  if (user) return <Navigate to="/dashboard" replace />
  const submit = async (event) => { event.preventDefault(); setError(''); setLoading(true); try { await signup(form); navigate('/dashboard') } catch (err) { setError(messageFromError(err, 'Unable to create account')) } finally { setLoading(false) } }
  return <main className="auth-page"><section className="auth-copy"><div className="auth-top"><Link className="brand" to="/">↗ Pulseboard</Link><ThemeToggle /></div><p className="eyebrow">START A CONVERSATION</p><h1>Make every voice count.</h1><p>Create your workspace and turn a question into a live room.</p></section><section className="auth-card"><h2>Create account</h2><form className="form" onSubmit={submit}><label>Name<input required minLength="2" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></label><label>Email<input type="email" required value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></label><label>Password<input type="password" required minLength="8" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} /></label>{error && <p className="form-error">{error}</p>}<button className="button button-primary" disabled={loading}>{loading ? 'Creating...' : 'Create account'}</button><p className="form-note">Already have an account? <Link to="/login">Sign in</Link></p></form></section></main>
}
