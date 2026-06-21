// Post editor: a local web UI for creating and editing blog posts.
// Runs a server at http://localhost:3000 — open that in your browser.
//
//	go run ./cmd/postcli
//
// Saving a post automatically rebuilds the homepage writing list and RSS feed.
// No external dependencies (stdlib only).
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	gohtml "html"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode"
)

// --- Constants and types ---

const (
	listMarker  = `<ul class="post-list">` // in index.html: start of the Writing list
	itemMarker  = "<item>"                 // in feed.xml: first RSS item
	baseURL     = "https://jordanjoecooper.com"
	suffixStrip = " - Jordan Joe Cooper"
	port        = "3000"
)

type post struct {
	Slug        string
	File        string
	Title       string
	Date        string
	DisplayDate string
	Description string
	Keywords    string
	Draft       bool
	FullPath    string
}

// --- Finding the repo ---

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "post-template.html")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repo root not found (run from project directory)")
		}
		dir = parent
	}
}

// --- String helpers ---

func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			b.WriteRune(r)
		} else if r == ' ' || r == '\t' {
			b.WriteRune('-')
		}
	}
	return regexp.MustCompile("-+").ReplaceAllString(
		regexp.MustCompile("^-|-$").ReplaceAllString(b.String(), ""), "-")
}

func formatDisplayDate(iso string) string {
	var y, m, d int
	_, _ = fmt.Sscanf(iso, "%d-%d-%d", &y, &m, &d)
	if m < 1 || m > 12 || d < 1 || d > 31 {
		return iso
	}
	return fmt.Sprintf("%02d/%02d/%d", d, m, y)
}

func formatRSSDate(iso string) (string, error) {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return "", err
	}
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700"), nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func escapeXML(s string) string { return escapeHTML(s) }

// jsonStr returns s as the inner content of a JSON string (properly escaped, no outer quotes).
func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}

// buildPostJSONLD returns the Article JSON-LD <script> block for a post.
func buildPostJSONLD(slug, title, description, date string) string {
	postURL := baseURL + "/posts/" + slug + ".html"
	return fmt.Sprintf(`  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "Article",
    "headline": "%s",
    "description": "%s",
    "url": "%s",
    "datePublished": "%s",
    "author": {
      "@type": "Person",
      "@id": "%s/#person",
      "name": "Jordan Joe Cooper",
      "url": "%s"
    },
    "publisher": {
      "@type": "Person",
      "@id": "%s/#person",
      "name": "Jordan Joe Cooper"
    },
    "image": "%s/images/dp.jpg",
    "isPartOf": {
      "@type": "WebSite",
      "@id": "%s/#website"
    }
  }
  </script>`,
		jsonStr(title), jsonStr(description), postURL, date,
		baseURL, baseURL, baseURL, baseURL, baseURL)
}

// fileModDate returns a file's last modification date as YYYY-MM-DD.
func fileModDate(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now().Format("2006-01-02")
	}
	return info.ModTime().Format("2006-01-02")
}

func validateDate(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func normalizeSlug(s, defaultSlug string) string {
	s = strings.TrimSuffix(s, ".html")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else if (r >= 'A' && r <= 'Z') || unicode.IsLetter(r) {
			b.WriteRune(unicode.ToLower(r))
		} else if r == ' ' || r == '_' {
			b.WriteRune('-')
		}
	}
	out := regexp.MustCompile("-+").ReplaceAllString(
		regexp.MustCompile("^-|-$").ReplaceAllString(b.String(), ""), "-")
	if out == "" {
		return defaultSlug
	}
	return out
}

// --- HTML parsing ---

func extractMeta(html, attr, value string) string {
	pat := fmt.Sprintf(`<meta\s+[^>]*%s="%s"[^>]*content="([^"]*)"`, regexp.QuoteMeta(attr), regexp.QuoteMeta(value))
	re := regexp.MustCompile("(?i)" + pat)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return gohtml.UnescapeString(strings.TrimSpace(m[1]))
}

func extractTitle(html string) string {
	s := extractMeta(html, "property", "og:title")
	if s == "" {
		s = extractMeta(html, "name", "twitter:title")
	}
	if s != "" {
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), suffixStrip))
	}
	re := regexp.MustCompile("(?i)<title>([^<]+)</title>")
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(m[1]), suffixStrip))
}

// extractPostBody pulls the editable content out of a post HTML file:
// everything between the closing </div> of .post-meta and </article>.
func extractPostBody(html string) string {
	re := regexp.MustCompile(`(?s)<div class="post-meta">.*?</div>(.*?)</article>`)
	m := re.FindStringSubmatch(html)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	// fallback: everything after h1 inside article
	re2 := regexp.MustCompile(`(?s)<article[^>]*>(?:.*?</h1>)?(.*?)</article>`)
	m2 := re2.FindStringSubmatch(html)
	if len(m2) >= 2 {
		return strings.TrimSpace(m2[1])
	}
	return ""
}

// --- Listing posts ---

func getPosts(root string) ([]post, error) {
	postsDir := filepath.Join(root, "posts")
	entries, err := os.ReadDir(postsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var posts []post
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		fullPath := filepath.Join(postsDir, e.Name())
		body, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}
		h := string(body)
		title := extractTitle(h)
		if title == "" {
			title = strings.TrimSuffix(e.Name(), ".html")
		}
		date := extractMeta(h, "name", "article:published_time")
		if date == "" {
			date = extractMeta(h, "property", "article:published_time")
		}
		description := extractMeta(h, "name", "description")
		if description == "" {
			description = extractMeta(h, "property", "og:description")
		}
		keywords := extractMeta(h, "name", "keywords")
		draft := extractMeta(h, "name", "status") == "draft"
		slug := strings.TrimSuffix(e.Name(), ".html")
		posts = append(posts, post{
			Slug:        slug,
			File:        e.Name(),
			Title:       title,
			Date:        date,
			DisplayDate: formatDisplayDate(date),
			Description: description,
			Keywords:    keywords,
			Draft:       draft,
			FullPath:    fullPath,
		})
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date > posts[j].Date })
	return posts, nil
}

// --- Site rebuild ---

func buildWritingListItems(posts []post) string {
	var b strings.Builder
	for _, p := range posts {
		if p.Slug == "" || p.Title == "" || p.Date == "" || p.Draft {
			continue
		}
		titleForHTML := strings.ReplaceAll(strings.ReplaceAll(p.Title, "&", "&amp;"), "<", "&lt;")
		b.WriteString(fmt.Sprintf("        <li>\n          <h2 class=\"post-title\">\n            <a href=\"posts/%s.html\">%s</a>\n          </h2>\n        </li>\n\n",
			p.Slug, titleForHTML))
	}
	return strings.TrimRight(b.String(), "\n")
}

func rebuildHTMLList(path string, posts []post) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	s := string(data)
	start := strings.Index(s, listMarker)
	if start == -1 {
		return fmt.Errorf("could not find Writing list in %s", filepath.Base(path))
	}
	afterStart := start + len(listMarker)
	end := strings.Index(s[afterStart:], "</ul>")
	if end == -1 {
		return fmt.Errorf("could not find closing </ul> in %s", filepath.Base(path))
	}
	end = afterStart + end
	out := s[:afterStart] + "\n" + buildWritingListItems(posts) + "\n      " + s[end:]
	return os.WriteFile(path, []byte(out), 0644)
}

func buildFeedItems(posts []post) (string, error) {
	var b strings.Builder
	for _, p := range posts {
		if p.Slug == "" || p.Title == "" || p.Date == "" || p.Draft {
			continue
		}
		rssDate, err := formatRSSDate(p.Date)
		if err != nil {
			continue
		}
		url := baseURL + "/posts/" + p.Slug + ".html"
		b.WriteString(fmt.Sprintf("    <item>\n      <title>%s</title>\n      <link>%s</link>\n      <guid>%s</guid>\n      <pubDate>%s</pubDate>\n      <description>%s</description>\n    </item>\n\n",
			escapeXML(p.Title), url, url, rssDate, escapeXML(p.Description)))
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func rebuildFeed(root string, posts []post) error {
	feedPath := filepath.Join(root, "feed.xml")
	data, err := os.ReadFile(feedPath)
	if err != nil {
		return err
	}
	s := string(data)
	items, err := buildFeedItems(posts)
	if err != nil {
		return err
	}
	re := regexp.MustCompile("(?m)^[ \t]*<item>")
	loc := re.FindStringIndex(s)
	chClose := strings.Index(s, "</channel>")
	if chClose == -1 {
		return fmt.Errorf("could not find </channel> in feed.xml")
	}
	var out string
	if loc == nil {
		out = s[:chClose] + items + "\n" + s[chClose:]
	} else {
		out = s[:loc[0]] + items + "\n" + s[chClose:]
	}
	return os.WriteFile(feedPath, []byte(out), 0644)
}

func rebuildSite(root string) error {
	posts, err := getPosts(root)
	if err != nil {
		return err
	}
	if err := rebuildHTMLList(filepath.Join(root, "index.html"), posts); err != nil {
		return err
	}
	if err := rebuildHTMLList(filepath.Join(root, "writing.html"), posts); err != nil {
		return err
	}
	return rebuildFeed(root, posts)
}

// --- Creating and updating post files ---

// newPostHTML creates a new post HTML file from post-template.html.
func newPostHTML(root, title, description, keywords, date, slug, body string, draft bool) (string, error) {
	tplBytes, err := os.ReadFile(filepath.Join(root, "post-template.html"))
	if err != nil {
		return "", fmt.Errorf("post-template.html not found: %w", err)
	}
	displayDate := formatDisplayDate(date)
	s := string(tplBytes)
	s = strings.ReplaceAll(s, "POST_TITLE", escapeHTML(title))
	s = strings.ReplaceAll(s, "POST_DESCRIPTION", escapeHTML(description))
	s = strings.ReplaceAll(s, "POST_KEYWORDS", escapeHTML(keywords))
	s = strings.ReplaceAll(s, "POST_SLUG", slug)
	s = strings.ReplaceAll(s, "YYYY-MM-DD", date)
	s = strings.ReplaceAll(s, "DD/MM/YYYY", displayDate)
	// Replace the template's placeholder body with the real body
	re := regexp.MustCompile(`(?s)(<div class="post-meta">.*?</div>)(.*?)(</article>)`)
	s = re.ReplaceAllString(s, "${1}\n\n    "+strings.TrimSpace(body)+"\n\n  ${3}")
	// Draft status
	if draft {
		s = strings.Replace(s, `<meta name="author" content="Jordan Joe Cooper">`,
			`<meta name="author" content="Jordan Joe Cooper">`+"\n  "+`<meta name="status" content="draft">`, 1)
	}
	// Ensure RSS feed link is present
	if !strings.Contains(s, "application/rss+xml") {
		s = strings.Replace(s,
			`<link rel="stylesheet" href="../styles.css">`,
			`<link rel="alternate" type="application/rss+xml" title="Jordan Joe Cooper" href="../feed.xml">
  <link rel="stylesheet" href="../styles.css">`, 1)
	}
	// Ensure footer is present
	if !strings.Contains(s, "site-footer") {
		s = strings.Replace(s, "</body>",
			`  <footer class="site-footer">
    <nav class="footer-nav">
      <a href="../index.html">Home</a>
      <a href="../index.html#writing">Writing</a>
      <a href="../about.html">About</a>
      <a href="../experience.html">Experience</a>
    </nav>
    <div class="footer-links">
      <a href="https://x.com/jordanjoecooper" target="_blank" rel="noopener">X</a>
      <a href="https://github.com/jordanjoecooper" target="_blank" rel="noopener">GitHub</a>
      <a href="https://linkedin.com/in/jordanjoecooper" target="_blank" rel="noopener">LinkedIn</a>
      <a href="https://reddit.com/user/jordanjoecooper" target="_blank" rel="noopener">Reddit</a>
      <a href="https://margins.app/u/d7d96b0a9a7c43ff9b7586f5eee5f9d8" target="_blank" rel="noopener">Margins</a>
    </div>
    <p class="footer-copy">jordancooper [at] hey dot com &nbsp;·&nbsp; © 2026 Jordan Joe Cooper</p>
  </footer>
</body>`, 1)
	}
	return s, nil
}

// updatePostHTML updates metadata and body in an existing post HTML file.
func updatePostHTML(original, title, description, keywords, date, body string, draft bool) string {
	displayDate := formatDisplayDate(date)
	s := original

	replace := func(pattern, replacement string) {
		re := regexp.MustCompile("(?i)" + pattern)
		s = re.ReplaceAllString(s, replacement)
	}

	replace(`<title>[^<]*</title>`, "<title>"+escapeHTML(title)+" - Jordan Joe Cooper</title>")
	replace(`<meta\s+name="description"\s+content="[^"]*"`, `<meta name="description" content="`+escapeHTML(description)+`"`)
	replace(`<meta\s+name="keywords"\s+content="[^"]*"`, `<meta name="keywords" content="`+escapeHTML(keywords)+`"`)
	replace(`<meta\s+property="og:title"\s+content="[^"]*"`, `<meta property="og:title" content="`+escapeHTML(title)+" - Jordan Joe Cooper"+`"`)
	replace(`<meta\s+property="og:description"\s+content="[^"]*"`, `<meta property="og:description" content="`+escapeHTML(description)+`"`)
	replace(`<meta\s+name="twitter:title"\s+content="[^"]*"`, `<meta name="twitter:title" content="`+escapeHTML(title)+" - Jordan Joe Cooper"+`"`)
	replace(`<meta\s+name="twitter:description"\s+content="[^"]*"`, `<meta name="twitter:description" content="`+escapeHTML(description)+`"`)
	replace(`<meta\s+name="article:published_time"\s+content="[^"]*"`, `<meta name="article:published_time" content="`+date+`"`)
	replace(`<meta\s+property="article:published_time"\s+content="[^"]*"`, `<meta property="article:published_time" content="`+date+`"`)

	// Draft status: add or remove the meta tag
	hasDraft := strings.Contains(s, `<meta name="status" content="draft">`)
	if draft && !hasDraft {
		s = strings.Replace(s, `<meta name="author" content="Jordan Joe Cooper">`,
			`<meta name="author" content="Jordan Joe Cooper">`+"\n  "+`<meta name="status" content="draft">`, 1)
	} else if !draft && hasDraft {
		s = strings.ReplaceAll(s, "\n  <meta name=\"status\" content=\"draft\">", "")
		s = strings.ReplaceAll(s, "<meta name=\"status\" content=\"draft\">\n  ", "")
		s = strings.ReplaceAll(s, "<meta name=\"status\" content=\"draft\">", "")
	}

	// Update article: h1, post-meta time, and body
	re := regexp.MustCompile(`(?s)<article class="post-content">.*?</article>`)
	newArticle := "<article class=\"post-content\">\n    <h1>" + escapeHTML(title) + "</h1>\n    <div class=\"post-meta\">\n      <time datetime=\"" + date + "\">" + displayDate + "</time>\n    </div>\n\n    " + strings.TrimSpace(body) + "\n\n  </article>"
	s = re.ReplaceAllString(s, newArticle)

	return s
}

// --- HTTP handlers ---

type saveRequest struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	Date        string `json:"date"`
	Body        string `json:"body"`
	Draft       bool   `json:"draft"`
	IsNew       bool   `json:"is_new"`
}

type saveResponse struct {
	OK       bool   `json:"ok"`
	Slug     string `json:"slug,omitempty"`
	Redirect string `json:"redirect,omitempty"`
	Error    string `json:"error,omitempty"`
}

func handleList(root string, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		posts, err := getPosts(root)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "list", posts)
	}
}

type editorData struct {
	Title       string
	Description string
	Keywords    string
	Date        string
	DisplayDate string
	Slug        string
	Body        template.HTML
	Draft       bool
	IsNew       bool
}

func handleEditor(root string, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isNew := r.URL.Path == "/new"
		var data editorData

		if isNew {
			data = editorData{
				Date:        time.Now().Format("2006-01-02"),
				DisplayDate: formatDisplayDate(time.Now().Format("2006-01-02")),
				IsNew:       true,
				Body:        template.HTML("<p></p>"),
			}
		} else {
			// /edit/{slug}
			slug := strings.TrimPrefix(r.URL.Path, "/edit/")
			if slug == "" {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			fullPath := filepath.Join(root, "posts", slug+".html")
			content, err := os.ReadFile(fullPath)
			if err != nil {
				http.Error(w, "Post not found: "+slug, 404)
				return
			}
			h := string(content)
			date := extractMeta(h, "name", "article:published_time")
			if date == "" {
				date = extractMeta(h, "property", "article:published_time")
			}
			data = editorData{
				Title:       extractTitle(h),
				Description: extractMeta(h, "name", "description"),
				Keywords:    extractMeta(h, "name", "keywords"),
				Date:        date,
				DisplayDate: formatDisplayDate(date),
				Slug:        slug,
				Body:        template.HTML(extractPostBody(h)),
				Draft:       extractMeta(h, "name", "status") == "draft",
				IsNew:       false,
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "editor", data)
	}
}

func handleSave(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			json.NewEncoder(w).Encode(saveResponse{Error: "POST required"})
			return
		}

		var req saveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(saveResponse{Error: "invalid JSON: " + err.Error()})
			return
		}

		req.Title = strings.TrimSpace(req.Title)
		req.Date = strings.TrimSpace(req.Date)
		req.Slug = strings.TrimSpace(req.Slug)

		if req.Title == "" {
			json.NewEncoder(w).Encode(saveResponse{Error: "title is required"})
			return
		}
		if !validateDate(req.Date) {
			json.NewEncoder(w).Encode(saveResponse{Error: "date must be YYYY-MM-DD"})
			return
		}

		slug := normalizeSlug(req.Slug, slugify(req.Title))
		if slug == "" {
			json.NewEncoder(w).Encode(saveResponse{Error: "could not derive slug from title"})
			return
		}

		outPath := filepath.Join(root, "posts", slug+".html")

		var fileContent string
		if req.IsNew {
			if _, err := os.Stat(outPath); err == nil {
				json.NewEncoder(w).Encode(saveResponse{Error: "post already exists: " + slug})
				return
			}
			var err error
			fileContent, err = newPostHTML(root, req.Title, req.Description, req.Keywords, req.Date, slug, req.Body, req.Draft)
			if err != nil {
				json.NewEncoder(w).Encode(saveResponse{Error: err.Error()})
				return
			}
		} else {
			existing, err := os.ReadFile(outPath)
			if err != nil {
				json.NewEncoder(w).Encode(saveResponse{Error: "post not found: " + slug})
				return
			}
			fileContent = updatePostHTML(string(existing), req.Title, req.Description, req.Keywords, req.Date, req.Body, req.Draft)
		}

		if err := os.WriteFile(outPath, []byte(fileContent), 0644); err != nil {
			json.NewEncoder(w).Encode(saveResponse{Error: "write failed: " + err.Error()})
			return
		}

		if err := rebuildSite(root); err != nil {
			// Post saved but rebuild failed — non-fatal, report it
			json.NewEncoder(w).Encode(saveResponse{OK: true, Slug: slug, Redirect: "/edit/" + slug, Error: "saved but rebuild failed: " + err.Error()})
			return
		}

		json.NewEncoder(w).Encode(saveResponse{OK: true, Slug: slug, Redirect: "/edit/" + slug})
	}
}

// --- Open browser ---

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd /c start"
	default:
		cmd = "xdg-open"
	}
	exec.Command(cmd, url).Start()
}

// --- Templates ---

var listTmplSrc = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Posts — Editor</title>
  <link rel="stylesheet" href="/site/styles.css">
  <style>
    .editor-header { background: var(--surface); border-bottom: 1px solid var(--border); padding: 0.6em 1.5em; font-family: var(--mono); font-size: 0.8em; color: var(--text-secondary); display: flex; align-items: center; justify-content: space-between; }
    .new-btn { background: var(--accent); color: var(--bg); padding: 0.35em 1em; text-decoration: none; border-radius: 3px; font-weight: 600; font-size: 0.85em; }
    .new-btn:hover { background: var(--accent-hover); color: var(--bg); }
  </style>
</head>
<body>
  <div class="editor-header">
    <span>Post Editor — jordanjoecooper.com</span>
    <a class="new-btn" href="/new">+ New Post</a>
  </div>
  <div class="profile-header">
    <img src="/site/images/dp.jpg" alt="Jordan Joe Cooper" class="profile-image">
    <h1>Jordan Joe Cooper</h1>
    <p class="tagline"><span class="tc-1">も</span><span class="tc-2">の</span><span class="tc-3">づ</span><span class="tc-4">く</span><span class="tc-5">り</span></p>
  </div>
  <div class="post-content">
    <h2>Writing</h2>
    {{if .}}
    <ul class="post-list">
      {{range .}}
      <li>
        <h2 class="post-title">
          <a href="/edit/{{.Slug}}">{{.Title}}</a>
          <span style="font-size:0.78em; color:var(--text-secondary); font-family:var(--mono); margin-left:0.5em;">{{.DisplayDate}}{{if .Draft}} · draft{{end}}</span>
        </h2>
      </li>
      {{end}}
    </ul>
    {{else}}
    <p>No posts yet. <a href="/new">Create your first post →</a></p>
    {{end}}
  </div>
</body>
</html>`

var editorTmplSrc = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{if .IsNew}}New Post{{else}}Editing: {{.Title}}{{end}} — Editor</title>
  <link rel="stylesheet" href="/site/styles.css">
  <style>
    #editor-bar {
      position: sticky; top: 0; z-index: 100;
      background: var(--surface); border-bottom: 1px solid var(--border);
      padding: 0.6em 1.2em; display: flex; flex-direction: column; gap: 0.45em;
    }
    .meta-row { display: flex; gap: 0.5em; flex-wrap: wrap; align-items: center; }
    .meta-row input {
      background: var(--bg); border: 1px solid var(--border); color: var(--text);
      padding: 0.28em 0.55em; font-family: var(--mono); font-size: 0.8em; border-radius: 3px;
    }
    .meta-row input:focus { outline: none; border-color: var(--accent); }
    .meta-row input[name=title]       { flex: 2; min-width: 180px; }
    .meta-row input[name=description] { flex: 3; min-width: 180px; }
    .meta-row input[name=date]        { width: 110px; }
    .meta-row input[name=slug]        { width: 150px; }
    .toolbar-row { display: flex; gap: 0.35em; align-items: center; flex-wrap: wrap; }
    .toolbar-row button {
      background: var(--bg); border: 1px solid var(--border); color: var(--text-secondary);
      padding: 0.22em 0.55em; font-family: var(--mono); font-size: 0.78em;
      cursor: pointer; border-radius: 3px; line-height: 1.4;
    }
    .toolbar-row button:hover { color: var(--accent); border-color: var(--accent); }
    #save-btn {
      margin-left: auto; background: var(--accent); color: var(--bg); border: none;
      padding: 0.3em 1em; font-family: var(--mono); font-size: 0.82em;
      cursor: pointer; border-radius: 3px; font-weight: 600;
    }
    #save-btn:hover { background: var(--accent-hover); }
    #save-btn:disabled { opacity: 0.5; cursor: default; }
    #status { font-size: 0.78em; color: var(--text-secondary); margin-left: 0.4em; min-width: 60px; }
    #back-link { font-size: 0.78em; color: var(--text-secondary); text-decoration: none; margin-right: 0.5em; }
    #back-link:hover { color: var(--accent); }
    .post-content[contenteditable] { outline: none; min-height: 12em; }
    .post-content[contenteditable]:focus { outline: none; }
    .post-content[contenteditable] * { cursor: text; }
    #preview-title { pointer-events: none; }
    .post-meta { pointer-events: none; }
  </style>
</head>
<body>
  <div id="editor-bar">
    <div class="meta-row">
      <input type="text" name="title" id="f-title" placeholder="Title" value="{{.Title}}" autocomplete="off">
      <input type="text" name="description" id="f-desc" placeholder="Description (meta)" value="{{.Description}}" autocomplete="off">
      <input type="text" name="date" id="f-date" placeholder="YYYY-MM-DD" value="{{.Date}}" autocomplete="off">
      <input type="text" name="slug" id="f-slug" placeholder="slug" value="{{.Slug}}" autocomplete="off" {{if not .IsNew}}readonly title="Slug cannot be changed after creation"{{end}}>
      <label style="font-family:var(--mono); font-size:0.8em; color:var(--text-secondary); display:flex; align-items:center; gap:0.3em; white-space:nowrap;">
        <input type="checkbox" id="f-draft" {{if .Draft}}checked{{end}}> Draft
      </label>
    </div>
    <div class="toolbar-row">
      <a id="back-link" href="/">← Posts</a>
      <button type="button" onclick="fmt('bold')"><strong>B</strong></button>
      <button type="button" onclick="fmt('italic')"><em>I</em></button>
      <button type="button" onclick="fmt('h2')">H2</button>
      <button type="button" onclick="fmt('h3')">H3</button>
      <button type="button" onclick="fmt('p')">P</button>
      <button type="button" onclick="fmtLink()">Link</button>
      <button type="button" onclick="fmt('blockquote')">Quote</button>
      <button type="button" onclick="fmt('ul')">List</button>
      <button id="save-btn" type="button" onclick="save()">Save</button>
      <span id="status"></span>
    </div>
  </div>

  <div class="profile-header">
    <img src="/site/images/dp.jpg" alt="Jordan Joe Cooper" class="profile-image">
    <h1>Jordan Joe Cooper</h1>
    <p class="tagline"><span class="tc-1">も</span><span class="tc-2">の</span><span class="tc-3">づ</span><span class="tc-4">く</span><span class="tc-5">り</span></p>
    <nav class="main-nav">
      <a href="/">← Posts</a>
    </nav>
  </div>

  <article class="post-content" id="article-preview">
    <h1 id="preview-title">{{.Title}}</h1>
    <div class="post-meta">
      <time id="preview-date" datetime="{{.Date}}">{{.DisplayDate}}</time>
    </div>
    <div id="body" contenteditable="true">{{.Body}}</div>
  </article>

  <script>
    var isNew = {{if .IsNew}}true{{else}}false{{end}};

    // Make Enter create <p> tags instead of <div> or <br>
    document.getElementById('body').addEventListener('keydown', function(e) {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        document.execCommand('insertParagraph');
      }
    });

    document.getElementById('f-title').addEventListener('input', function() {
      document.getElementById('preview-title').textContent = this.value;
      if (isNew) {
        var slug = this.value.toLowerCase()
          .replace(/[^a-z0-9\s-]/g, '').replace(/\s+/g, '-')
          .replace(/-+/g, '-').replace(/^-|-$/g, '');
        document.getElementById('f-slug').value = slug;
      }
    });

    document.getElementById('f-date').addEventListener('input', function() {
      var p = this.value.split('-');
      if (p.length === 3 && p[0].length === 4 && p[1].length === 2 && p[2].length === 2) {
        document.getElementById('preview-date').textContent = p[2]+'/'+p[1]+'/'+p[0];
        document.getElementById('preview-date').setAttribute('datetime', this.value);
      }
    });

    function fmt(cmd) {
      document.getElementById('body').focus();
      if (['h2','h3','p','blockquote'].includes(cmd)) {
        document.execCommand('formatBlock', false, cmd);
      } else if (cmd === 'ul') {
        document.execCommand('insertUnorderedList');
      } else {
        document.execCommand(cmd);
      }
    }

    function fmtLink() {
      var url = prompt('URL (leave empty to remove link):');
      if (url === null) return;
      document.getElementById('body').focus();
      if (url === '') {
        document.execCommand('unlink');
      } else {
        document.execCommand('createLink', false, url);
        var sel = window.getSelection();
        if (sel && sel.anchorNode) {
          var el = sel.anchorNode.nodeType === 3 ? sel.anchorNode.parentElement : sel.anchorNode;
          var a = el.closest ? el.closest('a') : null;
          if (a && !url.startsWith('/') && !url.startsWith('#')) {
            a.target = '_blank'; a.rel = 'noopener';
          }
        }
      }
    }

    function save() {
      var btn = document.getElementById('save-btn');
      var status = document.getElementById('status');
      btn.disabled = true;
      status.textContent = 'Saving...';

      fetch('/save', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
          slug:        document.getElementById('f-slug').value.trim(),
          title:       document.getElementById('f-title').value.trim(),
          description: document.getElementById('f-desc').value.trim(),
          keywords:    '',
          date:        document.getElementById('f-date').value.trim(),
          draft:       document.getElementById('f-draft').checked,
          body:        document.getElementById('body').innerHTML,
          is_new:      isNew
        })
      })
      .then(r => r.json())
      .then(data => {
        btn.disabled = false;
        if (data.ok) {
          status.textContent = 'Saved ✓';
          if (isNew && data.redirect) { window.location.href = data.redirect; return; }
          isNew = false;
          setTimeout(function(){ status.textContent = ''; }, 2500);
        } else {
          status.textContent = 'Error: ' + (data.error || 'unknown');
        }
      })
      .catch(function(err) {
        btn.disabled = false;
        status.textContent = 'Error: ' + err;
      });
    }

    document.addEventListener('keydown', function(e) {
      if ((e.metaKey || e.ctrlKey) && e.key === 's') { e.preventDefault(); save(); }
    });
  </script>
</body>
</html>`

// --- Main ---

func main() {
	root, err := findRoot()
	if err != nil {
		log.Fatal(err)
	}

	// Quick rebuild without starting the server
	if len(os.Args) > 1 && os.Args[1] == "rebuild" {
		if err := rebuildSite(root); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Done — homepage and RSS rebuilt.")
		return
	}

	// Parse templates
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"not": func(v interface{}) bool {
			if v == nil {
				return true
			}
			if b, ok := v.(bool); ok {
				return !b
			}
			return false
		},
	}).Parse(listTmplSrc + editorTmplSrc))

	// Give each template a name by registering them separately
	tmpl = template.Must(template.New("list").Parse(listTmplSrc))
	template.Must(tmpl.New("editor").Parse(editorTmplSrc))

	mux := http.NewServeMux()

	// Serve the site's static assets (CSS, images) at /site/
	mux.Handle("/site/", http.StripPrefix("/site/", http.FileServer(http.Dir(root))))

	// Editor routes
	mux.HandleFunc("/", handleList(root, tmpl))
	mux.HandleFunc("/new", handleEditor(root, tmpl))
	mux.HandleFunc("/edit/", handleEditor(root, tmpl))
	mux.HandleFunc("/save", handleSave(root))

	addr := "http://localhost:" + port
	fmt.Printf("Post editor running at %s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")

	go func() {
		time.Sleep(150 * time.Millisecond)
		openBrowser(addr)
	}()

	log.Fatal(http.ListenAndServe(":"+port, mux))
}
