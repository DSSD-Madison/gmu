package components

import (
	"reflect"
	"testing"

	"github.com/DSSD-Madison/gmu/pkg/model/search"
)

func Test_getNumberOfButtons(t *testing.T) {
	type args struct {
		totalPages int
		maxPages   int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "20 pages with 5 max",
			args: args{
				totalPages: 20,
				maxPages:   5,
			},
			want: 5,
		},
		{
			name: "5 pages with 5 max",
			args: args{
				totalPages: 5,
				maxPages:   5,
			},
			want: 5,
		},
		{
			name: "3 pages with 5 max",
			args: args{
				totalPages: 3,
				maxPages:   5,
			},
			want: 3,
		},
		{
			name: "1 pages with 5 max",
			args: args{
				totalPages: 1,
				maxPages:   5,
			},
			want: 1,
		},
		{
			name: "1 pages with 5 max",
			args: args{
				totalPages: 5,
				maxPages:   1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNumberOfButtons(tt.args.totalPages, tt.args.maxPages); got != tt.want {
				t.Errorf("getNumberOfButtons() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLowerPageNumber(t *testing.T) {
	type args struct {
		currentPage int
		totalPages  int
		maxPages    int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "First page with 5 max",
			args: args{
				currentPage: 1,
				totalPages:  20,
				maxPages:    5,
			},
			want: 1, // (1) 2 3 4 5
		},
		{
			name: "Second page with 5 max",
			args: args{
				currentPage: 2,
				totalPages:  20,
				maxPages:    5,
			},
			want: 1, // 1 (2) 3 4 5
		},
		{
			name: "Third page with 5 max",
			args: args{
				currentPage: 3,
				totalPages:  20,
				maxPages:    5,
			},
			want: 1, // 1 2 (3) 4 ... 20
		},
		{
			name: "Fifth page with 5 max",
			args: args{
				currentPage: 5,
				totalPages:  20,
				maxPages:    5,
			},
			want: 3, // 1 ... 3 4 (5) 6 7 ... 20
		},
		{
			name: "19th page with 5 max",
			args: args{
				currentPage: 19,
				totalPages:  20,
				maxPages:    5,
			},
			want: 16, // 1 ... 16 17 18 (19) 20
		},
		{
			name: "20th page with 5 max",
			args: args{
				currentPage: 20,
				totalPages:  20,
				maxPages:    5,
			},
			want: 16, // 1 ... 16 17 18 19 (20)
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLowerPageNumber(tt.args.currentPage, tt.args.totalPages, tt.args.maxPages); got != tt.want {
				t.Errorf("getLowerPageNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUpperPageNumber(t *testing.T) {
	type args struct {
		currentPage int
		totalPages  int
		maxPages    int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "First page with 5 max",
			args: args{
				currentPage: 1,
				totalPages:  20,
				maxPages:    5,
			},
			want: 5, // (1) 2 3 4 5 ... 20
		},
		{
			name: "Third page with 5 max",
			args: args{
				currentPage: 3,
				totalPages:  20,
				maxPages:    5,
			},
			want: 5, // 1 2 (3) 4 5 ... 20
		},
		{
			name: "Fifth page with 5 max",
			args: args{
				currentPage: 5,
				totalPages:  20,
				maxPages:    5,
			},
			want: 7, // 1 ... 3 4 (5) 6 7 ... 20
		},
		{
			name: "20th page with 5 max",
			args: args{
				currentPage: 20,
				totalPages:  20,
				maxPages:    5,
			},
			want: 20, // 1 ... 16 17 18 19 (20)
		},
		{
			name: "18th page with 5 max",
			args: args{
				currentPage: 18,
				totalPages:  20,
				maxPages:    5,
			},
			want: 20, // 1 ... 16 17 (18) 19 20
		},
		{
			name: "16th page with 5 max",
			args: args{
				currentPage: 16,
				totalPages:  20,
				maxPages:    5,
			},
			want: 18, // 1 ... 14 15 (16) 17 18 ... 20
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUpperPageNumber(tt.args.currentPage, tt.args.totalPages, tt.args.maxPages); got != tt.want {
				t.Errorf("getUpperPageNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUpperOverflow(t *testing.T) {
	type args struct {
		upperPage  int
		totalPages int
		maxPages   int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "16th page with 5 max",
			args: args{
				upperPage:  16,
				totalPages: 20,
				maxPages:   5,
			},
			want: true, // 1 ... 14 15 (16) 17 18 ... 20
		},
		{
			name: "18th page with 5 max",
			args: args{
				upperPage:  20,
				totalPages: 20,
				maxPages:   5,
			},
			want: false, // 1 ... 16 17 (18) 19 20
		},
		{
			name: "20th page with 5 max",
			args: args{
				upperPage:  20,
				totalPages: 20,
				maxPages:   5,
			},
			want: false, // 1 ... 16 17 18 19 (20)
		},
		{
			name: "First page with 5 max",
			args: args{
				upperPage:  5,
				totalPages: 20,
				maxPages:   5,
			},
			want: true, // (1) 2 3 4 5 ... 20
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUpperOverflow(tt.args.upperPage, tt.args.totalPages, tt.args.maxPages); got != tt.want {
				t.Errorf("getUpperOverflow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLowerOverflow(t *testing.T) {
	type args struct {
		currentPage int
		maxPages    int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "First page with 5 max",
			args: args{
				currentPage: 1,
				maxPages:    5,
			},
			want: false, // (1) 2 3 4 5 ... 20
		},
		{
			name: "Third page with 5 max",
			args: args{
				currentPage: 3,
				maxPages:    5,
			},
			want: false, // 1 2 (3) 4 5 ... 20
		},
		{
			name: "Fifth page with 5 max",
			args: args{
				currentPage: 5,
				maxPages:    5,
			},
			want: true, // 1 ... 3 4 (5) 6 7 ... 20
		},
		{
			name: "Tenth page with 5 max",
			args: args{
				currentPage: 10,
				maxPages:    5,
			},
			want: true, // 1 ... 8 9 (10) 11 12 ... 20
		},
		{
			name: "16th page with 5 max",
			args: args{
				currentPage: 16,
				maxPages:    5,
			},
			want: true, // 1 ... 14 15 (16) 17 18 ... 20
		},
		{
			name: "18th page with 5 max",
			args: args{
				currentPage: 18,
				maxPages:    5,
			},
			want: true, // 1 ... 16 17 (18) 19 20
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLowerOverflow(tt.args.currentPage, tt.args.maxPages); got != tt.want {
				t.Errorf("getLowerOverflow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPaginationVM(t *testing.T) {
	type args struct {
		status   search.PageStatus
		maxPages int
	}
	tests := []struct {
		name string
		args args
		want paginationViewModel
	}{
		{
			name: "Lower end of range",
			args: args{
				status: search.PageStatus{
					CurrentPage: 1,
					HasPrev:     false,
					HasNext:     true,
					PrevPage:    0,
					NextPage:    2,
					TotalPages:  20,
				},
				maxPages: 5,
			},
			want: paginationViewModel{
				sideCount:      2,
				maxPages:       5,
				upperPage:      5,
				lowerPage:      1,
				buttonCount:    5,
				lower_overflow: false,
				upper_overflow: true,
			},
		},
		{
			name: "Middle of range",
			args: args{
				status: search.PageStatus{
					CurrentPage: 10,
					HasPrev:     true,
					HasNext:     true,
					PrevPage:    9,
					NextPage:    11,
					TotalPages:  20,
				},
				maxPages: 5,
			},
			want: paginationViewModel{
				sideCount:      2,
				maxPages:       5,
				upperPage:      12,
				lowerPage:      8,
				buttonCount:    5,
				lower_overflow: true,
				upper_overflow: true,
			},
		},
		{
			name: "Upper end of range",
			args: args{
				status: search.PageStatus{
					CurrentPage: 20,
					HasPrev:     true,
					HasNext:     false,
					PrevPage:    19,
					NextPage:    0,
					TotalPages:  20,
				},
				maxPages: 5,
			},
			want: paginationViewModel{
				sideCount:      2,
				maxPages:       5,
				upperPage:      20,
				lowerPage:      16,
				buttonCount:    5,
				lower_overflow: true,
				upper_overflow: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPaginationVM(tt.args.status, tt.args.maxPages); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPaginationVM() = %v, want %v", got, tt.want)
			}
		})
	}
}
