package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           int
	GitHubToken    string
	GitHubOrg      string
	CORSOrigins    []string
	CacheTTL       time.Duration
}

// LoadEnvFiles loads the first existing .env files found (later files do not override earlier).
func LoadEnvFiles() []string {
	var loaded []string
	seen := make(map[string]struct{})

	for _, p := range envCandidates() {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}

		if _, err := os.Stat(p); err != nil {
			continue
		}
		if err := godotenv.Load(p); err != nil {
			slog.Warn("failed to parse .env file", "path", p, "error", err)
			continue
		}
		loaded = append(loaded, p)
	}
	return loaded
}

func envCandidates() []string {
	var paths []string
	paths = append(paths, ".env", "backend/.env")

	if root := moduleRoot(); root != "" {
		paths = append(paths, filepath.Join(root, ".env"))
	}
	if wd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(wd, ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), ".env"))
	}
	return paths
}

func moduleRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func Load() Config {
	loaded := LoadEnvFiles()
	if len(loaded) > 0 {
		slog.Info("loaded env files", "paths", loaded)
	} else {
		slog.Warn("no .env file found", "hint", "run: cp backend/.env.example backend/.env && add GITHUB_TOKEN")
	}

	port := 8081
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	org := os.Getenv("GITHUB_ORG")
	if org == "" {
		org = "tektoncd"
	}

	origins := []string{"http://localhost:3000"}
	if raw := os.Getenv("CORS_ORIGINS"); raw != "" {
		origins = nil
		for _, o := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	cacheTTL := 15 * time.Minute
	if m := os.Getenv("CACHE_TTL_MINUTES"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v > 0 {
			cacheTTL = time.Duration(v) * time.Minute
		}
	}

	return Config{
		Port:        port,
		GitHubToken: strings.TrimSpace(os.Getenv("GITHUB_TOKEN")),
		GitHubOrg:   org,
		CORSOrigins: origins,
		CacheTTL:    cacheTTL,
	}
}

func (c Config) GitHubAuthenticated() bool {
	return c.GitHubToken != ""
}
