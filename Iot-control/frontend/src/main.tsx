import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import { getInitialTheme } from './components/ThemeToggle'
import './index.css'

// Đặt theme trước khi render để không bị "nháy" màu sai lúc tải trang.
document.documentElement.setAttribute('data-theme', getInitialTheme())

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
