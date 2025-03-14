package filesystem

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PatternMatcher handles .dockerignore pattern matching
type PatternMatcher struct {
	patterns []pattern
}

// pattern represents a single .dockerignore pattern
type pattern struct {
	val       string
	isNegated bool
}

// NewPatternMatcher creates a new pattern matcher from a list of .dockerignore patterns
func NewPatternMatcher(patterns []string) *PatternMatcher {
	pm := &PatternMatcher{
		patterns: make([]pattern, 0, len(patterns)),
	}

	for _, p := range patterns {
		isNegated := false
		if strings.HasPrefix(p, "!") {
			isNegated = true
			p = p[1:]
		}

		pm.patterns = append(pm.patterns, pattern{
			val:       p,
			isNegated: isNegated,
		})
	}

	return pm
}

// Matches checks if a path matches any of the patterns
func (pm *PatternMatcher) Matches(path string) bool {
	// Last match determines the outcome
	ignored := false

	// Normalize path
	path = filepath.ToSlash(path)

	for _, pattern := range pm.patterns {
		matched := matchPattern(pattern.val, path)

		if matched {
			ignored = !pattern.isNegated
		}
	}

	return ignored
}

// matchPattern checks if a path matches a single pattern
func matchPattern(pattern, path string) bool {
	// Simple cases
	if pattern == path {
		return true
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		// Handle patterns like *.txt
		if strings.HasPrefix(pattern, "*") {
			suffix := pattern[1:]
			return strings.HasSuffix(path, suffix)
		}

		// Handle patterns like dir/*
		if strings.HasSuffix(pattern, "*") {
			prefix := pattern[:len(pattern)-1]
			return strings.HasPrefix(path, prefix)
		}

		// Handle patterns like dir/*.txt
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(path, parts[0]) && strings.HasSuffix(path, parts[1])
		}
	}

	// Handle directory patterns with trailing slash
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, pattern) || path == pattern[:len(pattern)-1]
	}

	// Handle patterns like dir/ which should match dir and all its contents
	return strings.HasPrefix(path, pattern+"/") || path == pattern
}

func CopyToTemp(source, tempDir string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("error getting source info: %w", err)
	}

	// Get the base name of the source
	baseName := filepath.Base(source)
	destPath := filepath.Join(tempDir, baseName)

	// Handle directory copy
	if sourceInfo.IsDir() {

		patterns, err := parseDockerignore(source)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error parsing .dockerignore: %w", err)
		}

		var matcher *PatternMatcher
		if len(patterns) > 0 {
			matcher = NewPatternMatcher(patterns)
			if err != nil {
				return fmt.Errorf("error creating pattern matcher: %w", err)
			}
		}

		// replace
		return copyDirWithDockerignore(source, destPath, matcher)
	}

	// Handle file copy
	return copyFile(source, destPath)
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	fmt.Printf("Copying file %s to %s\n", src, dst)

	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	// Copy contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying file contents: %w", err)
	}

	// Copy file permissions from source to destination
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting source file info: %w", err)
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func copyDirWithDockerignore(src, dst string, matcher *PatternMatcher) error {
	fmt.Printf("Copying directory %s to %s (with .dockerignore support)\n", src, dst)

	// Create the destination directory
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting source directory info: %w", err)
	}

	err = os.MkdirAll(dst, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == src {
			return nil
		}

		// Calculate the relative path from source
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("error calculating relative path: %w", err)
		}

		// Check if file should be ignored (if we have a matcher)
		if matcher != nil {
			if matcher.Matches(relPath) {
				fmt.Printf("Ignoring %s (matched .dockerignore rule)\n", relPath)
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Calculate the destination path
		destPath := filepath.Join(dst, relPath)

		// Handle directories and files
		if info.IsDir() {
			os.Chmod(destPath, 0644)
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// For files, handle any dot files (hidden files) specially
			baseName := filepath.Base(path)
			if len(baseName) > 0 && baseName[0] == '.' {
				fmt.Printf("Copying hidden file: %s\n", relPath)
			}
			os.Chmod(path, 0644)
			return copyFile(path, destPath)
		}
	})
}

func copyDir(src, dst string) error {
	fmt.Printf("Copying directory %s to %s\n", src, dst)

	// Create the destination directory
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error getting source directory info: %w", err)
	}

	err = os.MkdirAll(dst, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == src {
			return nil
		}

		// Calculate the relative path from source
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("error calculating relative path: %w", err)
		}

		// Calculate the destination path
		destPath := filepath.Join(dst, relPath)

		// Handle directories and files
		if info.IsDir() {
			os.Chmod(destPath, 0644)
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// For files, handle any dot files (hidden files) specially
			baseName := filepath.Base(path)
			if len(baseName) > 0 && baseName[0] == '.' {
				fmt.Printf("Copying hidden file: %s\n", relPath)
			}
			os.Chmod(path, 0644)
			return copyFile(path, destPath)
		}
	})
}

// parseDockerignore reads .dockerignore file and returns patterns
func parseDockerignore(dir string) ([]string, error) {
	dockerignorePath := filepath.Join(dir, ".dockerignore")

	file, err := os.Open(dockerignorePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		// Handle negated patterns (those starting with !)
		if strings.HasPrefix(pattern, "!") {
			pattern = "!" + filepath.ToSlash(pattern[1:])
		} else {
			pattern = filepath.ToSlash(pattern)
		}

		patterns = append(patterns, pattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}
