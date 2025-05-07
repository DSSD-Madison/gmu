package components

import "github.com/DSSD-Madison/gmu/internal/domain/search"

// max=5 & total=4, show only 4 buttons
func getNumberOfButtons(totalPages, maxPages int) int {
	if totalPages <= maxPages {
		return totalPages
	}
	return maxPages
}

// (1) 2 3 4 5
func getLowerPageNumber(currentPage, totalPages, maxPages int) int {
	sideCount := maxPages / 2

	if currentPage-sideCount < 1 {
		return 1
	}
	if currentPage+sideCount > totalPages {
		return totalPages - (maxPages - 1)
	}
	return currentPage - sideCount
}

// 1 (2) 3 4 5
func getUpperPageNumber(currentPage, totalPages, maxPages int) int {
	sideCount := maxPages / 2

	if currentPage+sideCount > totalPages {
		return totalPages
	}
	if currentPage-sideCount < 1 {
		return maxPages
	}
	return currentPage + sideCount
}

type paginationViewModel struct {
	sideCount     int  // number of buttons to show on each side
	maxPages      int  // max pages buttons to show
	upperPage     int  // upper value of page range
	lowerPage     int  // bottom value of page range
	buttonCount   int  // calculated total of buttons to be shown
	lowerOverflow bool // used to determine if ... should be shown to indicate inbetween pages
	upperOverflow bool // used to determine if ... should be shown to indicate inbetween pages
}

func getUpperOverflow(currentPage, totalPages, maxPages int) bool {
	// if range of currenPage + sideCount doesnt include totalPages, then overflow
	sideCount := maxPages / 2
	return currentPage+sideCount < totalPages
}

func getLowerOverflow(currentPage, maxPages int) bool {
	// if range of currenPage - sideCount doesnt include 1, then overflow
	sideCount := maxPages / 2
	return currentPage-sideCount > 1
}

func getPaginationVM(status search.PageStatus, maxPages int) paginationViewModel {
	sideCount := maxPages / 2
	upperPage := getUpperPageNumber(status.CurrentPage, status.TotalPages, maxPages)
	lowerPage := getLowerPageNumber(status.CurrentPage, status.TotalPages, maxPages)
	buttonCount := getNumberOfButtons(status.TotalPages, maxPages)
	upperOverflow := getUpperOverflow(status.CurrentPage, status.TotalPages, maxPages)
	lowerOverflow := getLowerOverflow(status.CurrentPage, maxPages)

	return paginationViewModel{
		sideCount:     sideCount,
		maxPages:      maxPages,
		upperPage:     upperPage,
		lowerPage:     lowerPage,
		buttonCount:   buttonCount,
		upperOverflow: upperOverflow,
		lowerOverflow: lowerOverflow,
	}
}
