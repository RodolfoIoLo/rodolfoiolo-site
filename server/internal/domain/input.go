package domain

import "time"

type ArticleFilter struct {
	Page, PageSize int
	Category       string
	Tag            string
	Keyword        string
	Status         string
	Sort           string
	IncludePrivate bool
}

type ArticleInput struct {
	Slug       string
	Title      string
	Summary    string
	Content    string
	CoverImage *string
	CategoryID *int64
	TagIDs     []int64
	Status     string
	IsTop      bool
}

type ProjectFilter struct {
	Page, PageSize int
	Technology     string
	Status         string
	Featured       *bool
	IncludePrivate bool
}

type ProjectInput struct {
	Slug          string
	Name          string
	Summary       string
	Content       string
	CoverImage    *string
	Technologies  []string
	RepositoryURL *string
	DemoURL       *string
	Featured      bool
	SortOrder     int
	Status        string
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

type CategoryInput struct {
	Name        string
	Slug        string
	Description *string
	Icon        *string
	ParentID    *int64
	SortOrder   int
}

type TagInput struct {
	Name  string
	Slug  string
	Color *string
}

type CommentInput struct {
	ArticleSlug string
	ParentID    *int64
	Nickname    string
	EmailHash   *string
	Content     string
	VisitorID   string
	IPHash      string
	UserAgent   string
}
