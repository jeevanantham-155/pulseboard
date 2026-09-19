import { useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { messageFromError } from '../services/api'
import ThemeToggle from '../components/ThemeToggle'

export default function Login() {
  const { user, login } = useAuth(); const navigate = useNavigate(); const [form, setForm] = useState({ email: '', password: '' }); const [error, setError] = useState(''); const [loading, setLoading] = useState(false)
  if (user) return <Navigate to="/dashboard" replace />
  const submit = async (event) => { event.preventDefault(); setError(''); setLoading(true); try { await login(form); navigate('/dashboard') } catch (err) { setError(messageFromError(err, 'Unable to log in')) } finally { setLoading(false) } }
  return <main className="auth-page"><section className="auth-copy"><div className="auth-top"><Link className="brand" to="/">↗ Pulseboard</Link><ThemeToggle /></div><p className="eyebrow">LIVE POLLING</p><h1>Welcome back</h1><p>Sign in to see what your audience is saying.</p></section><section className="auth-card"><form className="form" onSubmit={submit}><label>Email<input type="email" required value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></label><label>Password<input type="password" required value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} /></label>{error && <p className="form-error">{error}</p>}<button className="button button-primary" disabled={loading}>{loading ? 'Signing in...' : 'Sign in'}</button><p className="form-note">New here? <Link to="/signup">Create an account</Link></p></form></section></main>
}
