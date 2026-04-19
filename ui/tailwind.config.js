/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        background: 'var(--background)',
        foreground: 'var(--foreground)',
        surface: '#0e131e',
        'surface-dim': '#0e131e',
        'surface-bright': '#343945',
        'surface-container-lowest': '#080e18',
        'surface-container-low': '#161c26',
        'surface-container': '#1a202a',
        'surface-container-high': '#242a35',
        'surface-container-highest': '#2f3540',
        'on-surface': '#dde2f1',
        'on-surface-variant': '#bbcbb8',
        primary: '#3fe56c',
        'primary-container': '#00c853',
        'on-primary': '#003912',
        secondary: '#8ed793',
        'secondary-container': '#02531e',
        'on-secondary-container': '#7ec583',
        error: '#ffb4ab',
        'error-container': '#93000a',
        tertiary: '#ffb7ae',
        'on-tertiary-container': '#76251f',
        outline: '#869583',
        'outline-variant': '#3c4a3c',
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        display: ['Space Grotesk', 'sans-serif'],
      },
      animation: {
        'ping-slow': 'ping 2s cubic-bezier(0,0,0.2,1) infinite',
        'pulse-fast': 'pulse 1s cubic-bezier(0.4,0,0.6,1) infinite',
      },
    },
  },
  plugins: [],
}
