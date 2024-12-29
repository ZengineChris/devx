package projects


type (
	Project struct {
		Name    string `toml:"name" json:"name"`
		Context string `toml:"context" json:"context"`
	}
)
