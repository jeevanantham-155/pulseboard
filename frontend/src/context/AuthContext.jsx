import { createContext, useContext, useMemo, useState } from 'react'
import api from '../services/api'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => JSON.parse(localStorage.getItem('live_polling_user') || 'null'))
  const save = (data) => { localStorage.setItem('live_polling_token', data.token); localStorage.setItem('live_polling_user', JSON.stringify(data.user)); setUser(data.user) }
  const login = async (body) => save((await api.post('/auth/login', body)).data)
  const signup = async (body) => save((await api.post('/auth/signup', body)).data)
  const logout = () => { localStorage.removeItem('live_polling_token'); localStorage.removeItem('live_polling_user'); setUser(null) }
  return <AuthContext.Provider value={useMemo(() => ({ user, login, signup, logout }), [user])}>{children}</AuthContext.Provider>
}

export function useAuth() { return useContext(AuthContext) }
