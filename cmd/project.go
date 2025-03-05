package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"slices"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zenginechris/devx/cmd/root"
	"github.com/zenginechris/devx/config"
	"github.com/zenginechris/devx/internal/clients"
	"github.com/zenginechris/devx/internal/filesystem"
	"github.com/zenginechris/devx/internal/projects"
)

func init() {
	listProjectsCmd.Flags().BoolVarP(&listProjectsCmdArgs.json, "json", "j", false, "print json output")
	root.Cmd().AddCommand(listProjectsCmd)
	root.Cmd().AddCommand(setProjectContextCmd)
	root.Cmd().AddCommand(buildProjectCmd)
	root.Cmd().AddCommand(editProjectCmd)
    root.Cmd().AddCommand(createProjectCmd)
}

var listProjectsCmdArgs struct {
	json bool
}

var listProjectsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Short:   "List all projects",
	Long:    "List all configured projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := config.Load()
		if err != nil {
			logrus.Error(err)
		}

		if len(config.Projects) == 0 {
			logrus.Warn("No projects configured. Run `devx project new` to create one.")
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 4, 8, 4, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tCONTEXT\tVERSION\t")

		for _, pro := range config.Projects {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", pro.Name, pro.Context, "v100")
		}

		return w.Flush()

	},
}

var createProjectCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Create new project",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.Load()
		if err != nil {
			logrus.Error(err)
		}

		projectName := args[0]
		configPath := slugify(projectName)

		cfg.AddProject(projects.Project{
			Name:       projectName,
			ConfigPath: path.Join(config.ProjectsDir(), configPath),
		})

		err = config.Save(cfg, config.GetProfile().File())
		if err != nil {
			logrus.Error(err)
		}

		CreateDockerfileContent(path.Join(config.ProjectsDir(), configPath, "Dockerfile"))
		return nil
	},
}

var setProjectContextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"c"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Set the build context for a project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string

		if len(args) == 1 {
			path, _ = os.Getwd()
		} else {
			path = args[1]
		}

		cfg, err := config.Load()
		if err != nil {
			logrus.Error(err)
		}

		pro := cfg.FindProject(args[0])
		pro.Context = path
		cfg.UpdateProject(pro)

		err = config.Save(cfg, config.GetProfile().File())
		if err != nil {
			logrus.Error(err)
		}
		return nil
	},
}

// we create some config files for each project and make them editable
// we create the the project based on that confics

var editProjectCmd = &cobra.Command{
	Use:     "edit",
	Aliases: []string{"e"},
	Args:    cobra.MinimumNArgs(2),
	Short:   "Edit the configuration for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			logrus.Error(err)
		}
		project := cfg.FindProject(args[0])
		logrus.Info(config.ProjectsDir())
		// open the editor
		editor := getEditor()

		switch args[1] {
		case "docker":
			c := exec.Command(editor, path.Join(project.ConfigPath, "Dockerfile"))
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		}

		return nil
	},
}

var buildProjectCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Build and update",
	Long:    "Build a new image from the current context and updates the image in the current k8s cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			logrus.Error(err)
		}
		project := cfg.FindProject(args[0])

		tempDir, err := os.MkdirTemp("", "docker-build-*")
		if err != nil {
			logrus.Error(fmt.Sprintf("Failed to create temp directory: %v\n", err))
			return err
		}
		logrus.Info(fmt.Sprintf("Created temporary directory: %s\n", tempDir))

		defer func() {
			logrus.Info(fmt.Sprintf("Cleaning up temporary directory: %s\n", tempDir))
			os.RemoveAll(tempDir)
		}()

		sourcePaths := []string{
			project.Context,
			fmt.Sprintf("%s/%s", project.ConfigPath, "Dockerfile"),
		}

		sourcePaths = append(sourcePaths, project.Contexts...)

		for _, sourcePath := range sourcePaths {
			err := filesystem.CopyToTemp(sourcePath, tempDir)
			if err != nil {
				logrus.Error(fmt.Sprintf("Failed to copy %s: %v\n", sourcePath, err))
				return err
			}
		}

		dockerfilePath := filepath.Join(tempDir, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			logrus.Error(fmt.Sprintln("Error: No Dockerfile found in the temporary directory."))
			logrus.Error(fmt.Sprintln("Make sure one of your source paths contains a Dockerfile."))
			return err
		}

		checkDriverCmd := exec.Command("docker", "buildx", "inspect")
		driverOutput, err := checkDriverCmd.CombinedOutput()
		usingDockerDriver := true
		if err == nil && len(driverOutput) > 0 {
			usingDockerDriver = len(driverOutput) > 0 && string(driverOutput) != "" && string(driverOutput) != "null" &&
				(string(driverOutput) != "Driver: docker" || string(driverOutput) != "docker-container")
		}

		newImage := fmt.Sprintf("devx_%s:%v", slugify(project.Name), time.Now().UnixMilli())

		buildOpts := []string{
			"buildx",
			"build",
			"--platform=linux/amd64",
			"--tag", newImage,
			"--build-arg", "BUILD_DATE=" + currentTimeRFC3339(),
			"--label", "org.opencontainers.image.created=" + currentTimeRFC3339(),
			"--progress=plain",
			"--load",
			".",
		}

		if !usingDockerDriver {
			buildOpts[2] = "--platform=linux/amd64,linux/arm64"
			buildOpts = append(buildOpts[:4], append([]string{
				"--cache-from=type=local,src=/tmp/buildcache",
				"--cache-to=type=local,dest=/tmp/buildcache,mode=max",
			}, buildOpts[4:]...)...)

			buildOpts = removeOption(buildOpts, "--load")
		}

		dockerCmd := exec.Command("docker", buildOpts...)
		dockerCmd.Dir = tempDir
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		logrus.Info(fmt.Sprintln("Running docker build command..."))
		err = dockerCmd.Run()
		if err != nil {
			logrus.Error(fmt.Sprintf("Docker build failed: %v\n", err))
			return err
		}

		fmt.Println("Docker build completed successfully")

		clients.UpdateDeployment(project, newImage)

		return nil

	},
}

func currentTimeRFC3339() string {
	return time.Now().Format(time.RFC3339)
}

func removeOption(slice []string, option string) []string {
	for i, item := range slice {
		if item == option {
			return slices.Delete(slice, i, i+1)
		}
	}
	return slice
}

func getEditor() string {

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else if _, err := exec.LookPath("vim"); err == nil {
			editor = "vim"
		} else if _, err := exec.LookPath("vi"); err == nil {
			editor = "vi"
		} else {
			return "no suitable editor found. Please set EDITOR environment variable"
		}
	}
	return editor
}

func slugify(input string) string {
	lowercase := strings.ToLower(input)
	slugified := strings.ReplaceAll(lowercase, " ", "-")
	return slugified
}

type DockerfileTemplate struct {
	BaseImage   string
	WorkDir     string
	ExposedPort string
	Command     string
}

func CreateDockerfileContent(outputPath string) {

	dockerfileTemplate := `FROM {{.BaseImage}}
WORKDIR {{.WorkDir}}

COPY . .

RUN go mod download
RUN go build -o app

EXPOSE {{.ExposedPort}}

CMD {{.Command}}
`

	data := DockerfileTemplate{
		BaseImage:   "golang:1.21-alpine",
		WorkDir:     "/app",
		ExposedPort: "8080",
		Command:     "[\"./app\"]",
	}

	if err := writeFileFromTemplate(dockerfileTemplate, data, outputPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

}

func writeFileFromTemplate(templateStr string, data any, outputPath string) error {
	tmpl, err := template.New("dockerfile").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
