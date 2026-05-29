package filter

import (
	"fmt"
	"strings"
)

// ToSQL converts a FilterSet into parameterized SQL fragments.
// baseCondition is an optional literal SQL predicate prepended to the WHERE
// clause (e.g. "c.status = 'published'"). Pass "" if not needed.
// startArgIdx is the 1-based parameter index to start from.
func ToSQL(fs *FilterSet, baseCondition string, startArgIdx int) SQLResult {
	var res SQLResult
	idx := startArgIdx
	var whereParts []string

	if baseCondition != "" {
		whereParts = append(whereParts, baseCondition)
	}

	for _, cond := range fs.Conditions {
		switch cond.Op {
		case OpEq:
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", cond.DBColumn, idx))
			res.Args = append(res.Args, cond.Value)
			idx++
		case OpNeq:
			whereParts = append(whereParts, fmt.Sprintf("%s != $%d", cond.DBColumn, idx))
			res.Args = append(res.Args, cond.Value)
			idx++
		case OpGt:
			whereParts = append(whereParts, fmt.Sprintf("%s > $%d", cond.DBColumn, idx))
			res.Args = append(res.Args, cond.Value)
			idx++
		case OpGte:
			whereParts = append(whereParts, fmt.Sprintf("%s >= $%d", cond.DBColumn, idx))
			res.Args = append(res.Args, cond.Value)
			idx++
		case OpLt:
			whereParts = append(whereParts, fmt.Sprintf("%s < $%d", cond.DBColumn, idx))
			res.Args = append(res.Args, cond.Value)
			idx++
		case OpLte:
			whereParts = append(whereParts, fmt.Sprintf("%s <= $%d", cond.DBColumn, idx))
			res.Args = append(res.Args, cond.Value)
			idx++
		case OpLike:
			whereParts = append(whereParts, fmt.Sprintf("%s ILIKE $%d", cond.DBColumn, idx))
			res.Args = append(res.Args, "%"+cond.Value.(string)+"%")
			idx++
		case OpArrayContains:
			whereParts = append(whereParts, fmt.Sprintf("$%d = ANY(%s)", idx, cond.DBColumn))
			res.Args = append(res.Args, cond.Value)
			idx++
		case OpIn:
			placeholders := make([]string, len(cond.Values))
			for i, v := range cond.Values {
				placeholders[i] = fmt.Sprintf("$%d", idx)
				res.Args = append(res.Args, v)
				idx++
			}
			whereParts = append(whereParts, fmt.Sprintf("%s IN (%s)", cond.DBColumn, strings.Join(placeholders, ", ")))
		}
	}

	if len(whereParts) > 0 {
		res.WhereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	// ORDER BY
	if fs.Sort.IsDefault {
		res.OrderClause = fmt.Sprintf("ORDER BY %s", fs.Sort.DBExpr)
	} else {
		res.OrderClause = fmt.Sprintf("ORDER BY %s %s", fs.Sort.DBExpr, fs.Sort.Order)
	}

	// LIMIT / OFFSET
	res.LimitClause = fmt.Sprintf("LIMIT $%d OFFSET $%d", idx, idx+1)
	res.Args = append(res.Args, fs.Pagination.Limit(), fs.Pagination.Offset())
	idx += 2

	res.NextArgIdx = idx
	return res
}

// CountSQL returns the WHERE clause and args for a COUNT query (no ORDER BY or LIMIT).
func CountSQL(fs *FilterSet, baseCondition string, startArgIdx int) (whereClause string, args []any, nextArgIdx int) {
	full := ToSQL(fs, baseCondition, startArgIdx)
	// Strip the 2 LIMIT/OFFSET args from the end
	countArgs := full.Args
	if len(countArgs) >= 2 {
		countArgs = countArgs[:len(countArgs)-2]
	}
	return full.WhereClause, countArgs, full.NextArgIdx - 2
}
