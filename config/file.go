package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

// requiredDir is a directory that must exist on the filesystem
type requiredDir struct {
	once sync.Once

	// dir is a func to enable deferring the value of the directory
	// until execution time.
	// if dir() returns an error, a fatal error is triggered.
	dir func() (string, error)

	computedDir *string
}

// Dir returns the directory path.
// It ensures the directory is created on the filesystem by calling
// `mkdir` prior to returning the directory path.
func (r *requiredDir) Dir() string {
	if r.computedDir != nil {
		return *r.computedDir
	}

	dir, err := r.dir()
	if err != nil {
		logrus.Fatal(fmt.Errorf("cannot fetch required directory: %w", err))
	}

	r.once.Do(func() {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logrus.Fatal(fmt.Errorf("cannot make required directory: %w", err))
		}
	})

	r.computedDir = &dir
	return dir
}

var (
	configBaseDir = requiredDir{
		dir: func() (string, error) {
			dir := os.Getenv("DEVX_HOME")
			if _, err := os.Stat(dir); err == nil {
				return dir, nil
			}

			// user home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			dir = filepath.Join(homeDir, ".devx")
			_, err = os.Stat(dir)

			// extra xdg config directory
			xdgDir, xdg := os.LookupEnv("XDG_CONFIG_HOME")

			if err == nil {
				// ~/.devx is found but xdg dir is set
				if xdg {
					logrus.Warnln("found ~/.devx, ignoring $XDG_CONFIG_HOME...")
					logrus.Warnln("delete ~/.devx to use $XDG_CONFIG_HOME as config directory")
					logrus.Warnf("or run `mv ~/.devx \"%s\"`", filepath.Join(xdgDir, "devx"))
				}
				return dir, nil
			} else {
				// ~/.devx is missing and xdg dir is set
				if xdg {
					return filepath.Join(xdgDir, "devx"), nil
				}
			}

			// macOS users are accustomed to ~/.devx
			if runtime.GOOS == "darwin" {
				return dir, nil
			}

			// other environments fall back to user config directory
			dir, err = os.UserConfigDir()
			if err != nil {
				return "", err
			}

			return filepath.Join(dir, "devx"), nil
		},
	}

	cacheDir = requiredDir{
		dir: func() (string, error) {
			dir := os.Getenv("XDG_CACHE_HOME")
			if dir != "" {
				return filepath.Join(dir, "devx"), nil
			}
			// else
			dir, err := os.UserCacheDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(dir, "devx"), nil
		},
	}

	templatesDir = requiredDir{
		dir: func() (string, error) {
			dir, err := configBaseDir.dir()
			if err != nil {
				return "", err
			}
			return filepath.Join(dir, "_templates"), nil
		},
	}


	projectsDir = requiredDir{
		dir: func() (string, error) {
			dir, err := configBaseDir.dir()
			if err != nil {
				return "", err
			}
			return filepath.Join(dir, "projects"), nil
		},
	}
)

// CacheDir returns the cache directory.
func CacheDir() string { return cacheDir.Dir() }

// TemplatesDir returns the templates' directory.
func TemplatesDir() string { return templatesDir.Dir() }


func ProjectsDir() string { return projectsDir.Dir() }

const configFileName = "devx.toml"
