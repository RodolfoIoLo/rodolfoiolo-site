package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/domain"
)

type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}
type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}
type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	Published   string `xml:"pubDate"`
}
type sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}
type sitemapURL struct {
	Location     string `xml:"loc"`
	LastModified string `xml:"lastmod,omitempty"`
}

func (s *ContentService) RSS(ctx context.Context, publicURL string) ([]byte, error) {
	settings, err := s.site.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	page, err := s.articles.List(ctx, domain.ArticleFilter{Page: 1, PageSize: 20})
	if err != nil {
		return nil, err
	}
	channel := rssChannel{Title: settings.SiteTitle, Link: publicURL, Description: settings.SiteDescription, Items: make([]rssItem, 0, len(page.Items))}
	for _, article := range page.Items {
		link := publicURL + "/articles/" + url.PathEscape(article.Slug) + "/"
		published := ""
		if article.PublishedAt != nil {
			published = article.PublishedAt.Format(time.RFC1123Z)
		}
		channel.Items = append(channel.Items, rssItem{Title: article.Title, Link: link, GUID: link, Description: article.Summary, Published: published})
	}
	body, err := xml.MarshalIndent(rss{Version: "2.0", Channel: channel}, "", "  ")
	return append([]byte(xml.Header), body...), err
}
func (s *ContentService) Sitemap(ctx context.Context, publicURL string) ([]byte, error) {
	locations := []sitemapURL{{Location: publicURL + "/"}, {Location: publicURL + "/articles/"}, {Location: publicURL + "/projects/"}, {Location: publicURL + "/archive/"}, {Location: publicURL + "/about/"}}
	for pageNumber := 1; ; pageNumber++ {
		page, err := s.articles.List(ctx, domain.ArticleFilter{Page: pageNumber, PageSize: 100})
		if err != nil {
			return nil, err
		}
		for _, article := range page.Items {
			locations = append(locations, sitemapURL{Location: publicURL + "/articles/" + url.PathEscape(article.Slug) + "/", LastModified: article.UpdatedAt.Format("2006-01-02")})
		}
		if pageNumber >= page.TotalPages {
			break
		}
	}
	projects, err := s.projects.List(ctx, domain.ProjectFilter{Page: 1, PageSize: 100})
	if err != nil {
		return nil, err
	}
	for _, project := range projects.Items {
		locations = append(locations, sitemapURL{Location: publicURL + "/projects/" + url.PathEscape(project.Slug) + "/", LastModified: project.UpdatedAt.Format("2006-01-02")})
	}
	body, err := xml.MarshalIndent(sitemap{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: locations}, "", "  ")
	return append([]byte(xml.Header), body...), err
}
func (s *ContentService) Archive(ctx context.Context) ([]domain.Article, error) {
	items := make([]domain.Article, 0)
	for pageNumber := 1; ; pageNumber++ {
		page, err := s.articles.List(ctx, domain.ArticleFilter{Page: pageNumber, PageSize: 100})
		if err != nil {
			return nil, fmt.Errorf("archive page %d: %w", pageNumber, err)
		}
		items = append(items, page.Items...)
		if pageNumber >= page.TotalPages {
			break
		}
	}
	return items, nil
}
