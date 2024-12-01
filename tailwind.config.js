/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./views/*.html"
    ],
    theme: {
        extend: {
            colors: {
                dark_green: '#005239', // dark green
                green: '#EFF0E7', // green 
                yellow: '#FEBE10', // yellow
                white: '#FFFFFF', // white
              },
        },
    },
    plugins: [],
}
