package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/zenginechris/devx/cmd/root"
	"github.com/zenginechris/devx/config"
	"github.com/zenginechris/devx/internal/projects"
)

func init() {
	projectCmd.AddCommand(setProjectContextCmd)
	projectCmd.AddCommand(createProjectCmd)
	projectCmd.AddCommand(listProjectsCmd)
	projectCmd.AddCommand(buildProjectCmd)
	listProjectsCmd.Flags().BoolVarP(&listProjectsCmdArgs.json, "json", "j", false, "print json output")

	root.Cmd().AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:     "project",
	Short:   "modify and work with projects",
	Aliases: []string{"p"},
}

var listProjectsCmdArgs struct {
	json bool
}

var listProjectsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Short:   "list all projects",
	Long:    "list all configured projects",
	RunE: func(cmd *cobra.Command, args []string) error {

		// read the configuratation
		config, err := config.Load()
		// if we have an error here we want to write the config file

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
	Short:   "create new project",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.Load()
		if err != nil {
			logrus.Error(err)
		}

		projectName := args[0]

		cfg.AddProject(projects.Project{
			Name: projectName,
		})

		err = config.Save(cfg, config.GetProfile().File())
		if err != nil {
			logrus.Error(err)
		}

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

var buildProjectCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Build a new image from the current context",
	RunE: func(cmd *cobra.Command, args []string) error {
		// find the project
		ctx := context.Background()
		cfg, err := config.Load()
		if err != nil {
			logrus.Error(err)
		}
		project := cfg.FindProject(args[0])

		// Frist we build it with the dockerfile in the context
		dockerClient, err := client.NewClientWithOpts(
            client.FromEnv, 
            client.WithAPIVersionNegotiation(),
        )
		if err != nil {
			return err
		}
		defer dockerClient.Close()

		buildOptions := types.ImageBuildOptions{
			Dockerfile: "./Dockerfile",
			Tags:       []string{"my-multiarch-app:latest"},
			BuildArgs: map[string]*string{
				"VERSION": stringPtr("1.0"),
			},
		}

		return buildxBuild(
			ctx,
			dockerClient,
			project.Context,
            buildOptions,
		)

	},
}

func stringPtr(s string) *string {
	return &s
}

func buildDockerImage(ctx context.Context, dockerClient *client.Client, dockerfilePath, contextPath, imageName string) error {
	buildContext, err := archive.TarWithOptions(contextPath, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("error creating build context: %v", err)
	}
	defer buildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: filepath.Base(dockerfilePath),
		Tags:       []string{imageName},
		Context:    buildContext,
	}

	response, err := dockerClient.ImageBuild(
		ctx,
		buildContext,
		buildOptions,
	)
	if err != nil {
		return fmt.Errorf("error building image: %v", err)
	}
	defer response.Body.Close()

	// Stream build output
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return fmt.Errorf("error reading build output: %v", err)
	}

	return nil
}

func buildxBuild(ctx context.Context, dockerClient *client.Client, projectContext string, options types.ImageBuildOptions) error {
	// Create buildx builder if not exists
	buildContext, err := archive.TarWithOptions(projectContext, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("error creating build context: %v", err)
	}
	defer buildContext.Close()

	// Build image with buildx
	response, err := dockerClient.ImageBuild(
		ctx,
		buildContext, // context reader
		types.ImageBuildOptions{
			Dockerfile:  options.Dockerfile,
			Tags:        options.Tags,
			BuildArgs:   options.BuildArgs,
			Platform:    "linux/arm64", // Multi-arch support
			PullParent:  true,
			Remove:      true,
			ForceRemove: true,
		},
	)
	if err != nil {
		return fmt.Errorf("buildx build error: %v", err)
	}
	defer response.Body.Close()

	// Stream build output
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return fmt.Errorf("build output error: %v", err)
	}

	return nil
}
