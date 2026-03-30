package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FetchSkill downloads a skill from GitHub to a temp directory and returns the path.
// spec format: "owner/repo/path/to/skill" or "owner/repo" (checks root for SKILL.md)
func FetchSkill(spec string) (string, func(), error) {
	parts := strings.SplitN(spec, "/", 3)
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("invalid GitHub spec %q: need at least owner/repo", spec)
	}

	owner := parts[0]
	repo := parts[1]
	skillPath := ""
	if len(parts) == 3 {
		skillPath = parts[2]
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "skillscore-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	client := &http.Client{}
	token := os.Getenv("GITHUB_TOKEN")

	// Fetch SKILL.md
	skillMDPath := skillPath
	if skillMDPath != "" {
		skillMDPath += "/"
	}
	skillMDPath += "SKILL.md"

	content, err := fetchFile(client, token, owner, repo, skillMDPath)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("fetch SKILL.md: %w", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), content, 0644); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("write SKILL.md: %w", err)
	}

	// Fetch directory listing to get supporting files
	dirEntries, err := fetchDirListing(client, token, owner, repo, skillPath)
	if err == nil {
		for _, entry := range dirEntries {
			if entry.Name == "SKILL.md" {
				continue
			}
			if entry.Type == "file" {
				fileContent, err := fetchFile(client, token, owner, repo, entry.Path)
				if err != nil {
					continue
				}
				destPath := filepath.Join(tmpDir, entry.Name)
				os.WriteFile(destPath, fileContent, 0644)
			} else if entry.Type == "dir" {
				// Fetch subdirectory contents
				subEntries, err := fetchDirListing(client, token, owner, repo, entry.Path)
				if err != nil {
					continue
				}
				subDir := filepath.Join(tmpDir, entry.Name)
				os.MkdirAll(subDir, 0755)
				for _, sub := range subEntries {
					if sub.Type == "file" {
						fileContent, err := fetchFile(client, token, owner, repo, sub.Path)
						if err != nil {
							continue
						}
						os.WriteFile(filepath.Join(subDir, sub.Name), fileContent, 0644)
					}
				}
			}
		}
	}

	return tmpDir, cleanup, nil
}

// DirEntry represents a file or directory in a GitHub repo.
type DirEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"` // "file" or "dir"
}

// ListDir lists the contents of a directory in a GitHub repo.
func ListDir(owner, repo, path string) ([]DirEntry, error) {
	client := &http.Client{}
	token := os.Getenv("GITHUB_TOKEN")
	return fetchDirListing(client, token, owner, repo, path)
}

// treeEntry represents a single entry from the Git Tree API response.
type treeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"` // "blob" or "tree"
}

// treeResponse represents the Git Tree API response.
type treeResponse struct {
	Tree      []treeEntry `json:"tree"`
	Truncated bool        `json:"truncated"`
}

// TreeDiscover returns all directory paths containing SKILL.md in a repo.
// Uses the Git Tree API (/git/trees/:sha?recursive=1) which has no 1000-item cap.
// Returns paths like "skills/deploy-app" for each directory containing a SKILL.md.
func TreeDiscover(owner, repo string) ([]string, error) {
	client := &http.Client{}
	token := os.Getenv("GITHUB_TOKEN")

	// Try "main" branch first, fall back to "HEAD"
	branches := []string{"main", "HEAD"}
	var treeResp treeResponse
	var lastErr error

	for _, branch := range branches {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, branch)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 404 {
			lastErr = fmt.Errorf("branch %q not found", branch)
			continue
		}
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("HTTP %d for tree %s/%s@%s: %s", resp.StatusCode, owner, repo, branch, string(body))
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
			lastErr = fmt.Errorf("decode tree response: %w", err)
			continue
		}
		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("tree discover %s/%s: %w", owner, repo, lastErr)
	}

	if treeResp.Truncated {
		fmt.Fprintf(os.Stderr, "  Warning: tree for %s/%s was truncated\n", owner, repo)
	}

	// Find all paths ending with SKILL.md and return their parent directories
	var paths []string
	for _, entry := range treeResp.Tree {
		if entry.Type == "blob" && strings.HasSuffix(entry.Path, "/SKILL.md") {
			dir := strings.TrimSuffix(entry.Path, "/SKILL.md")
			paths = append(paths, dir)
		} else if entry.Type == "blob" && entry.Path == "SKILL.md" {
			// SKILL.md at repo root
			paths = append(paths, "")
		}
	}

	return paths, nil
}

func fetchFile(client *http.Client, token, owner, repo, path string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		// Try HEAD branch
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/HEAD/%s", owner, repo, path)
		req, _ = http.NewRequest("GET", url, nil)
		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d for %s/%s/%s", resp.StatusCode, owner, repo, path)
	}

	return io.ReadAll(resp.Body)
}

func fetchDirListing(client *http.Client, token, owner, repo, path string) ([]DirEntry, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d listing %s", resp.StatusCode, path)
	}

	var entries []DirEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}
