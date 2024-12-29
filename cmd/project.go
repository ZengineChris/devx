package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ZengineChris/devx/cmd/root"
	"github.com/ZengineChris/devx/config"
	"github.com/ZengineChris/devx/internal/projects"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	projectCmd.AddCommand(setProjectContextCmd)
	projectCmd.AddCommand(createProjectCmd)
	projectCmd.AddCommand(listProjectsCmd)
	listProjectsCmd.Flags().BoolVarP(&listProjectsCmdArgs.json, "json", "j", false, "print json output")

	root.Cmd().AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "modify and work with projects",
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
		_, _ = fmt.Fprintln(w, "NAME\tCONTEXT\tARCH\tCPUS\tMEMORY\tDISK\tRUNTIME\tADDRESS")

		for _, pro := range config.Projects {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", pro.Name, pro.Context)
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
