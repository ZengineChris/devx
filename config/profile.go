package config

import "path/filepath"

type (
	Profile struct {
		configDir *requiredDir
	}
)

func GetProfile() *Profile {
	return &Profile{}
}

func (p *Profile) ConfigDir() string {
	if p.configDir == nil {
		p.configDir = &requiredDir{
			dir: func() (string, error) {
				return filepath.Join(configBaseDir.Dir()), nil
			},
		}
	}
	return p.configDir.Dir()
}

func (p *Profile) File() string {
	return filepath.Join(p.ConfigDir(), configFileName)
}
