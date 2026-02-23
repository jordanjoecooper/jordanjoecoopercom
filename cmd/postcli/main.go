// Post CLI: create and edit posts, and keep the homepage + RSS feed in sync.
//
// No external dependencies (stdlib only). Run from the repo root:
//
//	go run ./cmd/postcli              # interactive menu
//	go run ./cmd/postcli new          # create a new post
//	go run ./cmd/postcli edit [slug]  # edit a post (opens $EDITOR)
//	go run ./cmd/postcli list         # list all posts
//	go run ./cmd/postcli update-links posts/your-post.html
//
// Or build once: go build -o postcli ./cmd/postcli
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

// --- Constants and types ---

// HTML/XML markers we search for when updating the site.
const (
	listMarker  = "<ul class=\"post-list\">" // in index.html: start of the Writing list
	itemMarker  = "<item>"                   // in feed.xml: first RSS item
	baseURL     = "https://jordanjoecooper.com"
	suffixStrip = " - Jordan Joe Cooper"     // stripped from titles in meta
)

var months = []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

// post is a single blog post we know about (from the posts/ directory).
type post struct {
	Slug     string // URL-safe name, no .html
	File     string // filename, e.g. my-post.html
	Title    string // from og:title or <title>
	Date     string // YYYY-MM-DD from article:published_time
	FullPath string // absolute path to the .html file
}

// --- Finding the repo ---

// findRoot walks up from the current directory until it finds both
// post-template.html and index.html. Returns the repo root path or an error.
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

// --- String helpers (dates, slugs, escaping) ---

// slugify turns a title into a URL-safe slug: lowercase, spaces to hyphens,
// only letters, numbers, and hyphens. Used as the default filename for a post.
func slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			b.WriteRune(r)
		} else if r == ' ' || r == '\t' {
			b.WriteRune('-')
		}
	}
	// Collapse multiple hyphens and trim leading/trailing ones.
	return regexp.MustCompile("-+").ReplaceAllString(regexp.MustCompile("^-|-$").ReplaceAllString(b.String(), ""), "-")
}

// formatDisplayDate converts YYYY-MM-DD to "Month Day, Year" (e.g. "February 23, 2026").
func formatDisplayDate(iso string) string {
	var y, m, d int
	_, _ = fmt.Sscanf(iso, "%d-%d-%d", &y, &m, &d)
	if m < 1 || m > 12 {
		return iso
	}
	return fmt.Sprintf("%s %d, %d", months[m-1], d, y)
}

// formatRSSDate converts YYYY-MM-DD to RSS pubDate format (RFC-style, UTC).
func formatRSSDate(iso string) (string, error) {
	t, err := time.Parse("2006-01-02", iso)
	if err != nil {
		return "", err
	}
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700"), nil
}

// escapeHTML escapes &, <, >, " for safe use inside HTML attributes/text.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// escapeXML does the same for XML (e.g. RSS description and title).
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// --- Reading meta from post HTML ---

// extractMeta finds a <meta ... attr="value" ... content="VALUE"> and returns VALUE.
// attr/value are e.g. "property"/"og:title" or "name"/"description".
func extractMeta(html, attr, value string) string {
	pat := fmt.Sprintf(`<meta\s+[^>]*%s="%s"[^>]*content="([^"]*)"`, regexp.QuoteMeta(attr), regexp.QuoteMeta(value))
	re := regexp.MustCompile("(?i)" + pat)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// extractTitle returns the post title from HTML: og:title or twitter:title or <title>,
// with the " - Jordan Joe Cooper" suffix removed.
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

// --- Listing and prompting ---

// getPosts reads the posts/ directory and returns one post per .html file,
// with title and date parsed from meta. Sorted by date newest first.
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
		html := string(body)
		title := extractTitle(html)
		if title == "" {
			title = strings.TrimSuffix(e.Name(), ".html")
		}
		date := extractMeta(html, "name", "article:published_time")
		if date == "" {
			date = extractMeta(html, "property", "article:published_time")
		}
		slug := strings.TrimSuffix(e.Name(), ".html")
		posts = append(posts, post{Slug: slug, File: e.Name(), Title: title, Date: date, FullPath: fullPath})
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].Date > posts[j].Date })
	return posts, nil
}

// readLine prints a prompt (and optional default in parentheses), reads one line
// from stdin, and returns it trimmed. If the line is empty, returns defaultVal.
func readLine(prompt, defaultVal string) (string, error) {
	fmt.Fprint(os.Stdout, prompt)
	if defaultVal != "" {
		fmt.Fprintf(os.Stdout, " (%s)", defaultVal)
	}
	fmt.Fprint(os.Stdout, ": ")
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return "", sc.Err()
	}
	s := strings.TrimSpace(sc.Text())
	if s == "" {
		return defaultVal, nil
	}
	return s, nil
}

// validateDate returns true if s is YYYY-MM-DD and parses as a valid date.
func validateDate(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// normalizeSlug cleans a slug string (e.g. from user input): letters, numbers, hyphens only,
// collapsed and trimmed. If the result is empty, returns defaultSlug.
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
	out := regexp.MustCompile("-+").ReplaceAllString(regexp.MustCompile("^-|-$").ReplaceAllString(b.String(), ""), "-")
	if out == "" {
		return defaultSlug
	}
	return out
}

// --- Commands: new, edit, list, update-links ---

// runNew creates a new post. If args is non-empty, they are used as
// title, description, keywords, date, slug (missing ones are prompted).
// Writes posts/<slug>.html from the template and updates index.html + feed.xml.
func runNew(root string, args []string) error {
	templatePath := filepath.Join(root, "post-template.html")
	postsDir := filepath.Join(root, "posts")
	if _, err := os.Stat(templatePath); err != nil {
		return fmt.Errorf("post-template.html not found: %w", err)
	}
	if err := os.MkdirAll(postsDir, 0755); err != nil {
		return err
	}

	var title, description, keywords, dateArg, slug string
	if len(args) >= 1 && args[0] != "" {
		title = args[0]
	}
	if title == "" {
		var err error
		title, err = readLine("Title", "")
		if err != nil {
			return err
		}
		if title == "" {
			return fmt.Errorf("title is required")
		}
	}
	title = strings.TrimSpace(title)
	defaultSlug := slugify(title)

	// Optional args skip prompts when provided.
	if len(args) >= 2 {
		description = args[1]
	}
	if description == "" {
		description, _ = readLine("Description (meta, one line)", "")
	}
	if len(args) >= 3 {
		keywords = args[2]
	}
	if keywords == "" {
		keywords, _ = readLine("Keywords (comma-separated)", "")
	}
	if len(args) >= 4 {
		dateArg = args[3]
	}
	if dateArg == "" {
		dateArg, _ = readLine("Date (YYYY-MM-DD)", time.Now().UTC().Format("2006-01-02"))
	}
	dateArg = strings.TrimSpace(dateArg)
	if !validateDate(dateArg) {
		return fmt.Errorf("invalid date, use YYYY-MM-DD")
	}
	if len(args) >= 5 {
		slug = args[4]
	}
	if slug == "" {
		slug, _ = readLine("Slug (filename)", defaultSlug)
	}
	slug = normalizeSlug(slug, defaultSlug)
	if slug == "" {
		return fmt.Errorf("slug is required")
	}

	outPath := filepath.Join(postsDir, slug+".html")
	if _, err := os.Stat(outPath); err == nil {
		return fmt.Errorf("file already exists: %s", outPath)
	}

	tpl, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	displayDate := formatDisplayDate(dateArg)
	content := string(tpl)
	// Fill template placeholders (see post-template.html).
	content = strings.ReplaceAll(content, "POST_TITLE", escapeHTML(title))
	content = strings.ReplaceAll(content, "POST_DESCRIPTION", escapeHTML(strings.TrimSpace(description)))
	content = strings.ReplaceAll(content, "POST_KEYWORDS", escapeHTML(strings.TrimSpace(keywords)))
	content = strings.ReplaceAll(content, "POST_SLUG", slug)
	content = strings.ReplaceAll(content, "YYYY-MM-DD", dateArg)
	content = strings.ReplaceAll(content, "Month Day, Year", displayDate)
	// Ensure RSS link exists in <head> for post pages.
	if !strings.Contains(content, "application/rss+xml") {
		content = strings.Replace(content,
			"  <link rel=\"manifest\" href=\"../site.webmanifest\">\n  \n  <link rel=\"stylesheet\"",
			"  <link rel=\"manifest\" href=\"../site.webmanifest\">\n  <link rel=\"alternate\" type=\"application/rss+xml\" title=\"Jordan Joe Cooper\" href=\"../feed.xml\">\n\n  <link rel=\"stylesheet\"",
			1)
	}
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Println("Created:", outPath)

	postRel := filepath.Join("posts", slug+".html")
	if err := updateIndexAndFeed(root, filepath.Join(root, postRel)); err != nil {
		return fmt.Errorf("post created but failed to update index/feed: %w", err)
	}
	fmt.Println()
	fmt.Println("Next: edit the post body in", outPath)
	return nil
}

// updateIndexAndFeed reads a post file, extracts title/description/date from meta,
// and inserts a new entry at the top of the Writing list in index.html and at the
// top of the item list in feed.xml. Use after creating a new post or when you've
// manually added a post file.
func updateIndexAndFeed(root, postFullPath string) error {
	body, err := os.ReadFile(postFullPath)
	if err != nil {
		return err
	}
	html := string(body)
	title := extractTitle(html)
	description := extractMeta(html, "name", "description")
	if description == "" {
		description = extractMeta(html, "property", "og:description")
	}
	dateRaw := extractMeta(html, "name", "article:published_time")
	if dateRaw == "" {
		dateRaw = extractMeta(html, "property", "article:published_time")
	}
	if title == "" {
		return fmt.Errorf("could not extract title from post")
	}
	if description == "" {
		return fmt.Errorf("could not extract description from post")
	}
	if dateRaw == "" || !validateDate(dateRaw) {
		return fmt.Errorf("could not extract date (article:published_time YYYY-MM-DD)")
	}

	slug := strings.TrimSuffix(filepath.Base(postFullPath), ".html")
	displayDate := formatDisplayDate(dateRaw)
	rssDate, err := formatRSSDate(dateRaw)
	if err != nil {
		return err
	}
	url := baseURL + "/posts/" + slug + ".html"
	descEscaped := escapeXML(description)
	titleEscaped := escapeXML(title)
	titleForHTML := strings.ReplaceAll(strings.ReplaceAll(title, "&", "&amp;"), "<", "&lt;")

	// Insert new <li> as first item in the Writing list.
	indexPath := filepath.Join(root, "index.html")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	idx := strings.Index(string(indexContent), listMarker)
	if idx == -1 {
		return fmt.Errorf("could not find Writing list in index.html")
	}
	newLi := fmt.Sprintf(`      <li>
        <h2 class="post-title">
          <span class="post-date">%s</span>
          <a href="posts/%s.html">%s</a>
        </h2>
      </li>
      `, displayDate, slug, titleForHTML)
	insertAt := idx + len(listMarker)
	indexNew := string(indexContent)[:insertAt] + "\n" + newLi + string(indexContent)[insertAt:]
	if err := os.WriteFile(indexPath, []byte(indexNew), 0644); err != nil {
		return err
	}
	fmt.Println("Updated index.html (Writing section).")

	// Insert new <item> before the first existing item in the RSS feed.
	feedPath := filepath.Join(root, "feed.xml")
	feedContent, err := os.ReadFile(feedPath)
	if err != nil {
		return err
	}
	feedIdx := strings.Index(string(feedContent), itemMarker)
	if feedIdx == -1 {
		return fmt.Errorf("could not find <item> in feed.xml")
	}
	newItem := fmt.Sprintf(`    <item>
      <title>%s</title>
      <link>%s</link>
      <guid>%s</guid>
      <pubDate>%s</pubDate>
      <description>%s</description>
    </item>
    `, titleEscaped, url, url, rssDate, descEscaped)
	feedNew := string(feedContent)[:feedIdx] + newItem + string(feedContent)[feedIdx:]
	if err := os.WriteFile(feedPath, []byte(feedNew), 0644); err != nil {
		return err
	}
	fmt.Println("Updated feed.xml.")
	fmt.Println("Done. New post added to homepage and RSS:", title)
	return nil
}

// runEdit opens a post in the user's editor ($EDITOR or $VISUAL, else nano).
// If slugArg is empty, we list posts and ask which one to edit; otherwise we open that slug.
func runEdit(root, slugArg string) error {
	posts, err := getPosts(root)
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		return fmt.Errorf("no posts yet; create one with: postcli new")
	}

	var target post
	if slugArg != "" {
		for _, p := range posts {
			if p.Slug == strings.TrimSuffix(slugArg, ".html") {
				target = p
				break
			}
		}
		if target.Slug == "" {
			return fmt.Errorf("post not found: %s", slugArg)
		}
	} else {
		fmt.Println()
		for i, p := range posts {
			dateStr := ""
			if p.Date != "" {
				dateStr = " (" + p.Date + ")"
			}
			fmt.Printf("  %d) %s%s\n", i+1, p.Title, dateStr)
		}
		fmt.Println()
		var choice string
		choice, err = readLine("Edit which? (1-"+fmt.Sprint(len(posts))+")", "")
		if err != nil {
			return err
		}
		var n int
		if _, err := fmt.Sscanf(choice, "%d", &n); err != nil || n < 1 || n > len(posts) {
			return fmt.Errorf("invalid number")
		}
		target = posts[n-1]
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "nano"
	}
	fmt.Printf("Opening %s in %s...\n\n", target.File, editor)
	// Run editor with stdin/stdout/stderr attached so the user can edit interactively.
	cmd := exec.Command(editor, target.FullPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println("Done. Run 'postcli update-links posts/" + target.File + "' if you changed title/date/description.")
	return nil
}

// runList prints all posts (title and path) sorted by date, newest first.
func runList(root string) error {
	posts, err := getPosts(root)
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		fmt.Println("  No posts.")
		return nil
	}
	fmt.Println()
	for _, p := range posts {
		fmt.Println("  " + p.Title)
		dateStr := ""
		if p.Date != "" {
			dateStr = "  " + p.Date
		}
		fmt.Println("    posts/" + p.File + dateStr)
	}
	fmt.Println()
	return nil
}

// runUpdateLinks adds one post to the homepage and feed by path (e.g. posts/my-post.html).
// Use when you created a post file manually and need to register it.
func runUpdateLinks(root, postPath string) error {
	if postPath == "" {
		return fmt.Errorf("usage: postcli update-links posts/your-post.html")
	}
	if !strings.HasPrefix(postPath, "posts/") || !strings.HasSuffix(postPath, ".html") {
		return fmt.Errorf("path must be posts/...html")
	}
	fullPath := filepath.Join(root, postPath)
	if _, err := os.Stat(fullPath); err != nil {
		return fmt.Errorf("post file not found: %s", fullPath)
	}
	return updateIndexAndFeed(root, fullPath)
}

// --- Interactive menu ---

// runMenu shows the main menu in a loop until the user quits.
func runMenu(root string) error {
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n  Posts — jordanjoecooper.com\n")
		fmt.Println("  1) New post")
		fmt.Println("  2) Edit post")
		fmt.Println("  3) List posts")
		fmt.Println("  q) Quit\n")
		fmt.Print("  Choice (1-3 or q): ")
		if !sc.Scan() {
			return sc.Err()
		}
		choice := strings.TrimSpace(strings.ToLower(sc.Text()))
		switch choice {
		case "", "q", "quit":
			return nil
		case "1", "new", "n":
			if err := runNew(root, nil); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case "2", "edit", "e":
			if err := runEdit(root, ""); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case "3", "list", "l":
			_ = runList(root)
		default:
			fmt.Println("  Unknown option. Use 1-3 or q.")
		}
	}
}

// --- Entry point ---

// main finds the repo root, then runs the requested command (or the interactive menu).
func main() {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	args := os.Args[1:]
	switch {
	case len(args) == 0:
		if err := runMenu(root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case args[0] == "new" || args[0] == "n":
		if err := runNew(root, args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case args[0] == "edit" || args[0] == "e":
		slug := ""
		if len(args) >= 2 {
			slug = args[1]
		}
		if err := runEdit(root, slug); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case args[0] == "list" || args[0] == "l":
		if err := runList(root); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case args[0] == "update-links" || args[0] == "sync":
		pathArg := ""
		if len(args) >= 2 {
			pathArg = args[1]
		}
		if err := runUpdateLinks(root, pathArg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "Usage: postcli [new|edit|list|update-links] [args...]")
		fmt.Fprintln(os.Stderr, "  new           create a new post (prompts for title, description, etc.)")
		fmt.Fprintln(os.Stderr, "  edit [slug]   edit a post (pick from list or open by slug)")
		fmt.Fprintln(os.Stderr, "  list          list all posts")
		fmt.Fprintln(os.Stderr, "  update-links  update index.html and feed.xml from a post file")
		os.Exit(1)
	}
}
</think>
Fixing the editor command: using `os/exec`.
<｜tool▁calls▁begin｜><｜tool▁call▁begin｜>
StrReplace