import axios from 'axios'

const configuredApiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const apiBaseUrl = configuredApiUrl.replace(/\/$/, '').endsWith('/api') ? configuredApiUrl : `${configuredApiUrl.replace(/\/$/, '')}/api`
const api = axios.create({ baseURL: apiBaseUrl, headers: { 'Content-Type': 'application/json' } })
api.interceptors.request.use((config) => { const token = localStorage.getItem('live_polling_token'); if (token) config.headers.Authorization = `Bearer ${token}`; return config })
export default api
export function messageFromError(error, fallback = 'Something went wrong') { return error.response?.data?.message || fallback }
