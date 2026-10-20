package domain

import "time"

type Category struct {
	ID           int64
	Name         string
	Slug         string
	Description  *string
	Icon         *string
	ParentID     *int64
	SortOrder    int
	ArticleCount int
	CreatedAt    time.Time
}

type Tag struct {
	ID           int64
	Name         string
	Slug         string
	Color        *string
	ArticleCount int
	CreatedAt    time.Time
}

type Article struct {
	ID           int64
	Slug         string
	Title        string
	Summary      string
	Content      string
	CoverImage   *string
	Category     *Category
	Tags         []Tag
	Author       UserSummary
	Status       string
	IsTop        bool
	ViewCount    int
	LikeCount    int
	CommentCount int
	PublishedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserSummary struct {
	ID        int64
	Username  string
	Nickname  string
	AvatarURL *string
	Bio       *string
	Role      string
	CreatedAt time.Time
}

type Project struct {
	ID            int64
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
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Page[T any] struct {
	Items      []T
	Page       int
	PageSize   int
	Total      int
	TotalPages int
}

type Comment struct {
	ID        int64
	ArticleID int64
	ParentID  *int64
	Nickname  string
	Content   string
	Status    string
	CreatedAt time.Time
}
type SiteSettings struct {
	SiteTitle       string       `json:"site_title"`
	SiteDescription string       `json:"site_description"`
	SiteKeywords    string       `json:"site_keywords"`
	AboutContent    string       `json:"about_content"`
	FooterText      string       `json:"footer_text"`
	ICPNumber       *string      `json:"icp_number"`
	SocialLinks     []SocialLink `json:"social_links"`
}

type SocialLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon"`
}

type Dashboard struct {
	TotalArticles   int
	DraftArticles   int
	TotalProjects   int
	PendingComments int
	TotalViews      int
	TotalLikes      int
}

type MediaAsset struct {
	ID           int64
	ObjectKey    string
	OriginalName string
	MIMEType     string
	ByteSize     int64
	Width        int
	Height       int
	SHA256       string
	CreatedAt    time.Time
}
