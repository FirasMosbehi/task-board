import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  define: {
    'process.env': {}
  },
  optimizeDeps: {
    include: [
      '@opentelemetry/api',
      '@opentelemetry/sdk-trace-web',
      '@opentelemetry/sdk-trace-base',
      '@opentelemetry/exporter-trace-otlp-http',
      '@opentelemetry/instrumentation-fetch',
      '@opentelemetry/instrumentation-user-interaction',
      '@opentelemetry/instrumentation-document-load'
    ]
  }
})
