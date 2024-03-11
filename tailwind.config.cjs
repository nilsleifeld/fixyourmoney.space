/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
	theme: {
		extend: {
			keyframes: {
				wiggle: {
					'0%, 100%': { transform: 'rotate(3deg)' },
					'5%, 95%': { transform: 'rotate(-3deg)' },
					'10%, 90%': { transform: 'rotate(0deg)' },
				}
			},
			animation: {
				wiggle: 'wiggle 2s infinite',
			}
		},
		fontFamily: {
			'display': ['Maven Pro', 'sans-serif'],
			'body': ['Maven Pro', 'sans-serif'],
			'sans': ['Maven Pro', 'sans-serif'],
		}
	},
	plugins: [],
}
