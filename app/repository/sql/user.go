package sqlrepo

import (
	sql_model "github.com/adkurnwn/gigsourcehub-general-api/domain/model/sql"
	"context"
	"time"

	sb "github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// for default query
func defaultUserQuery() *sb.SelectBuilder {
	sqb := sb.NewSelectBuilder()

	// default query
	sqb.Where(
		sqb.IsNull("deletedAt"),
	)

	return sqb
}

func (r *sqlRepo) FetchUser(ctx context.Context, options sql_model.UserFilter) (cur *sqlx.Rows, err error) {
	// generate query
	q := defaultUserQuery()

	options.Query(
		q.Select("*").
			From(r.userTable),
	)

	// build query
	sql, args := q.Build()

	cur, err = r.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		logrus.Error("FetchUser Find:", err)
		return
	}

	return
}

func (r *sqlRepo) FetchOneUser(ctx context.Context, options sql_model.UserFilter) (row *sql_model.User, err error) {
	// generate query
	q := defaultUserQuery()

	options.Query(
		q.Select("*").
			From(r.userTable).
			Limit(1),
	)

	// build query
	sql, args := q.Build()

	// set row
	row = new(sql_model.User)

	err = r.db.GetContext(ctx, row, sql, args...)
	if err != nil {
		logrus.Error("FetchOneUser Get:", err)
		return
	}

	return
}

func (r *sqlRepo) CountUser(ctx context.Context, options sql_model.UserFilter) (total int64) {
	// generate query
	q := defaultUserQuery()

	// remove limit and offset
	options.Limit = nil
	options.Offset = nil
	options.Sorts = make([]map[string]string, 0)

	options.Query(
		q.Select("COUNT(*)").
			From(r.userTable),
	)

	// build query
	sql, args := q.Build()

	err := r.db.GetContext(ctx, &total, sql, args...)
	if err != nil {
		logrus.Error("CountUser", err)
		return 0
	}

	return
}

func (r *sqlRepo) CreateUser(ctx context.Context, row *sql_model.User) (err error) {
	// set created at and updated at
	now := time.Now().UTC()
	row.CreatedAt = &now
	row.UpdatedAt = &now

	sql, args := sb.NewInsertBuilder().
		InsertInto(r.userTable).
		Cols("name", "email", "password", "createdAt", "updatedAt").
		Values(row.Name, row.Email, row.Password, now, now).
		Build()
	res, err := r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		logrus.Error("CreateUser Exec:", err)
		return
	}

	// set id
	row.ID, _ = res.LastInsertId()

	return
}
