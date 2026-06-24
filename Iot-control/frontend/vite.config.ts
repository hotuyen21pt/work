import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import basicSsl from '@vitejs/plugin-basic-ssl'

export default defineConfig({
  plugins: [react(), basicSsl()],
  server: {
    // Lắng nghe trên mọi địa chỉ để thiết bị khác (qua VPN/LAN) truy cập được.
    host: true,
    port: 5173,
    // Cho phép mọi host (dev): truy cập qua IP LAN/VPN và Cloudflare Tunnel.
    allowedHosts: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
