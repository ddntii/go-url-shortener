package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type URLEntry struct {
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int       `json:"clicks"`
	Title     string    `json:"title,omitempty"`
	LastClick time.Time `json:"last_click,omitempty"`
}

type Store struct {
	Items map[string]URLEntry `json:"items"`
	Stats struct {
		TotalClicks int `json:"total_clicks"`
		TotalURLs   int `json:"total_urls"`
	} `json:"stats"`
}

const charset = "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

func genCode(n int) string {
	b := make([]byte, n)
	io.ReadFull(rand.Reader, b)
	
	code := make([]byte, n)
	for i := range b {
		code[i] = charset[int(b[i])%len(charset)]
	}
	return string(code)
}

func loadDB() Store {
	data, err := os.ReadFile("urls.json")
	if err != nil {
		return Store{Items: make(map[string]URLEntry)}
	}
	
	var s Store
	if json.Unmarshal(data, &s) != nil {
		return Store{Items: make(map[string]URLEntry)}
	}
	
	if s.Items == nil {
		s.Items = make(map[string]URLEntry)
	}
	return s
}

func saveDB(s Store) {
	s.Stats.TotalURLs = len(s.Items)
	s.Stats.TotalClicks = 0
	for _, entry := range s.Items {
		s.Stats.TotalClicks += entry.Clicks
	}
	
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile("urls.json", data, 0644)
}

func fetchTitle(rawURL string) string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	body := make([]byte, 8192) 
	n, _ := resp.Body.Read(body)
	
	re := regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	matches := re.FindSubmatch(body[:n])
	if len(matches) > 1 {
		title := strings.TrimSpace(string(matches[1]))
		if len(title) > 60 {
			title = title[:60] + "..."
		}
		return title
	}
	return ""
}

func validateURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

		return u.Scheme != "" && u.Host != "" && 
		   (u.Scheme == "http" || u.Scheme == "https")
}

funcFindExisting(store, Store, targetURL string) string {
	for code, entry := range store.Items {
		if entry.URL = targetURL {
			return code
		}
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("urlsh v1.2 - URL shortener\n")
		fmt.Printf("Commands: shorten, expand, list, stats, clean\n")
		return
	}

	cmd := os.Args[1]
	store := loadDB()

	switch cmd {
	case "shorten", "s":
		if len(os.Args) < 3 {
			fmt.Printf("Usage: urlsh s <url> [custom-code]\n")
			return
		}
		
		targetURL := os.Args[2]
		if !validateURL(targetURL) {
			fmt.Printf("Invalid URL format\n")
			return
		}





