import type { Config } from 'tailwindcss'

const config: Config = {
  content: ['./index.html', './src/**/*.{vue,ts,tsx,js,jsx}'],
  theme: {
    extend: {
      colors: {
        'void-black': '#030305',
        'text-hero': '#ECEFFE',
        'text-body': '#C9D2E5',
        'text-muted': '#8B94AA',
        'text-caption': '#6B7280',
        'champagne-gold': '#D4AF37',
        'royal-purple': '#4B0082',
        'midnight-blue': '#191970',
      },
      fontFamily: {
        sans: ['Montserrat', 'sans-serif'],
        serif: ['Cinzel', 'serif'],
        display: ['Playfair Display', 'serif'],
      },
      letterSpacing: {
        extra: '0.35em',
      },
      transitionTimingFunction: {
        'expo-out': 'cubic-bezier(0.19, 1, 0.22, 1)',
      },
    },
  },
  plugins: [],
}

export default config
