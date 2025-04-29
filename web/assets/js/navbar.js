document.addEventListener("DOMContentLoaded", () => {
	// Select the mobile menu button using its aria-controls attribute
	const mobileMenuButton = document.querySelector(
		'button[aria-controls="mobile-menu"]',
	);
	// Select the mobile menu itself by its ID
	const mobileMenu = document.getElementById("mobile-menu");
	// Select the icons within the button
	const openIcon = mobileMenuButton?.querySelector("svg:nth-of-type(1)"); // First SVG is open (hamburger)
	const closeIcon = mobileMenuButton?.querySelector("svg:nth-of-type(2)"); // Second SVG is close (X)

	// Check if all elements were found
	if (mobileMenuButton && mobileMenu && openIcon && closeIcon) {
		mobileMenuButton.addEventListener("click", () => {
			// Check the current state via aria-expanded
			const isExpanded =
				mobileMenuButton.getAttribute("aria-expanded") === "true";

			// Toggle the aria-expanded attribute
			mobileMenuButton.setAttribute("aria-expanded", !isExpanded);

			// Toggle the 'hidden' class on the menu itself
			mobileMenu.classList.toggle("hidden");

			// Toggle the 'hidden' class on the icons to swap them
			openIcon.classList.toggle("hidden");
			closeIcon.classList.toggle("hidden");

			// Optional: Toggle 'block' as well if needed, though toggling 'hidden'
			// usually suffices if the default display is 'block' or 'inline-block'
			// openIcon.classList.toggle('block');
			// closeIcon.classList.toggle('block');
		});
	} else {
		console.error(
			"Mobile menu button, menu, or icons not found. Check selectors.",
		);
	}
});
