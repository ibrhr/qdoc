package docsource

import (
	"strings"

	"golang.org/x/net/html"
)

func extractLinks(htmlContent, prefix, baseURL string) []Entry {
	seen := map[string]bool{}
	var entries []Entry

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return entries
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.HasPrefix(attr.Val, prefix) {
					path := strings.TrimPrefix(attr.Val, prefix)
					path = strings.TrimSuffix(path, "/")
					if path == "" {
						continue
					}

					if blockedPath(path) {
						continue
					}

					fullURL := strings.TrimSuffix(baseURL, "/") + "/" + path
					if !seen[path] {
						seen[path] = true
						title := extractText(n)
						if title == "" {
							title = path
						}
						entries = append(entries, Entry{
							URL:   fullURL,
							Title: title,
						})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return entries
}

func blockedPath(p string) bool {
	suffixes := []string{".png", ".jpg", ".svg", ".gif", ".ico", ".css", ".js"}
	for _, s := range suffixes {
		if strings.HasSuffix(p, s) {
			return true
		}
	}
	bad := []string{"#", "mailto:", "http://", "twitter.com/", "github.com/", "youtube.com/"}
	for _, b := range bad {
		if strings.Contains(p, b) {
			return true
		}
	}
	return false
}

func extractText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}

func extractMainContent(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return truncate(htmlContent)
	}

	var content strings.Builder
	var inTarget bool
	depth := 0
	foundTarget := false

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && !foundTarget {
			for _, attr := range n.Attr {
				v := strings.ToLower(attr.Val)
				k := strings.ToLower(attr.Key)
				if (k == "id" || k == "class") &&
					(strings.Contains(v, "main-content") ||
						strings.Contains(v, "article") ||
						strings.Contains(v, "content") ||
						strings.Contains(v, "markdown") ||
						strings.Contains(v, "prose") ||
						strings.Contains(v, "doc") ||
						strings.Contains(v, "documentation")) {
					inTarget = true
					foundTarget = true
				}
			}
		}

		wasTarget := inTarget
		if wasTarget {
			depth++
		}

		if inTarget && n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				content.WriteString(text)
				content.WriteString(" ")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}

		if wasTarget {
			depth--
			if depth == 0 {
				inTarget = false
			}
		}
	}
	walk(doc)

	result := strings.TrimSpace(content.String())
	if len(result) < 100 {
		result = stripHTML(htmlContent)
	}
	return truncate(result)
}

func stripHTML(htmlContent string) string {
	htmlContent = strings.ReplaceAll(htmlContent, "</p>", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "</div>", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "<br>", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "<br/>", "\n")
	htmlContent = strings.ReplaceAll(htmlContent, "</h1>", "\n\n")
	htmlContent = strings.ReplaceAll(htmlContent, "</h2>", "\n\n")
	htmlContent = strings.ReplaceAll(htmlContent, "</h3>", "\n")

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return truncate(htmlContent)
	}

	var text strings.Builder
	var walk func(*html.Node)
	skipDepth := 0
	skipData := map[string]bool{
		"script": true, "style": true, "nav": true,
		"header": true, "footer": true, "noscript": true,
		"svg": true,
	}

	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if skipData[n.Data] {
				skipDepth++
			}
		}
		if n.Type == html.TextNode && skipDepth == 0 {
			t := strings.TrimSpace(n.Data)
			if t != "" {
				text.WriteString(t)
				text.WriteString(" ")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if n.Type == html.ElementNode {
			if skipData[n.Data] {
				skipDepth--
			}
		}
	}
	walk(doc)
	return truncate(strings.TrimSpace(text.String()))
}

func truncate(s string) string {
	if len(s) > maxContentChars {
		return s[:maxContentChars]
	}
	return s
}