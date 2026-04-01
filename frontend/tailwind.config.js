/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        navy:  { DEFAULT: '#0d1b2a', light: '#1a2e45', dark: '#07111b' },
        gold:  { DEFAULT: '#f5c842', light: '#fbd96a', dark: '#c9a020' },
        ocean: { DEFAULT: '#1e6091', light: '#2980b9' },
        straw: '#e8d5b0',
      },
      fontFamily: {
        pirate: ['"Pirata One"', 'cursive'],
        body:   ['"Nunito"', 'sans-serif'],
      },
    },
  },
  plugins: [],
}

