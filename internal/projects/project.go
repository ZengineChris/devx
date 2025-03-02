package projects

type (
	Project struct {
		Name           string   `toml:"name" json:"name"`
		Context        string   `toml:"context" json:"context"`
		ConfigPath     string   `toml:"config_path" json:"config_path"`
		Contexts       []string `toml:"contexts" json:"contexts"`
		DeploymentName string   `toml:"deployment_name" json:"deployment_name"`
		Namespace      string   `toml:"namespace" json:"namespace"`
	}
)
