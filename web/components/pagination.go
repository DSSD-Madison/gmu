package components

import (
	"github.com/DSSD-Madison/gmu/pkg/awskendra"
)

// max=5 & total=4, show only 4 buttons
func getNumberOfButtons(totalPages, maxPages int) int {
	if totalPages <= maxPages {
		return totalPages
	} else {
		return maxPages
	}
}

// (1) 2 3 4 5
func getLowerPageNumber(currentPage, totalPages, maxPages int) int {
	sideCount := maxPages / 2

	if currentPage-sideCount < 1 {
		return 1
	} else {
		if currentPage+sideCount > totalPages {
			return totalPages - (maxPages - 1)
		} else {
			return currentPage - sideCount
		}
	}
}

// 1 (2) 3 4 5
func getUpperPageNumber(currentPage, totalPages, maxPages int) int {
	sideCount := maxPages / 2

	if currentPage+sideCount > totalPages {
		return totalPages
	} else {
		if currentPage-sideCount < 1 {
			return maxPages
		} else {
			return currentPage + sideCount
		}
	}
}

type paginationViewModel struct {
	sideCount      int  // number of buttons to show on each side
	maxPages       int  // max pages buttons to show
	upperPage      int  // upper value of page range
	lowerPage      int  // bottom value of page range
	buttonCount    int  // calculated total of buttons to be shown
	lower_overflow bool // used to determine if ... should be shown to indicate inbetween pages
	upper_overflow bool // used to determine if ... should be shown to indicate inbetween pages
}

func getUpperOverflow(currentPage, totalPages, maxPages int) bool {
	// if range of currenPage + sideCount doesnt include totalPages, then overflow
	sideCount := maxPages / 2
	if currentPage+sideCount < totalPages {
		return true
	} else {
		return false
	}
}

func getLowerOverflow(currentPage, maxPages int) bool {
	// if range of currenPage - sideCount doesnt include 1, then overflow
	sideCount := maxPages / 2
	if currentPage-sideCount > 1 {
		return true
	} else {
		return false
	}
}

func getPaginationVM(status awskendra.PageStatus, maxPages int) paginationViewModel {
	sideCount := maxPages / 2
	upperPage := getUpperPageNumber(status.CurrentPage, status.TotalPages, maxPages)
	lowerPage := getLowerPageNumber(status.CurrentPage, status.TotalPages, maxPages)
	buttonCount := getNumberOfButtons(status.TotalPages, maxPages)
	upper_overflow := getUpperOverflow(status.CurrentPage, status.TotalPages, maxPages)
	lower_overflow := getLowerOverflow(status.CurrentPage, maxPages)

	return paginationViewModel{
		sideCount:      sideCount,
		maxPages:       maxPages,
		upperPage:      upperPage,
		lowerPage:      lowerPage,
		buttonCount:    buttonCount,
		upper_overflow: upper_overflow,
		lower_overflow: lower_overflow,
	}
}
