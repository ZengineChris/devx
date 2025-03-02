package filesystem

import (
	"fmt"
	"io"
	"io/fs"
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

// copyDir recursively copies a directory from src to dst
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

	// Walk through the source directory
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
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
		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return fmt.Errorf("error getting directory info: %w", err)
			}
			return os.MkdirAll(destPath, info.Mode())
		} else {
			return copyFile(path, destPath)
		}
	})
}
