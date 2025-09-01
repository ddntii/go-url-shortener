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

type Config struct {
	DefaultCodeLength int `json:"default_code_length"`
	TitleMaxLength    int `json:"title_max_length"`
	CleanupDays       int `json:"cleanup_days"`
}

type Store struct {
	Items map[string]URLEntry `json:"items"`
	Stats struct {
		TotalClicks int `json:"total_clicks"`
		TotalURLs   int `json:"total_urls"`
	} `json:"stats"`
}

const charset = "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[90m"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[37m"
)

func loadConfig() Config {
	defaultConfig := Config{
		DefaultCodeLength: 4,
		TitleMaxLength:    60,
		CleanupDays:       30,
	}
	
	data, err := os.ReadFile("config.json")
	if err != nil {
		return defaultConfig
	}
	
	var config Config
	if json.Unmarshal(data, &config) != nil {
		return defaultConfig
	}
	
	return config
}

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

func fetchTitle(rawURL string, config Config) string {
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
		if len(title) > config.TitleMaxLength {
			title = title[:config.TitleMaxLength] + "..."
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

func findExisting(store Store, targetURL string) string {
	for code, entry := range store.Items {
		if entry.URL == targetURL {
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
	config := loadConfig()

	switch cmd {
	case "shorten", "s":
		if len(os.Args) < 3 {
			fmt.Printf("Usage: urlsh s <url> [custom-code]\n")
			return
		}
		
		targetURL := os.Args[2]
		if !validateURL(targetURL) {
			fmt.Printf("%sInvalid URL format%s\n", ColorRed, ColorReset)
			return
		}

		if existing := findExisting(store, targetURL); existing != "" {
			fmt.Printf("%s%s%s %s(exists)%s\n", ColorGreen, existing, ColorReset, ColorGray, ColorReset)
			return
		}

		var code string
		if len(os.Args) > 3 {
			
			code = os.Args[3]
			if len(code) < 3 || len(code) > 20 {
				fmt.Printf("Custom code must be 3-20 characters\n")
				return
			}
			if _, exists := store.Items[code]; exists {
				fmt.Printf("%sCode '%s' already taken%s\n", ColorRed, code, ColorReset)
				return
			}
		} else {
		
			codeLen := config.DefaultCodeLength
			if len(store.Items) > 1000 {
				codeLen = 5
			}
			if len(store.Items) > 10000 {
				codeLen = 6
			}
			
			attempts := 0
			for {
				code = genCode(codeLen)
				if _, exists := store.Items[code]; !exists {
					break
				}
				attempts++
				if attempts > 10 {
					codeLen++
					attempts = 0
				}
			}
		}

		title := fetchTitle(targetURL, config)

		store.Items[code] = URLEntry{
			URL:       targetURL,
			CreatedAt: time.Now(),
			Clicks:    0,
			Title:     title,
		}

		saveDB(store)
		fmt.Printf("%s%s%s\n", ColorGreen, code, ColorReset)
		if title != "" {
			fmt.Printf("%sTitle: %s%s\n", ColorGray, title, ColorReset)
		}

	case "expand", "e":
		if len(os.Args) < 3 {
			fmt.Printf("Usage: urlsh e <code>\n")
			return
		}
		
		code := os.Args[2]
		entry, exists := store.Items[code]
		if !exists {
			fmt.Printf("%sError: Code '%s' not found%s\n", ColorRed, code, ColorReset)
			return
		}

		entry.Clicks++
		entry.LastClick = time.Now()
		store.Items[code] = entry
		saveDB(store)

		fmt.Printf("%s%s%s\n", ColorBlue, entry.URL, ColorReset)
		if entry.Title != "" {
			fmt.Printf("%sTitle: %s%s\n", ColorGray, entry.Title, ColorReset)
		}
		fmt.Printf("%sClicks: %d%s\n", ColorYellow, entry.Clicks, ColorReset)

	case "list", "l":
		if len(store.Items) == 0 {
			fmt.Printf("No URLs stored\n")
			return
		}

		type sortEntry struct {
			code  string
			entry URLEntry
		}
		var entries []sortEntry
		for code, entry := range store.Items {
			entries = append(entries, sortEntry{code, entry})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].entry.CreatedAt.After(entries[j].entry.CreatedAt)
		})

		fmt.Printf("Stored URLs (%d total):\n\n", len(store.Items))
		for _, e := range entries {
			fmt.Printf("Code: %s\n", e.code)
			fmt.Printf("URL:  %s\n", e.entry.URL)
			if e.entry.Title != "" {
				fmt.Printf("Title: %s\n", e.entry.Title)
			}
			fmt.Printf("Created: %s\n", e.entry.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Clicks: %d\n", e.entry.Clicks)
			if !e.entry.LastClick.IsZero() {
				fmt.Printf("Last clicked: %s\n", e.entry.LastClick.Format("2006-01-02 15:04:05"))
			}
			fmt.Printf("\n")
		}

	case "stats":
		fmt.Printf("URL Shortener Statistics\n")
		fmt.Printf("========================\n")
		fmt.Printf("Total URLs: %d\n", len(store.Items))
		
		totalClicks := 0
		var mostClicked URLEntry
		var mostClickedCode string
		oldestTime := time.Now()
		newestTime := time.Time{}
		
		for code, entry := range store.Items {
			totalClicks += entry.Clicks
			if entry.Clicks > mostClicked.Clicks {
				mostClicked = entry
				mostClickedCode = code
			}
			if entry.CreatedAt.Before(oldestTime) {
				oldestTime = entry.CreatedAt
			}
			if entry.CreatedAt.After(newestTime) {
				newestTime = entry.CreatedAt
			}
		}
		
		fmt.Printf("Total clicks: %d\n", totalClicks)
		if len(store.Items) > 0 {
			fmt.Printf("Average clicks per URL: %.1f\n", float64(totalClicks)/float64(len(store.Items)))
			fmt.Printf("Most clicked: %s (%d clicks)\n", mostClickedCode, mostClicked.Clicks)
			fmt.Printf("Oldest URL: %s\n", oldestTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("Newest URL: %s\n", newestTime.Format("2006-01-02 15:04:05"))
		}

	case "clean":
		fmt.Printf("Clean unused URLs (0 clicks, older than 30 days)? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Printf("Cancelled\n")
			return
		}

		cutoff := time.Now().AddDate(0, 0, -config.CleanupDays)
		removed := 0
		
		for code, entry := range store.Items {
			if entry.Clicks == 0 && entry.CreatedAt.Before(cutoff) {
				delete(store.Items, code)
				removed++
			}
		}
		
		saveDB(store)
		fmt.Printf("%sRemoved %d unused URLs%s\n", ColorYellow, removed, ColorReset)

	case "delete", "del", "rm":
		if len(os.Args) < 3 {
			fmt.Printf("Usage: urlsh delete <code>\n")
			return
		}
		
		code := os.Args[2]
		entry, exists := store.Items[code]
		if !exists {
			fmt.Printf("%sError: Code '%s' not found%s\n", ColorRed, code, ColorReset)
			return
		}

		fmt.Printf("Delete '%s' -> %s? [y/N]: ", code, entry.URL)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Printf("Cancelled\n")
			return
		}

		delete(store.Items, code)
		saveDB(store)
		fmt.Printf("Deleted %s\n", code)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Printf("Available commands: shorten, expand, list, stats, clean, delete\n")
	}
}
