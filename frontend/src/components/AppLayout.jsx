import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import ThemeToggle from './ThemeToggle'

export default function AppLayout() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  return <div className="app-frame"><header className="topbar"><Link className="brand" to="/dashboard">↗ Pulseboard</Link><nav className="nav-actions"><ThemeToggle /><span className="user-name">{user?.name}</span><button className="button button-quiet" onClick={() => { logout(); navigate('/login') }}>Log out</button></nav></header><Outlet /></div>
}
