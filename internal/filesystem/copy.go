package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
		return copyDir(source, destPath)
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
