package docsource

import (
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxContentChars = 12000

type Entry struct {
	URL   string
	Title string
}

type Source struct {
	Name         string
	BaseURL      string
	IndexURL     string
	LinkPrefix   string
	Local        bool
	SystemPrompt string
}

var KnownSources = []Source{
	{
		Name:       "go",
		BaseURL:    "https://go.dev/doc",
		IndexURL:   "https://go.dev/doc/",
		LinkPrefix: "/doc/",
		SystemPrompt: `Go documentation is at https://go.dev/doc. All URLs in the file list below are complete — use them as-is.

Where things live under /doc/:
  tutorial/     — Step-by-step tutorials (getting-started, create-module, workspaces, web-service-gin, generics, etc.)
  modules/      — Module system (layout, dependencies, go.mod, publishing, versioning, etc.)
  database/     — Database access guides
  articles/     — In-depth articles (wiki, race_detector, go_command)
  codewalk/     — Guided code walkthroughs
  security/     — Security docs (fuzz, fips140)
  go1.X         — Release notes per Go version
  Top-level     — effective_go, code, install, faq, gc-guide, godebug, pgo, asm, gdb, editors, diagnostics, contribute, devel/release

Standard library packages: pkg.go.dev/<import-path>. Include the URL when relevant.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
	{
		Name:       "fastapi",
		BaseURL:    "https://fastapi.tiangolo.com",
		IndexURL:   "https://fastapi.tiangolo.com/",
		LinkPrefix: "",
		SystemPrompt: `FastAPI documentation is at https://fastapi.tiangolo.com. All URLs in the file list below are complete — use them as-is.

Where things live:
  tutorial/                  — Tutorial - User Guide: step-by-step walkthrough
    tutorial/first-steps       creating a minimal app with path operations
    tutorial/path-params       declaring path parameters with type hints
    tutorial/query-params      query parameters, defaults, and optional params
    tutorial/body              request body with Pydantic models
    tutorial/*-validations     string/numeric validations for query, path, cookie, header params
    tutorial/query-param-models  grouping query params into Pydantic models
    tutorial/body-*            multiple body params, Field(), nested models
    tutorial/response-model    return type annotation and response shaping
    tutorial/extra-models      multiple models, Union types, model inheritance
    tutorial/response-status-code  setting HTTP status codes
    tutorial/request-forms     Form data, Form Models, file uploads, forms+files
    tutorial/handling-errors   HTTPException, custom error handlers
    tutorial/path-operation-configuration  tags, summary, description, operation_id
    tutorial/encoder           JSON-compatible encoder for non-Pydantic returns
    tutorial/body-updates      PATCH with partial updates (exclude_unset)
    tutorial/background-tasks  BackgroundTasks for post-response work
    tutorial/metadata          title, description, version, docs URLs, tags metadata
    tutorial/static-files      serving static files with StaticFiles
    tutorial/testing           writing tests with TestClient
    tutorial/debugging         debug mode
    tutorial/bigger-applications  structuring with APIRouter
    tutorial/middleware        adding custom middleware
    tutorial/cors              CORS configuration
    tutorial/sql-databases     SQL (SQLAlchemy) database integration
    tutorial/sse               Server-Sent Events (SSE)
    tutorial/dependencies/     Depends(), function deps, classes as deps, sub-deps, yield deps
    tutorial/security/         OAuth2, JWT, HTTP Basic, get-current-user, scopes
  advanced/                  — Advanced User Guide
    advanced/custom-response   HTMLResponse, RedirectResponse, StreamingResponse, FileResponse
    advanced/response-headers  setting custom response headers
    advanced/response-cookies  setting cookies in response
    advanced/additional-responses  multiple response models per status code
    advanced/additional-status-codes  additional status codes in OpenAPI
    advanced/websockets        WebSocket endpoints
    advanced/events           startup/shutdown lifespan events
    advanced/settings         Pydantic Settings for configuration
    advanced/templates        Jinja2 HTML templates
    advanced/middleware        advanced middleware techniques
    advanced/sub-applications  mounting sub-FastAPI apps, proxying
    advanced/behind-a-proxy   running behind nginx/traefik
    advanced/dataclasses       using dataclasses instead of Pydantic
    advanced/generate-clients  generating API clients from OpenAPI
    advanced/openapi-webhooks  OpenAPI webhook docs
    advanced/openapi-callbacks  OpenAPI callback docs
    advanced/using-request-directly  accessing the raw Request object
    advanced/wsgi              mounting WSGI apps (Flask, Django)
    advanced/async-tests       testing async code
    advanced/stream-data       streaming data
    advanced/custom-response   customizing response status codes and content type
    advanced/security/         HTTP Basic auth, OAuth2 scopes
  how-to/                    — How-To Guides
    how-to/graphql            adding GraphQL with Strawberry
    how-to/configure-swagger-ui  customizing Swagger UI
    how-to/extending-openapi   custom OpenAPI schema
    how-to/conditional-openapi  conditionally enabling OpenAPI docs
    how-to/migrate-from-pydantic-v1-to-pydantic-v2
    how-to/separate-openapi-schemas  separate input/output OpenAPI schemas
    how-to/custom-request-and-route  custom request classes and APIRoute
    how-to/testing-database    testing with a database
  deployment/                — Deployment
    deployment/docker          Docker deployment guide
    deployment/https           HTTPS setup
    deployment/concepts        deployment concepts
    deployment/server-workers  server workers (Gunicorn, Uvicorn)
    deployment/cloud           cloud deployment overview
  reference/                — API Reference (class and method docs)

For "how do I do X" questions, start with tutorial/. For advanced patterns, check advanced/. For specific class/method behavior, see reference/. For Docker, see deployment/docker.

Pick the most relevant pages and read them. Use the full URLs from the list below.`,
	},
}

func Find(name string) (Source, bool) {
	for _, s := range KnownSources {
		if strings.EqualFold(s.Name, name) {
			return s, true
		}
	}

	info, err := os.Stat(name)
	if err == nil && info.IsDir() {
		abs, _ := filepath.Abs(name)
		return Source{
			Name:     filepath.Base(abs),
			BaseURL:  "file://" + abs,
			IndexURL: "file://" + abs,
			Local:    true,
		}, true
	}

	return Source{}, false
}

func (s Source) FetchIndex() ([]Entry, error) {
	if s.Local {
		return s.fetchLocalIndex()
	}

	resp, err := retryableHTTPGet(s.IndexURL)
	if err != nil {
		return nil, fmt.Errorf("fetching index: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading index: %w", err)
	}

	entries := extractLinks(string(body), s.LinkPrefix, s.BaseURL)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].URL < entries[j].URL
	})
	return entries, nil
}

func (s Source) FetchContent(rawURL string) (string, error) {
	if s.Local {
		return s.fetchLocalContent(rawURL)
	}

	resp, err := retryableHTTPGet(rawURL)
	if err != nil {
		return "", fmt.Errorf("fetching %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", rawURL, err)
	}

	content := extractMainContent(string(body))
	if len(content) > maxContentChars {
		content = content[:maxContentChars]
	}
	return content, nil
}

func (s Source) fetchLocalIndex() ([]Entry, error) {
	u, _ := url.Parse(s.IndexURL)
	rootDir := filepath.Join(u.Host, u.Path)

	var entries []Entry
	seen := map[string]bool{}
	extensions := map[string]bool{
		".md": true, ".mdx": true, ".html": true,
		".rst": true, ".txt": true, ".adoc": true,
	}

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !extensions[ext] {
			return nil
		}
		rel, _ := filepath.Rel(rootDir, path)
		if seen[rel] {
			return nil
		}
		seen[rel] = true
		title := strings.TrimSuffix(strings.TrimSuffix(rel, ext), "/index")
		title = strings.ReplaceAll(title, "/", " / ")
		entries = append(entries, Entry{URL: rel, Title: title})
		return nil
	})

	if err != nil {
		return entries, nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].URL < entries[j].URL
	})
	return entries, nil
}

func (s Source) fetchLocalContent(path string) (string, error) {
	u, _ := url.Parse(s.BaseURL)
	rootDir := filepath.Join(u.Host, u.Path)
	fullPath := filepath.Join(rootDir, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)
	if len(content) > maxContentChars {
		content = content[:maxContentChars]
	}
	return content, nil
}

func fetchLocal(rawURL string) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(u.Host, u.Path)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       f,
	}, nil
}

type retryConfig struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
}

func backoffDelay(rc retryConfig, attempt int) time.Duration {
	delay := time.Duration(float64(rc.baseDelay) * math.Pow(2, float64(attempt)))
	if delay > rc.maxDelay {
		delay = rc.maxDelay
	}
	return delay + time.Duration(float64(delay)*0.3*rand.Float64())
}

var fetchRetry = retryConfig{
	maxAttempts: 3,
	baseDelay:   1 * time.Second,
	maxDelay:    10 * time.Second,
}

func retryableHTTPGet(rawURL string) (*http.Response, error) {
	if strings.HasPrefix(rawURL, "file://") {
		return fetchLocal(rawURL)
	}

	var lastErr error
	for attempt := 0; attempt < fetchRetry.maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(backoffDelay(fetchRetry, attempt-1))
		}

		resp, err := http.Get(rawURL)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, fmt.Errorf("not found: %s", rawURL)
		}

		return resp, nil
	}

	return nil, fmt.Errorf("fetch retry exhausted (%d attempts): %w", fetchRetry.maxAttempts, lastErr)
}