package sql_model

import (
	"time"

	sb "github.com/huandu/go-sqlbuilder"
)

type User struct {
	ID        int64      `db:"id"        json:"id"`
	Name      string     `db:"name"      json:"name"`
	Email     string     `db:"email"     json:"email"`
	Password  string     `db:"password"  json:"-"`
	CreatedAt *time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt *time.Time `db:"updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `db:"deletedAt" json:"-"`
}

var UserAllowedSort = []string{"name", "email", "createdAt", "updatedAt"}

type UserFilter struct {
	DefaultFilter
	Name *string
}

func (f *UserFilter) Query(q *sb.SelectBuilder) {

	// default query
	f.DefaultFilter.DefaultQuery(q)

	if f.Name != nil {
		q.Where(
			q.Equal("name", *f.Name),
		)
	}
}
