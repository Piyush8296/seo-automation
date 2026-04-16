/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      borderColor: {
        border: 'var(--border)',
      },
      outlineColor: {
        ring: 'var(--ring)',
      },
      colors: {
        background: 'var(--background)',
        foreground: 'var(--foreground)',
      },
      fontFamily: {
        sans: ['Geist Variable', 'sans-serif'],
      },
      animation: {
        'ping-slow': 'ping 2s cubic-bezier(0,0,0.2,1) infinite',
        'pulse-fast': 'pulse 1s cubic-bezier(0.4,0,0.6,1) infinite',
      },
    },
  },
  plugins: [],
}
