package db

import (
	"strconv"
	"strings"
)

// Update builds a dynamic UPDATE query for PATCH endpoints.
// Zero value is ready to use.
//
//	var u db.Update
//	if input.Name.Set { u.Set("name", input.Name.Value) }
//	sql, args := u.Build("links", "id = ? AND workspace_id = ?", id, wsID)
type Update struct {
	cols []string
	args []any
}

// Set adds a column = $N assignment.
func (u *Update) Set(col string, val any) {
	u.args = append(u.args, val)
	u.cols = append(u.cols, col+" = $"+strconv.Itoa(len(u.args)))
}

// IsEmpty reports whether any columns were set.
func (u *Update) IsEmpty() bool { return len(u.cols) == 0 }

// Build returns the full UPDATE … SET … WHERE … RETURNING * query and args.
// Use ? as positional placeholders in the WHERE clause; they are rewritten
// to $N parameters in the correct sequence.
func (u *Update) Build(table, where string, whereArgs ...any) (string, []any) {
	var b strings.Builder
	b.WriteString("UPDATE ")
	b.WriteString(table)
	b.WriteString(" SET updated_at = NOW()")
	for _, col := range u.cols {
		b.WriteString(", ")
		b.WriteString(col)
	}
	b.WriteString(" WHERE ")

	for _, c := range where {
		if c == '?' {
			u.args = append(u.args, whereArgs[0])
			whereArgs = whereArgs[1:]
			b.WriteString("$")
			b.WriteString(strconv.Itoa(len(u.args)))
		} else {
			b.WriteRune(c)
		}
	}

	b.WriteString(" RETURNING *")
	return b.String(), u.args
}
