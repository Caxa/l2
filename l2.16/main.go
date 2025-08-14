package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Job struct {
	URL   string
	Depth int
}

var (
	visited   = make(map[string]bool)
	visitedMu sync.Mutex
	client    = &http.Client{
		Timeout: 10 * time.Second,
	}
	domain      string
	basePath    = "mirror"
	maxFileSize = int64(5 * 1024 * 1024) // 5 MB
	robotsRules = make(map[string]bool)
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <url> <depth>")
		return
	}

	startURL := os.Args[1]
	depth := atoi(os.Args[2])

	// Путь для сохранения
	repoPath, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory:", err)
		return
	}
	basePath = filepath.Join(repoPath, "mirror")
	fmt.Println("Сохраняю сайт в папку:", basePath)

	// Создаём папку mirror
	if err := os.MkdirAll(basePath, 0755); err != nil {
		fmt.Println("Ошибка создания папки:", err)
		return
	}

	u, err := url.Parse(startURL)
	if err != nil {
		fmt.Println("Invalid URL:", err)
		return
	}
	domain = u.Host

	loadRobots(startURL)

	jobs := make(chan Job, 100)
	var wg sync.WaitGroup

	// worker pool
	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		go worker(jobs, &wg)
	}

	// стартовое задание
	wg.Add(1)
	jobs <- Job{URL: startURL, Depth: depth}

	// ждём выполнения всех заданий
	wg.Wait()
	close(jobs)

	fmt.Println("Скачивание завершено.")
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func worker(jobs chan Job, wg *sync.WaitGroup) {
	for job := range jobs {
		processPage(job.URL, job.Depth, jobs, wg)
		wg.Done()
	}
}

func processPage(rawurl string, depth int, jobs chan Job, wg *sync.WaitGroup) {
	if depth < 0 {
		return
	}

	visitedMu.Lock()
	if visited[rawurl] {
		visitedMu.Unlock()
		return
	}
	visited[rawurl] = true
	visitedMu.Unlock()

	if !isAllowedByRobots(rawurl) {
		fmt.Println("[ROBOTS BLOCKED]", rawurl)
		return
	}

	localPath, isHTML := downloadFile(rawurl)
	if localPath == "" || !isHTML {
		return
	}

	f, err := os.Open(localPath)
	if err != nil {
		return
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return
	}

	var walker func(*html.Node)
	walker = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i := range n.Attr {
				attr := &n.Attr[i]
				if isResourceAttr(n.Data, attr.Key) {
					resURL := resolveURL(rawurl, attr.Val)
					if resURL == "" {
						continue
					}
					parsed, err := url.Parse(resURL)
					if err != nil {
						continue
					}
					if parsed.Host == "" || parsed.Host == domain {
						resLocal, _ := downloadFile(resURL)
						if resLocal != "" {
							attr.Val, _ = filepath.Rel(filepath.Dir(localPath), resLocal)
						}
						if n.Data == "a" {
							wg.Add(1)
							jobs <- Job{URL: resURL, Depth: depth - 1}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walker(c)
		}
	}
	walker(doc)

	out, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer out.Close()
	html.Render(out, doc)
}

func isResourceAttr(tag, attr string) bool {
	return (tag == "a" && attr == "href") ||
		(tag == "img" && attr == "src") ||
		(tag == "link" && attr == "href") ||
		(tag == "script" && attr == "src")
}

func downloadFile(rawurl string) (string, bool) {
	resp, err := client.Get(rawurl)
	if err != nil {
		fmt.Println("[ERR]", err)
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("[ERR]", rawurl, resp.Status)
		return "", false
	}

	// Ограничение размера
	reader := io.LimitReader(resp.Body, maxFileSize)

	u, _ := url.Parse(rawurl)
	filePath := u.Path
	if filePath == "" || strings.HasSuffix(filePath, "/") {
		filePath += "index.html"
	}

	localPath := filepath.Join(basePath, u.Host, filePath)
	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		fmt.Println("[ERR mkdir]", err)
		return "", false
	}

	out, err := os.Create(localPath)
	if err != nil {
		fmt.Println("[ERR]", err)
		return "", false
	}
	defer out.Close()

	_, err = io.Copy(out, reader)
	if err != nil {
		return "", false
	}

	fmt.Println("[DOWNLOAD]", rawurl, "->", localPath)

	contentType := resp.Header.Get("Content-Type")
	isHTML := strings.HasPrefix(contentType, "text/html")
	return localPath, isHTML
}

func resolveURL(base string, href string) string {
	if strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "javascript:") {
		return ""
	}
	u, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(u).String()
}

func loadRobots(startURL string) {
	u, _ := url.Parse(startURL)
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", u.Scheme, u.Host)
	resp, err := client.Get(robotsURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var userAgentAllowed = false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToLower(line), "user-agent:") {
			ua := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "user-agent:")))
			if ua == "*" {
				userAgentAllowed = true
			} else {
				userAgentAllowed = false
			}
		}
		if userAgentAllowed && strings.HasPrefix(strings.ToLower(line), "disallow:") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "disallow:"))
			if path != "" {
				robotsRules[path] = true
			}
		}
	}
}

func isAllowedByRobots(rawurl string) bool {
	u, err := url.Parse(rawurl)
	if err != nil {
		return false
	}
	for p := range robotsRules {
		if strings.HasPrefix(u.Path, p) {
			return false
		}
	}
	return true
}
