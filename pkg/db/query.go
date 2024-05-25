package db

import (
	"strconv"
	"strings"
)

// Query helps build SQL queries.go using bind parameters.
// Use Query to construct parts of a query and use Param to add bind parameters.
// The final query and parameters can be retrieved using the Get method.
type Query struct {
	b      strings.Builder
	params []any
	err    error
}

// Unsafe writes a non-parameterized part of a query.
func (q *Query) Unsafe(s string) {
	q.b.WriteString(s)
}

// Param writes a parameterized part of a query.
func (q *Query) Param(count *int, v any) {
	*count++
	q.b.WriteString("$" + strconv.Itoa(*count))
	q.params = append(q.params, v)
}

// Params writes multiple parameterized parts of a query seperated by commas.
func (q *Query) Params(count *int, v ...any) {
	for i, p := range v {
		*count++
		if i > 0 {
			q.b.WriteString(", ")
		}
		q.b.WriteString("$" + strconv.Itoa(*count))
		q.params = append(q.params, p)
	}
}

// Get returns the constructed query and parameter values.
func (q *Query) Get() (string, []any, error) {
	return q.b.String(), q.params, q.err
}
