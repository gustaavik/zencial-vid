package filter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/zenfulcode/zencial/internal/domain/valueobject"
)

// FromRequest parses URL query parameters into a FilterSet according to cfg.
// Unknown params are silently ignored. Invalid values return an error.
func FromRequest(r *http.Request, cfg Config) (FilterSet, error) {
	var fs FilterSet
	query := r.URL.Query()

	for paramName, colDef := range cfg.Columns {
		for _, op := range colDef.AllowedOps {
			raw := ""

			// Plain param (e.g. ?type=film) for eq/in
			if op == OpEq || op == OpIn {
				raw = query.Get(paramName)
			}

			// Bracket syntax (e.g. ?release_year[gte]=2020) for all ops
			bracketKey := fmt.Sprintf("%s[%s]", paramName, string(op))
			if v := query.Get(bracketKey); v != "" {
				raw = v
			}

			if raw == "" {
				continue
			}

			cond, err := buildCondition(colDef, op, raw)
			if err != nil {
				return fs, fmt.Errorf("invalid filter %s: %w", paramName, err)
			}
			fs.Conditions = append(fs.Conditions, cond)
		}
	}

	// Parse sort
	sortBy := query.Get("sort_by")
	sortOrder := strings.ToUpper(query.Get("sort_order"))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	if sd, ok := cfg.SortColumns[sortBy]; ok {
		fs.Sort = Sort{DBExpr: sd.DBColumn, Order: sortOrder}
	} else {
		fs.Sort = Sort{DBExpr: cfg.DefaultSort, IsDefault: true}
	}

	// Parse pagination
	page := queryInt(query, "page", valueobject.DefaultPage)
	perPage := queryInt(query, "per_page", valueobject.DefaultPerPage)
	fs.Pagination = valueobject.NewPagination(page, perPage)

	return fs, nil
}

func buildCondition(col ColumnDef, op Op, raw string) (Condition, error) {
	c := Condition{DBColumn: col.DBColumn, Op: op}

	// Comma-separated values → IN
	if op == OpIn || (op == OpEq && strings.Contains(raw, ",")) {
		c.Op = OpIn
		parts := strings.Split(raw, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			v, err := coerce(p, col.Type)
			if err != nil {
				return c, err
			}
			c.Values = append(c.Values, v)
		}
		if len(c.Values) == 0 {
			return c, fmt.Errorf("empty value list")
		}
		return c, nil
	}

	v, err := coerce(raw, col.Type)
	if err != nil {
		return c, err
	}
	c.Value = v
	return c, nil
}

func coerce(raw string, vt ValueType) (interface{}, error) {
	switch vt {
	case TypeInt:
		return strconv.Atoi(raw)
	case TypeBool:
		return strconv.ParseBool(raw)
	case TypeUUID:
		return uuid.Parse(raw)
	default:
		return raw, nil
	}
}

func queryInt(q map[string][]string, key string, defaultVal int) int {
	vals, ok := q[key]
	if !ok || len(vals) == 0 || vals[0] == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(vals[0])
	if err != nil {
		return defaultVal
	}
	return v
}
