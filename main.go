package main

import (
"crypto/sha1"
"encoding/json"
"fmt"
"math/big"
"net/url"
"os"
"path/filepath"
"strings"
"time"
)

type URLEntry struct {
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int       `json:"clicks"`
}

type Store struct {
    Items map[string]URLEntry 'json."items"'
}

func dbPath() string {
    return filepath.Join(".", "urls.json")
}

func load() Store {
    var s Store
    s.Items = map[string]URLEntry{}
    data, err := os.ReadFile(dbPath())
    if err != nil {
        return s 
    }
    if err :=json.Unmarshal(data, &s); err != nil {
    return Store{Items: map[string]URLEntry{}}
    }
    if s.Items == nil {
        s.Items = map[string]URLEntry{}
    }
    return s 
}

const alphabet 


