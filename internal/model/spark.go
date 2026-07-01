package model

import "time"

type Spark struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	Important   bool       `json:"important"`
	Done        bool       `json:"done"`
	Deleted     bool       `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type ListOptions struct {
	IncludeDone bool
	IncludeAll  bool
}
