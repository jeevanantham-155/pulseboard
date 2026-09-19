import { Link, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { ThemeProvider } from './context/ThemeContext'
import AppLayout from './components/AppLayout'
import ProtectedRoute from './components/ProtectedRoute'
import CreatePoll from './pages/CreatePoll'
import Dashboard from './pages/Dashboard'
import Login from './pages/Login'
import PollManagement from './pages/PollManagement'
import PublicPoll from './pages/PublicPoll'
import Signup from './pages/Signup'

function Landing() {
  return <main className="landing"><section><Link className="brand" to="/">↗ Pulseboard</Link><p className="eyebrow">LIVE POLLING, WITHOUT THE LAG</p><h1>Ask the room.<br /><em>Hear it move.</em></h1><p className="landing-copy">Create a question, share one link, and watch every response reshape the conversation in real time.</p><div className="landing-actions"><Link className="button button-primary" to="/signup">Create a poll</Link><Link className="text-button" to="/login">Sign in →</Link></div></section><aside className="landing-note"><span className="signal">●</span><div><strong>Live by design</strong><p>Every vote travels from MongoDB to Redis Pub/Sub to every open screen.</p></div></aside></main>
}

export default function App() {
  return <ThemeProvider><AuthProvider><Routes><Route path="/" element={<Landing />} /><Route path="/login" element={<Login />} /><Route path="/signup" element={<Signup />} /><Route path="/poll/:id" element={<PublicPoll />} /><Route element={<ProtectedRoute />}><Route element={<AppLayout />}><Route path="/dashboard" element={<Dashboard />} /><Route path="/create-poll" element={<CreatePoll />} /><Route path="/polls/:id" element={<PollManagement />} /></Route></Route><Route path="*" element={<Navigate to="/" replace />} /></Routes></AuthProvider></ThemeProvider>
}
