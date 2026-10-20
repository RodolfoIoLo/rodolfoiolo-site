package service

import (
	"context"
	"strings"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
)

type ContentService struct {
	categories *repository.CategoryRepository
	tags       *repository.TagRepository
	articles   *repository.ArticleRepository
	projects   *repository.ProjectRepository
	site       *repository.SiteRepository
}

func NewContentService(categories *repository.CategoryRepository, tags *repository.TagRepository, articles *repository.ArticleRepository, projects *repository.ProjectRepository, site *repository.SiteRepository) *ContentService {
	return &ContentService{categories: categories, tags: tags, articles: articles, projects: projects, site: site}
}

func (s *ContentService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}
func (s *ContentService) GetCategory(ctx context.Context, slug string) (domain.Category, error) {
	return s.categories.GetBySlug(ctx, slug)
}
func (s *ContentService) SaveCategory(ctx context.Context, currentSlug string, input domain.CategoryInput) (domain.Category, error) {
	name, err := requiredText(input.Name, 1, 100)
	if err != nil {
		return domain.Category{}, err
	}
	input.Name = name
	input.Slug, err = normalizedSlug(input.Slug, input.Name)
	if err != nil {
		return domain.Category{}, err
	}
	if input.ParentID != nil {
		depth, err := s.categories.ParentDepth(ctx, *input.ParentID)
		if err != nil || depth > 1 {
			return domain.Category{}, domain.ErrBadRequest
		}
	}
	if currentSlug == "" {
		return s.categories.Create(ctx, input)
	}
	existing, err := s.categories.GetBySlug(ctx, currentSlug)
	if err != nil {
		return domain.Category{}, err
	}
	if input.ParentID != nil && *input.ParentID == existing.ID {
		return domain.Category{}, domain.ErrBadRequest
	}
	return s.categories.Update(ctx, currentSlug, input)
}
func (s *ContentService) DeleteCategory(ctx context.Context, slug string) error {
	return s.categories.Delete(ctx, slug)
}

func (s *ContentService) ListTags(ctx context.Context) ([]domain.Tag, error) { return s.tags.List(ctx) }
func (s *ContentService) GetTag(ctx context.Context, slug string) (domain.Tag, error) {
	return s.tags.GetBySlug(ctx, slug)
}
func (s *ContentService) SaveTag(ctx context.Context, currentSlug string, input domain.TagInput) (domain.Tag, error) {
	name, err := requiredText(input.Name, 1, 100)
	if err != nil {
		return domain.Tag{}, err
	}
	input.Name = name
	input.Slug, err = normalizedSlug(input.Slug, input.Name)
	if err != nil {
		return domain.Tag{}, err
	}
	if err := validColor(input.Color); err != nil {
		return domain.Tag{}, err
	}
	if currentSlug == "" {
		return s.tags.Create(ctx, input)
	}
	return s.tags.Update(ctx, currentSlug, input)
}
func (s *ContentService) DeleteTag(ctx context.Context, slug string) error {
	return s.tags.Delete(ctx, slug)
}

func (s *ContentService) ListArticles(ctx context.Context, filter domain.ArticleFilter) (domain.Page[domain.Article], error) {
	page, size, err := normalizedPage(&filter.Page, &filter.PageSize)
	if err != nil {
		return domain.Page[domain.Article]{}, err
	}
	filter.Page, filter.PageSize = page, size
	filter.Keyword = strings.TrimSpace(filter.Keyword)
	return s.articles.List(ctx, filter)
}
func (s *ContentService) GetArticle(ctx context.Context, slug string, private bool) (domain.Article, error) {
	return s.articles.Get(ctx, slug, private)
}
func (s *ContentService) CreateArticle(ctx context.Context, userID int64, input domain.ArticleInput) (domain.Article, error) {
	var err error
	input.Title, err = requiredText(input.Title, 1, 300)
	if err != nil {
		return domain.Article{}, err
	}
	input.Content, err = requiredText(input.Content, 1, 200000)
	if err != nil {
		return domain.Article{}, err
	}
	input.Slug, err = normalizedSlug(input.Slug, input.Title)
	if err != nil {
		return domain.Article{}, err
	}
	if strings.TrimSpace(input.Summary) == "" {
		input.Summary = excerpt(input.Content, 200)
	}
	input.Summary, err = requiredText(input.Summary, 1, 500)
	if err != nil {
		return domain.Article{}, err
	}
	if input.Status == "" {
		input.Status = "draft"
	}
	if err := statusAllowed(input.Status, "draft", "published"); err != nil {
		return domain.Article{}, err
	}
	return s.articles.Create(ctx, userID, input)
}
func (s *ContentService) UpdateArticle(ctx context.Context, slug string, patch domain.ArticleInput) (domain.Article, error) {
	current, err := s.articles.Get(ctx, slug, true)
	if err != nil {
		return domain.Article{}, err
	}
	if patch.Title == "" {
		patch.Title = current.Title
	}
	if patch.Content == "" {
		patch.Content = current.Content
	}
	if patch.Summary == "" {
		patch.Summary = current.Summary
	}
	if patch.Slug == "" {
		patch.Slug = current.Slug
	}
	if patch.CoverImage == nil {
		patch.CoverImage = current.CoverImage
	}
	if patch.CategoryID == nil && current.Category != nil {
		id := current.Category.ID
		patch.CategoryID = &id
	}
	if patch.TagIDs == nil {
		patch.TagIDs = make([]int64, 0, len(current.Tags))
		for _, tag := range current.Tags {
			patch.TagIDs = append(patch.TagIDs, tag.ID)
		}
	}
	patch.Status = current.Status
	patch.Title, err = requiredText(patch.Title, 1, 300)
	if err != nil {
		return domain.Article{}, err
	}
	patch.Content, err = requiredText(patch.Content, 1, 200000)
	if err != nil {
		return domain.Article{}, err
	}
	patch.Summary, err = requiredText(patch.Summary, 1, 500)
	if err != nil {
		return domain.Article{}, err
	}
	patch.Slug, err = normalizedSlug(patch.Slug, patch.Title)
	if err != nil {
		return domain.Article{}, err
	}
	return s.articles.Update(ctx, slug, patch)
}
func (s *ContentService) PublishArticle(ctx context.Context, slug string) (domain.Article, error) {
	current, err := s.articles.Get(ctx, slug, true)
	if err != nil {
		return domain.Article{}, err
	}
	if current.Status == "archived" {
		return domain.Article{}, domain.ErrConflict
	}
	return s.articles.SetStatus(ctx, slug, "published")
}
func (s *ContentService) UnpublishArticle(ctx context.Context, slug string) (domain.Article, error) {
	current, err := s.articles.Get(ctx, slug, true)
	if err != nil {
		return domain.Article{}, err
	}
	if current.Status != "published" {
		return domain.Article{}, domain.ErrConflict
	}
	return s.articles.SetStatus(ctx, slug, "draft")
}
func (s *ContentService) SetArticleTop(ctx context.Context, slug string, top bool) (domain.Article, error) {
	return s.articles.SetTop(ctx, slug, top)
}
func (s *ContentService) DeleteArticle(ctx context.Context, slug string) error {
	return s.articles.Delete(ctx, slug)
}

func (s *ContentService) ListProjects(ctx context.Context, filter domain.ProjectFilter) (domain.Page[domain.Project], error) {
	page, size, err := normalizedPage(&filter.Page, &filter.PageSize)
	if err != nil {
		return domain.Page[domain.Project]{}, err
	}
	filter.Page, filter.PageSize = page, size
	return s.projects.List(ctx, filter)
}
func (s *ContentService) GetProject(ctx context.Context, slug string, private bool) (domain.Project, error) {
	return s.projects.Get(ctx, slug, private)
}
func (s *ContentService) SaveProject(ctx context.Context, currentSlug string, input domain.ProjectInput) (domain.Project, error) {
	if currentSlug != "" {
		current, err := s.projects.Get(ctx, currentSlug, true)
		if err != nil {
			return domain.Project{}, err
		}
		if input.Name == "" {
			input.Name = current.Name
		}
		if input.Slug == "" {
			input.Slug = current.Slug
		}
		if input.Summary == "" {
			input.Summary = current.Summary
		}
		if input.Content == "" {
			input.Content = current.Content
		}
		if input.CoverImage == nil {
			input.CoverImage = current.CoverImage
		}
		if input.Technologies == nil {
			input.Technologies = current.Technologies
		}
		if input.RepositoryURL == nil {
			input.RepositoryURL = current.RepositoryURL
		}
		if input.DemoURL == nil {
			input.DemoURL = current.DemoURL
		}
		if input.Status == "" {
			input.Status = current.Status
		}
		if input.StartedAt == nil {
			input.StartedAt = current.StartedAt
		}
		if input.CompletedAt == nil {
			input.CompletedAt = current.CompletedAt
		}
	}
	var err error
	input.Name, err = requiredText(input.Name, 1, 200)
	if err != nil {
		return domain.Project{}, err
	}
	input.Slug, err = normalizedSlug(input.Slug, input.Name)
	if err != nil {
		return domain.Project{}, err
	}
	input.Summary, err = requiredText(input.Summary, 1, 500)
	if err != nil {
		return domain.Project{}, err
	}
	input.Content, err = requiredText(input.Content, 1, 200000)
	if err != nil {
		return domain.Project{}, err
	}
	input.Technologies, err = normalizedTechnologies(input.Technologies)
	if err != nil {
		return domain.Project{}, err
	}
	input.RepositoryURL, err = optionalHTTPURL(input.RepositoryURL)
	if err != nil {
		return domain.Project{}, err
	}
	input.DemoURL, err = optionalHTTPURL(input.DemoURL)
	if err != nil {
		return domain.Project{}, err
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if err := statusAllowed(input.Status, "active", "archived"); err != nil {
		return domain.Project{}, err
	}
	if input.StartedAt != nil && input.CompletedAt != nil && input.CompletedAt.Before(*input.StartedAt) {
		return domain.Project{}, domain.ErrBadRequest
	}
	if currentSlug == "" {
		return s.projects.Create(ctx, input)
	}
	return s.projects.Update(ctx, currentSlug, input)
}
func (s *ContentService) DeleteProject(ctx context.Context, slug string) error {
	return s.projects.Delete(ctx, slug)
}

func (s *ContentService) GetSettings(ctx context.Context) (domain.SiteSettings, error) {
	return s.site.GetSettings(ctx)
}
func (s *ContentService) UpdateSettings(ctx context.Context, userID int64, settings domain.SiteSettings) (domain.SiteSettings, error) {
	var err error
	settings.SiteTitle, err = requiredText(settings.SiteTitle, 1, 100)
	if err != nil {
		return domain.SiteSettings{}, err
	}
	settings.SiteDescription, err = requiredText(settings.SiteDescription, 1, 300)
	if err != nil {
		return domain.SiteSettings{}, err
	}
	settings.AboutContent, err = requiredText(settings.AboutContent, 1, 100000)
	if err != nil {
		return domain.SiteSettings{}, err
	}
	for _, link := range settings.SocialLinks {
		if _, err := requiredText(link.Name, 1, 50); err != nil {
			return domain.SiteSettings{}, err
		}
		value := link.URL
		if _, err := optionalHTTPURL(&value); err != nil {
			return domain.SiteSettings{}, err
		}
	}
	if err := s.site.UpdateSettings(ctx, settings, userID); err != nil {
		return domain.SiteSettings{}, err
	}
	return settings, nil
}
func (s *ContentService) PublicStats(ctx context.Context) (int, int, int, int, error) {
	return s.site.PublicStats(ctx)
}
func (s *ContentService) Dashboard(ctx context.Context) (domain.Dashboard, error) {
	return s.site.Dashboard(ctx)
}

func excerpt(value string, maximum int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= maximum {
		return value
	}
	return string(runes[:maximum]) + "..."
}
