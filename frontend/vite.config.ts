import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  preview: {
    // Serve index.html for all routes so /join/:code works in preview mode
    // (The dev server already handles this by default in Vite)
  },
})
