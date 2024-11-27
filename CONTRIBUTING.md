# Contributing to Better Evidence Project

## Project Overview

The Better Evidence Project is a web application designed to curate and provide useful evidence to inform decision-making in peacebuilding. It features search functionality, dynamic filters, and well-structured API endpoints for efficient data retrieval.

For additional information, check out:
- README.md - Project goals and setup instructions.
- API.md - Documentation of API endpoints and usage.

## Dependencies

To work on this project, you’ll need the following dependencies:
- Go (v1.19+): [Install Go](https://go.dev/doc/install) or [brew](https://formulae.brew.sh/formula/go)
- Echo (Web Framework): [Echo Documentation](https://echo.labstack.com/docs)
- TailwindCSS (for styling): [Tailwind Documentation](https://tailwindcss.com/docs/installation)
- Flowbite (UI component library for Tailwind): [Flowbite Documentation](https://flowbite.com/docs/getting-started/introduction/)

## Workflow Guidelines

### Pull Requests
1. Open a pull request against the main branch.
2. All PRs should:
	- Be rebased onto the latest main.
	- Be squashed during merging (use “Rebase and Merge”).
	- Include a detailed description of the changes and their purpose.
3. Resolve conflicts before submitting the PR.

## Setting Up the Development Environment

1.	Clone the repository:
```bash
git clone https://github.com/DSSD-Madison/gmu.git
cd project-name
```
2.	Install dependencies:
	- Go (v1.19+): Install Go
	- Echo: Already included via Go modules (see go.mod).
	- TailwindCSS: Set up through tailwind.config.js in the static/css directory.
	- Flowbite: Integrated directly into views/layouts/base.html.
3.	Run the development server:
```bash
go run main.go
```
4.	Optional: Run Air for hot reloads (if installed):
	- [Install Air](https://github.com/air-verse/air)
```bash
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s
```
	- Run air
```bash
./bin/air
```