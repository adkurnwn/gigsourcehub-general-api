package sql_model

import (
	"time"

	sb "github.com/huandu/go-sqlbuilder"
)

type DefaultFilter struct {
	ID  int64
	IDs []int64

	CreatedAtGt    *time.Time
	CreatedAtGte   *time.Time
	CreatedAtLt    *time.Time
	CreatedAtLte   *time.Time
	CreatedAtRange *DatetimeRange

	UpdatedAtGt    *time.Time
	UpdatedAtGte   *time.Time
	UpdatedAtLt    *time.Time
	UpdatedAtLte   *time.Time
	UpdatedAtRange *DatetimeRange

	Limit  *int64
	Offset *int64
	Sorts  []map[string]string
}

type DatetimeRange struct {
	Start time.Time
	End   time.Time
}

func (f *DefaultFilter) DefaultQuery(q *sb.SelectBuilder) {

	if f.ID > 0 {
		q.Where(q.Equal("id", f.ID))
	} else if len(f.IDs) > 0 {
		q.Where(q.In("id", f.IDs))
	}

	// created at
	if f.CreatedAtGt != nil {
		q.Where(
			q.GT("createdAt", f.CreatedAtGt),
		)
	} else if f.CreatedAtGte != nil {
		q.Where(
			q.GTE("createdAt", f.CreatedAtGte),
		)
	}

	if f.CreatedAtLt != nil {
		q.Where(
			q.LT("createdAt", f.CreatedAtLt),
		)
	} else if f.CreatedAtLte != nil {
		q.Where(
			q.LTE("createdAt", f.CreatedAtLte),
		)
	}

	if f.CreatedAtRange != nil {
		q.Where(
			q.GTE("createdAt", f.CreatedAtGte),
			q.LTE("createdAt", f.CreatedAtLte),
		)
	}

	// updated at
	if f.UpdatedAtGt != nil {
		q.Where(
			q.GT("updatedAt", f.UpdatedAtGt),
		)
	} else if f.UpdatedAtGte != nil {
		q.Where(
			q.GTE("updatedAt", f.UpdatedAtGte),
		)
	}

	if f.UpdatedAtLt != nil {
		q.Where(
			q.LT("updatedAt", f.UpdatedAtLt),
		)
	} else if f.UpdatedAtLte != nil {
		q.Where(
			q.LTE("updatedAt", f.UpdatedAtLte),
		)
	}

	if f.UpdatedAtRange != nil {
		q.Where(
			q.GTE("updatedAt", f.UpdatedAtGte),
			q.LTE("updatedAt", f.UpdatedAtLte),
		)
	}

	// Limit & Offset
	if f.Limit != nil {
		q.Limit(int(*f.Limit))
	}

	if f.Offset != nil {
		q.Offset(int(*f.Offset))
	}

	// Sorts
	if len(f.Sorts) > 0 {
		for _, sort := range f.Sorts {
			for key, value := range sort {
				q.OrderBy(key, value)
			}
		}
	}

}
