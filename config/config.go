package config

import (
	"fmt"
	"os"

	"github.com/ZengineChris/devx/internal/projects"
	"github.com/pelletier/go-toml/v2"
)

const (
	AppName = "devx"
)

type VersionInfo struct {
	Version  string
	Revision string
}

var (
	appVersion = "development"
	revision   = "unknown"
)

func AppVersion() VersionInfo {
	return VersionInfo{
		Version:  appVersion,
		Revision: revision,
	}
}

type Config struct {
	Projects []projects.Project `toml:"projects"`
}

func (c *Config) AddProject(p projects.Project) {
	c.Projects = append(c.Projects, p)
}

func (c *Config) UpdateProject(p projects.Project) {
	for i, pro := range c.Projects {
		if p.Name == pro.Name {
			c.Projects[i] = p
		}
	}
}
func (c *Config) FindProject(name string) projects.Project {
	for _, pro := range c.Projects {
		if name == pro.Name {
			return pro
		}
	}

	return projects.Project{}
}

func Save(c Config, file string) error {
	b, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, b, 0644); err != nil {
		return fmt.Errorf("error writing yaml file: %w", err)
	}

	return nil

}

func LoadFrom(file string) (Config, error) {
	var c Config
	b, err := os.ReadFile(file)
	if err != nil {
		return c, fmt.Errorf("could not load config from file: %w", err)
	}

	err = toml.Unmarshal(b, &c)
	if err != nil {
		return c, fmt.Errorf("could not load config from file: %w", err)
	}

	return c, nil
}

func Load() (Config, error) {
	f := GetProfile().File()
	if _, err := os.Stat(f); err != nil {
		// the file dose not exist, so write one from the config
		err := Save(Config{}, f)
		if err != nil {
			return Config{}, err
		}

	}

	return LoadFrom(f)
}
