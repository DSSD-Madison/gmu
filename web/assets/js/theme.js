;(function() {
	const stored = localStorage.getItem('theme')
	const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

	if (stored === 'dark' || (stored === null && prefersDark)) {
		document.documentElement.classList.add('dark')
	} else {
		document.documentElement.classList.remove('dark')
	}
})()

function toggleTheme() {
	const html = document.documentElement; // Get the main html element
	const isNowDark = html.classList.toggle('dark')
	localStorage.setItem('theme', isNowDark ? 'dark' : 'light')

}
