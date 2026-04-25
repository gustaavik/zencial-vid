package filter

import "github.com/zenfulcode/zencial/internal/domain/valueobject"

// Op represents a filter operation.
type Op string

const (
	OpEq   Op = "eq"
	OpNeq  Op = "neq"
	OpGt   Op = "gt"
	OpGte  Op = "gte"
	OpLt   Op = "lt"
	OpLte  Op = "lte"
	OpLike Op = "like"
	OpIn   Op = "in"
)

// ValueType controls how raw string values are coerced.
type ValueType int

const (
	TypeString ValueType = iota
	TypeInt
	TypeBool
	TypeUUID
)

// ColumnDef describes one filterable column.
type ColumnDef struct {
	// DBColumn is the SQL column expression (e.g. "c.release_year").
	DBColumn string
	// AllowedOps lists which operations this column supports.
	AllowedOps []Op
	// Type controls value coercion from the raw query string.
	Type ValueType
}

// SortDef maps an API sort name to a SQL expression.
type SortDef struct {
	DBColumn string
}

// Config is the per-resource whitelist defining filterable/sortable columns.
// Used by FromRequest to parse URL params and by ToSQL to generate SQL.
type Config struct {
	// Columns maps URL param names to column definitions.
	// Only params listed here are accepted; others are ignored.
	Columns map[string]ColumnDef

	// SortColumns maps API sort field names to SQL expressions.
	SortColumns map[string]SortDef

	// DefaultSort is the ORDER BY expression when no valid sort_by is given.
	DefaultSort string
}

// Condition is a single parsed filter condition.
type Condition struct {
	DBColumn string
	Op       Op
	Value    any   // for eq/neq/gt/gte/lt/lte/like
	Values   []any // for in
}

// Sort holds the parsed sort directive.
type Sort struct {
	DBExpr    string // e.g. "c.title"
	Order     string // "ASC" or "DESC"
	IsDefault bool   // true when using DefaultSort (already includes direction)
}

// FilterSet is the fully parsed set of filters + sort + pagination.
type FilterSet struct {
	Conditions []Condition
	Sort       Sort
	Pagination valueobject.Pagination
}

// SQLResult holds generated SQL fragments and arguments.
type SQLResult struct {
	// WhereClause is a SQL fragment starting with "WHERE ..." or empty.
	WhereClause string
	// OrderClause is "ORDER BY ...".
	OrderClause string
	// LimitClause is "LIMIT $N OFFSET $M".
	LimitClause string
	// Args is the ordered parameter values matching $1..$N.
	Args []any
	// NextArgIdx is the next available parameter index.
	NextArgIdx int
}
