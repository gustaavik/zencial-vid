package valueobject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPagination(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		perPage     int
		wantPage    int
		wantPerPage int
	}{
		{
			name:        "valid values",
			page:        2,
			perPage:     25,
			wantPage:    2,
			wantPerPage: 25,
		},
		{
			name:        "default page and perPage",
			page:        1,
			perPage:     20,
			wantPage:    DefaultPage,
			wantPerPage: DefaultPerPage,
		},
		{
			name:        "page less than 1 defaults to 1",
			page:        0,
			perPage:     20,
			wantPage:    DefaultPage,
			wantPerPage: DefaultPerPage,
		},
		{
			name:        "negative page defaults to 1",
			page:        -5,
			perPage:     20,
			wantPage:    DefaultPage,
			wantPerPage: DefaultPerPage,
		},
		{
			name:        "perPage less than 1 defaults to 20",
			page:        1,
			perPage:     0,
			wantPage:    DefaultPage,
			wantPerPage: DefaultPerPage,
		},
		{
			name:        "negative perPage defaults to 20",
			page:        1,
			perPage:     -10,
			wantPage:    DefaultPage,
			wantPerPage: DefaultPerPage,
		},
		{
			name:        "perPage exceeding max clamped to 100",
			page:        1,
			perPage:     200,
			wantPage:    DefaultPage,
			wantPerPage: MaxPerPage,
		},
		{
			name:        "perPage exactly at max",
			page:        1,
			perPage:     100,
			wantPage:    DefaultPage,
			wantPerPage: MaxPerPage,
		},
		{
			name:        "both invalid defaults applied",
			page:        -1,
			perPage:     -1,
			wantPage:    DefaultPage,
			wantPerPage: DefaultPerPage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPagination(tt.page, tt.perPage)
			assert.Equal(t, tt.wantPage, p.Page)
			assert.Equal(t, tt.wantPerPage, p.PerPage)
		})
	}
}

func TestPagination_Offset(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		perPage int
		want    int
	}{
		{"first page", 1, 20, 0},
		{"second page", 2, 20, 20},
		{"third page", 3, 10, 20},
		{"large page", 10, 50, 450},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPagination(tt.page, tt.perPage)
			assert.Equal(t, tt.want, p.Offset())
		})
	}
}

func TestPagination_Limit(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		perPage int
		want    int
	}{
		{"default perPage", 1, 20, 20},
		{"custom perPage", 1, 50, 50},
		{"max perPage", 1, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPagination(tt.page, tt.perPage)
			assert.Equal(t, tt.want, p.Limit())
		})
	}
}

func TestPagination_TotalPages(t *testing.T) {
	tests := []struct {
		name       string
		perPage    int
		totalItems int64
		want       int
	}{
		{"zero items", 20, 0, 0},
		{"exactly one page", 20, 20, 1},
		{"partial second page", 20, 21, 2},
		{"multiple full pages", 10, 50, 5},
		{"one item", 20, 1, 1},
		{"large item count", 25, 101, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPagination(1, tt.perPage)
			assert.Equal(t, tt.want, p.TotalPages(tt.totalItems))
		})
	}
}
