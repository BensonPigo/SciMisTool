/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: '#e60023',
        primaryDark: '#c0001d',
        accent: '#ff6f61',
        info: '#007bff',
        background: '#f9f9f9'
      }
    }
  },
  plugins: [],
};
