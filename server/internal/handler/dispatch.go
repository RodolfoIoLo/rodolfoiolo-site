package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"sort"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/generated/oapi"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/requestctx"
)

func (h *Handler) dispatch(ctx context.Context, operation string, rawRequest any) (*OperationResponse, error) {
	switch operation {
	case "ListCategories":
		items, err := h.content.ListCategories(ctx)
		return result(itemsToAny(items, categoryData), err, http.StatusOK)
	case "GetCategory":
		request := rawRequest.(oapi.GetCategoryRequestObject)
		item, err := h.content.GetCategory(ctx, request.Slug)
		return result(categoryData(item), err, http.StatusOK)
	case "CreateCategory":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.CreateCategoryRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.SaveCategory(ctx, "", domain.CategoryInput{Name: request.Body.Name, Slug: stringValue(request.Body.Slug), Description: request.Body.Description, Icon: request.Body.Icon, ParentID: request.Body.ParentID, SortOrder: intValue(request.Body.SortOrder)})
		return result(categoryData(item), err, http.StatusCreated)
	case "UpdateCategory":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.UpdateCategoryRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.SaveCategory(ctx, request.Slug, domain.CategoryInput{Name: request.Body.Name, Slug: stringValue(request.Body.Slug), Description: request.Body.Description, Icon: request.Body.Icon, ParentID: request.Body.ParentID, SortOrder: intValue(request.Body.SortOrder)})
		return result(categoryData(item), err, http.StatusOK)
	case "DeleteCategory":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.DeleteCategoryRequestObject)
		return emptyResult(h.content.DeleteCategory(ctx, request.Slug))
	case "ListTags":
		items, err := h.content.ListTags(ctx)
		return result(itemsToAny(items, tagData), err, http.StatusOK)
	case "GetTag":
		request := rawRequest.(oapi.GetTagRequestObject)
		item, err := h.content.GetTag(ctx, request.Slug)
		return result(tagData(item), err, http.StatusOK)
	case "CreateTag":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.CreateTagRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.SaveTag(ctx, "", domain.TagInput{Name: request.Body.Name, Slug: stringValue(request.Body.Slug), Color: request.Body.Color})
		return result(tagData(item), err, http.StatusCreated)
	case "UpdateTag":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.UpdateTagRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.SaveTag(ctx, request.Slug, domain.TagInput{Name: request.Body.Name, Slug: stringValue(request.Body.Slug), Color: request.Body.Color})
		return result(tagData(item), err, http.StatusOK)
	case "DeleteTag":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.DeleteTagRequestObject)
		return emptyResult(h.content.DeleteTag(ctx, request.Slug))
	case "ListArticles":
		request := rawRequest.(oapi.ListArticlesRequestObject)
		page, err := h.content.ListArticles(ctx, domain.ArticleFilter{Page: intPointerValue(request.Params.Page), PageSize: intPointerValue(request.Params.PageSize), Category: stringValue(request.Params.Category), Tag: stringValue(request.Params.Tag), Keyword: stringValue(request.Params.Keyword), Sort: enumString(request.Params.Sort)})
		return result(articlePageData(page, false), err, http.StatusOK)
	case "ListAdminArticles":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.ListAdminArticlesRequestObject)
		page, err := h.content.ListArticles(ctx, domain.ArticleFilter{Page: intPointerValue(request.Params.Page), PageSize: intPointerValue(request.Params.PageSize), Status: enumString(request.Params.Status), Keyword: stringValue(request.Params.Keyword), IncludePrivate: true})
		return result(articlePageData(page, false), err, http.StatusOK)
	case "GetArticle":
		request := rawRequest.(oapi.GetArticleRequestObject)
		item, err := h.content.GetArticle(ctx, request.Slug, false)
		return result(articleData(item, true), err, http.StatusOK)
	case "CreateArticle":
		identity, response := editorIdentity(ctx)
		if response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.CreateArticleRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		input := articleCreateInput(*request.Body)
		item, err := h.content.CreateArticle(ctx, identity.UserID, input)
		return result(articleData(item, true), err, http.StatusCreated)
	case "UpdateArticle":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.UpdateArticleRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.UpdateArticle(ctx, request.Slug, articleUpdateInput(*request.Body))
		return result(articleData(item, true), err, http.StatusOK)
	case "DeleteArticle":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.DeleteArticleRequestObject)
		return emptyResult(h.content.DeleteArticle(ctx, request.Slug))
	case "PublishArticle":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.PublishArticleRequestObject)
		item, err := h.content.PublishArticle(ctx, request.Slug)
		return result(articleData(item, true), err, http.StatusOK)
	case "UnpublishArticle":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.UnpublishArticleRequestObject)
		item, err := h.content.UnpublishArticle(ctx, request.Slug)
		return result(articleData(item, true), err, http.StatusOK)
	case "SetArticleTop":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.SetArticleTopRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.SetArticleTop(ctx, request.Slug, request.Body.IsTop)
		return result(articleData(item, true), err, http.StatusOK)
	case "ListArticleArchive":
		items, err := h.content.Archive(ctx)
		return result(archiveData(items), err, http.StatusOK)
	case "ListProjects":
		request := rawRequest.(oapi.ListProjectsRequestObject)
		page, err := h.content.ListProjects(ctx, domain.ProjectFilter{Page: intPointerValue(request.Params.Page), PageSize: intPointerValue(request.Params.PageSize), Featured: request.Params.Featured, Technology: stringValue(request.Params.Technology)})
		return result(projectPageData(page, false), err, http.StatusOK)
	case "ListAdminProjects":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.ListAdminProjectsRequestObject)
		page, err := h.content.ListProjects(ctx, domain.ProjectFilter{Page: intPointerValue(request.Params.Page), PageSize: intPointerValue(request.Params.PageSize), Status: enumString(request.Params.Status), IncludePrivate: true})
		return result(projectPageData(page, false), err, http.StatusOK)
	case "GetProject":
		request := rawRequest.(oapi.GetProjectRequestObject)
		item, err := h.content.GetProject(ctx, request.Slug, false)
		return result(projectData(item, true), err, http.StatusOK)
	case "CreateProject":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.CreateProjectRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.SaveProject(ctx, "", projectCreateInput(*request.Body))
		return result(projectData(item, true), err, http.StatusCreated)
	case "UpdateProject":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.UpdateProjectRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		item, err := h.content.SaveProject(ctx, request.Slug, projectUpdateInput(*request.Body))
		return result(projectData(item, true), err, http.StatusOK)
	case "DeleteProject":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.DeleteProjectRequestObject)
		return emptyResult(h.content.DeleteProject(ctx, request.Slug))
	case "GetSiteSettings":
		settings, err := h.content.GetSettings(ctx)
		return result(settings, err, http.StatusOK)
	case "UpdateSiteSettings":
		identity, response := adminIdentity(ctx)
		if response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.UpdateSiteSettingsRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		bytes, err := json.Marshal(request.Body)
		if err != nil {
			return nil, err
		}
		var settings domain.SiteSettings
		if err := json.Unmarshal(bytes, &settings); err != nil {
			return nil, err
		}
		settings, err = h.content.UpdateSettings(ctx, identity.UserID, settings)
		return result(settings, err, http.StatusOK)
	case "GetSiteStats":
		articles, projects, categories, tags, err := h.content.PublicStats(ctx)
		return result(map[string]int{"total_articles": articles, "total_projects": projects, "total_categories": categories, "total_tags": tags}, err, http.StatusOK)
	case "GetAdminDashboard":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		dashboard, err := h.content.Dashboard(ctx)
		return result(map[string]int{"total_articles": dashboard.TotalArticles, "draft_articles": dashboard.DraftArticles, "total_projects": dashboard.TotalProjects, "pending_comments": dashboard.PendingComments, "total_views": dashboard.TotalViews, "total_likes": dashboard.TotalLikes}, err, http.StatusOK)
	case "ListArticleComments":
		request := rawRequest.(oapi.ListArticleCommentsRequestObject)
		page, err := h.interactions.List(ctx, request.Slug, intPointerValue(request.Params.Page), intPointerValue(request.Params.PageSize))
		return result(commentPageData(page), err, http.StatusOK)
	case "CreateComment":
		request := rawRequest.(oapi.CreateCommentRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		metadata := requestctx.MetadataFrom(ctx)
		email := ""
		if request.Body.Email != nil {
			email = string(*request.Body.Email)
		}
		honeypot := stringValue(request.Body.WebsiteURL)
		item, err := h.interactions.Create(ctx, domain.CommentInput{ArticleSlug: request.Slug, ParentID: request.Body.ParentID, Nickname: request.Body.Nickname, Content: request.Body.Content, VisitorID: metadata.VisitorID, IPHash: metadata.IPHash, UserAgent: metadata.UserAgent}, email, honeypot)
		return result(commentData(item), err, http.StatusCreated)
	case "ListAdminComments":
		if response := requireEditor(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.ListAdminCommentsRequestObject)
		page, err := h.interactions.ListAdmin(ctx, intPointerValue(request.Params.Page), intPointerValue(request.Params.PageSize), enumString(request.Params.Status))
		return result(commentPageData(page), err, http.StatusOK)
	case "ApproveComment", "RejectComment":
		identity, response := editorIdentity(ctx)
		if response != nil {
			return response, nil
		}
		id := int64(0)
		status := "approved"
		if operation == "ApproveComment" {
			id = rawRequest.(oapi.ApproveCommentRequestObject).ID
		} else {
			id = rawRequest.(oapi.RejectCommentRequestObject).ID
			status = "rejected"
		}
		item, err := h.interactions.Review(ctx, id, identity.UserID, status)
		return result(commentData(item), err, http.StatusOK)
	case "DeleteComment":
		if response := requireAdmin(ctx); response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.DeleteCommentRequestObject)
		return emptyResult(h.interactions.Delete(ctx, request.ID))
	case "SetArticleLike":
		request := rawRequest.(oapi.SetArticleLikeRequestObject)
		if request.Body == nil {
			return Failure(400, 40000, "request body is required"), nil
		}
		metadata := requestctx.MetadataFrom(ctx)
		liked, count, err := h.interactions.SetLike(ctx, request.Slug, metadata.VisitorID, metadata.IPHash, string(request.Body.Action) == "like")
		return result(map[string]any{"liked": liked, "like_count": count}, err, http.StatusOK)
	case "GetArticleLikeStatus":
		request := rawRequest.(oapi.GetArticleLikeStatusRequestObject)
		metadata := requestctx.MetadataFrom(ctx)
		liked, count, err := h.interactions.LikeStatus(ctx, request.Slug, metadata.VisitorID)
		return result(map[string]any{"liked": liked, "like_count": count}, err, http.StatusOK)
	case "UploadImage":
		identity, response := editorIdentity(ctx)
		if response != nil {
			return response, nil
		}
		request := rawRequest.(oapi.UploadImageRequestObject)
		filename, data, err := readUpload(request.Body)
		if err != nil {
			return operationError(err), nil
		}
		uploaded, err := h.media.Upload(ctx, identity.UserID, filename, data)
		return result(map[string]any{"filename": uploaded.Asset.OriginalName, "url": uploaded.URL, "size": uploaded.Asset.ByteSize, "width": uploaded.Asset.Width, "height": uploaded.Asset.Height}, err, http.StatusOK)
	case "GetRss":
		body, err := h.content.RSS(ctx, h.publicURL)
		if err != nil {
			return nil, err
		}
		return &OperationResponse{Status: 200, ContentType: "application/rss+xml; charset=utf-8", Body: body}, nil
	case "GetSitemap":
		body, err := h.content.Sitemap(ctx, h.publicURL)
		if err != nil {
			return nil, err
		}
		return &OperationResponse{Status: 200, ContentType: "application/xml; charset=utf-8", Body: body}, nil
	case "GetHealth":
		status := "ok"
		return Success(200, map[string]any{"status": status, "timestamp": h.now().UTC()}), nil
	case "GetReady":
		pingContext, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := h.database.Ping(pingContext); err != nil {
			return JSON(503, map[string]string{"status": "not_ready", "database": "disconnected"}), nil
		}
		return JSON(200, map[string]string{"status": "ready", "database": "connected"}), nil
	default:
		return nil, fmt.Errorf("operation %q is not implemented", operation)
	}
}

func result(data any, err error, status int) (*OperationResponse, error) {
	if err != nil {
		return operationError(err), nil
	}
	return Success(status, data), nil
}
func emptyResult(err error) (*OperationResponse, error) {
	if err != nil {
		return operationError(err), nil
	}
	return Success(200, nil), nil
}
func operationError(err error) *OperationResponse {
	switch {
	case errors.Is(err, domain.ErrBadRequest):
		return Failure(400, 40000, "request is invalid")
	case errors.Is(err, domain.ErrUnauthorized):
		return Failure(401, 40100, "authentication is required")
	case errors.Is(err, domain.ErrForbidden):
		return Failure(403, 40300, "permission is denied")
	case errors.Is(err, domain.ErrNotFound):
		return Failure(404, 40400, "resource was not found")
	case errors.Is(err, domain.ErrConflict):
		return Failure(409, 40900, "resource state conflicts with the request")
	default:
		slog.Error("operation_failed", "error", err)
		return Failure(500, 50000, "internal server error")
	}
}
func requireAdmin(ctx context.Context) *OperationResponse {
	_, response := adminIdentity(ctx)
	return response
}
func requireEditor(ctx context.Context) *OperationResponse {
	_, response := editorIdentity(ctx)
	return response
}
func adminIdentity(ctx context.Context) (requestctx.Identity, *OperationResponse) {
	identity, ok := requestctx.IdentityFrom(ctx)
	if !ok {
		return requestctx.Identity{}, Failure(401, 40100, "authentication is required")
	}
	if identity.Role != "admin" {
		return requestctx.Identity{}, Failure(403, 40300, "administrator permission is required")
	}
	return identity, nil
}
func editorIdentity(ctx context.Context) (requestctx.Identity, *OperationResponse) {
	identity, ok := requestctx.IdentityFrom(ctx)
	if !ok {
		return requestctx.Identity{}, Failure(401, 40100, "authentication is required")
	}
	if identity.Role != "admin" && identity.Role != "editor" {
		return requestctx.Identity{}, Failure(403, 40300, "editor permission is required")
	}
	return identity, nil
}

func readUpload(reader *multipart.Reader) (string, []byte, error) {
	if reader == nil {
		return "", nil, domain.ErrBadRequest
	}
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", nil, domain.ErrBadRequest
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}
		data, err := io.ReadAll(io.LimitReader(part, (6<<20)+1))
		_ = part.Close()
		if err != nil || len(data) > (6<<20) {
			return "", nil, domain.ErrBadRequest
		}
		return part.FileName(), data, nil
	}
	return "", nil, domain.ErrBadRequest
}
func stringValue[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}
func enumString[T ~string](value *T) string { return stringValue(value) }
func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
func intPointerValue[T ~int](value *T) int {
	if value == nil {
		return 0
	}
	return int(*value)
}

func articleCreateInput(body oapi.ArticleCreateRequest) domain.ArticleInput {
	input := domain.ArticleInput{Slug: stringValue(body.Slug), Title: body.Title, Content: body.Content, CoverImage: body.CoverImage, CategoryID: body.CategoryID, IsTop: body.IsTop != nil && *body.IsTop}
	if body.Summary != nil {
		input.Summary = *body.Summary
	}
	if body.TagIds != nil {
		input.TagIDs = *body.TagIds
	}
	if body.Status != nil {
		input.Status = string(*body.Status)
	}
	return input
}
func articleUpdateInput(body oapi.ArticleUpdateRequest) domain.ArticleInput {
	input := domain.ArticleInput{Slug: stringValue(body.Slug), CoverImage: body.CoverImage, CategoryID: body.CategoryID}
	if body.Title != nil {
		input.Title = *body.Title
	}
	if body.Summary != nil {
		input.Summary = *body.Summary
	}
	if body.Content != nil {
		input.Content = *body.Content
	}
	if body.TagIds != nil {
		input.TagIDs = *body.TagIds
	}
	if body.IsTop != nil {
		input.IsTop = *body.IsTop
	}
	return input
}
func projectCreateInput(body oapi.ProjectCreateRequest) domain.ProjectInput {
	input := domain.ProjectInput{Name: body.Name, Slug: stringValue(body.Slug), CoverImage: body.CoverImage, RepositoryURL: body.RepositoryURL, DemoURL: body.DemoURL, Featured: body.Featured != nil && *body.Featured, SortOrder: intValue(body.SortOrder)}
	if body.Summary != nil {
		input.Summary = *body.Summary
	}
	if body.Content != nil {
		input.Content = *body.Content
	}
	if body.Technologies != nil {
		input.Technologies = *body.Technologies
	}
	if body.Status != nil {
		input.Status = string(*body.Status)
	}
	if body.StartedAt != nil {
		value := body.StartedAt.Time
		input.StartedAt = &value
	}
	if body.CompletedAt != nil {
		value := body.CompletedAt.Time
		input.CompletedAt = &value
	}
	return input
}
func projectUpdateInput(body oapi.ProjectUpdateRequest) domain.ProjectInput {
	input := domain.ProjectInput{Slug: stringValue(body.Slug), CoverImage: body.CoverImage, RepositoryURL: body.RepositoryURL, DemoURL: body.DemoURL}
	if body.Name != nil {
		input.Name = *body.Name
	}
	if body.Summary != nil {
		input.Summary = *body.Summary
	}
	if body.Content != nil {
		input.Content = *body.Content
	}
	if body.Technologies != nil {
		input.Technologies = *body.Technologies
	}
	if body.Featured != nil {
		input.Featured = *body.Featured
	}
	if body.SortOrder != nil {
		input.SortOrder = *body.SortOrder
	}
	if body.Status != nil {
		input.Status = string(*body.Status)
	}
	if body.StartedAt != nil {
		value := body.StartedAt.Time
		input.StartedAt = &value
	}
	if body.CompletedAt != nil {
		value := body.CompletedAt.Time
		input.CompletedAt = &value
	}
	return input
}

func categoryData(item domain.Category) map[string]any {
	return map[string]any{"id": item.ID, "name": item.Name, "slug": item.Slug, "description": item.Description, "icon": item.Icon, "parent_id": item.ParentID, "sort_order": item.SortOrder, "article_count": item.ArticleCount, "created_at": item.CreatedAt}
}
func tagData(item domain.Tag) map[string]any {
	return map[string]any{"id": item.ID, "name": item.Name, "slug": item.Slug, "color": item.Color, "article_count": item.ArticleCount, "created_at": item.CreatedAt}
}
func articleData(item domain.Article, detail bool) map[string]any {
	tags := itemsToAny(item.Tags, tagData)
	result := map[string]any{"id": item.ID, "slug": item.Slug, "title": item.Title, "summary": item.Summary, "cover_image": item.CoverImage, "category": nil, "tags": tags, "author": map[string]any{"id": item.Author.ID, "username": item.Author.Username, "nickname": item.Author.Nickname, "avatar_url": item.Author.AvatarURL, "bio": item.Author.Bio, "role": item.Author.Role, "created_at": item.Author.CreatedAt}, "status": item.Status, "is_top": item.IsTop, "view_count": item.ViewCount, "like_count": item.LikeCount, "comment_count": item.CommentCount, "published_at": item.PublishedAt, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
	if item.Category != nil {
		result["category"] = categoryData(*item.Category)
	}
	if detail {
		result["content"] = item.Content
	}
	return result
}
func projectData(item domain.Project, detail bool) map[string]any {
	result := map[string]any{"id": item.ID, "slug": item.Slug, "name": item.Name, "summary": item.Summary, "cover_image": item.CoverImage, "technologies": item.Technologies, "repository_url": item.RepositoryURL, "demo_url": item.DemoURL, "featured": item.Featured, "sort_order": item.SortOrder, "status": item.Status, "started_at": dateValue(item.StartedAt), "completed_at": dateValue(item.CompletedAt), "created_at": item.CreatedAt, "updated_at": item.UpdatedAt}
	if detail {
		result["content"] = item.Content
	}
	return result
}
func commentData(item domain.Comment) map[string]any {
	return map[string]any{"id": item.ID, "article_id": item.ArticleID, "parent_id": item.ParentID, "nickname": item.Nickname, "content": item.Content, "status": item.Status, "created_at": item.CreatedAt, "website": nil}
}
func articlePageData(page domain.Page[domain.Article], detail bool) map[string]any {
	items := make([]any, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, articleData(item, detail))
	}
	return pageData(items, page.Page, page.PageSize, page.Total, page.TotalPages)
}
func projectPageData(page domain.Page[domain.Project], detail bool) map[string]any {
	items := make([]any, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, projectData(item, detail))
	}
	return pageData(items, page.Page, page.PageSize, page.Total, page.TotalPages)
}
func commentPageData(page domain.Page[domain.Comment]) map[string]any {
	items := make([]any, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, commentData(item))
	}
	return pageData(items, page.Page, page.PageSize, page.Total, page.TotalPages)
}
func pageData(items []any, page, pageSize, total, totalPages int) map[string]any {
	return map[string]any{"items": items, "pagination": map[string]int{"page": page, "page_size": pageSize, "total": total, "total_pages": totalPages}}
}
func itemsToAny[T any](items []T, convert func(T) map[string]any) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		result = append(result, convert(item))
	}
	return result
}
func dateValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format("2006-01-02")
}
func archiveData(items []domain.Article) []any {
	sort.SliceStable(items, func(i, j int) bool {
		left, right := items[i].PublishedAt, items[j].PublishedAt
		if left == nil {
			return false
		}
		if right == nil {
			return true
		}
		return left.After(*right)
	})
	years := make([]any, 0)
	yearIndex := -1
	monthIndex := -1
	currentYear, currentMonth := -1, -1
	var months []any
	var articles []any
	flushMonth := func() {
		if monthIndex >= 0 {
			months = append(months, map[string]any{"month": currentMonth, "articles": articles})
		}
	}
	flushYear := func() {
		if yearIndex >= 0 {
			flushMonth()
			years = append(years, map[string]any{"year": currentYear, "months": months})
		}
	}
	for _, item := range items {
		if item.PublishedAt == nil {
			continue
		}
		year, month, _ := item.PublishedAt.Date()
		if year != currentYear {
			flushYear()
			yearIndex++
			currentYear = year
			currentMonth = -1
			monthIndex = -1
			months = []any{}
		}
		if int(month) != currentMonth {
			flushMonth()
			monthIndex++
			currentMonth = int(month)
			articles = []any{}
		}
		articles = append(articles, map[string]any{"slug": item.Slug, "title": item.Title, "published_at": item.PublishedAt})
	}
	flushYear()
	return years
}
