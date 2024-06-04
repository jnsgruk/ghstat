package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"

	"jnsgruk/ghstat/internal/ghstat"

	"github.com/slok/gospinner"
	"github.com/spf13/cobra"
)

var (
	version string = "dev"
	commit  string = "dev"
	// Whether or not to use verbose logging
	verbose bool

	// The formatter specified at the command line with -f/--format
	selectedFormatter string   = "pretty"
	validFormatters   []string = []string{"pretty", "markdown", "json"}
	// Instantiation of the selected formatter
	formatter ghstat.Formatter

	configFile string
)

var shortDesc = "A utility for gathering role-specific statistics from Greenhouse."
var longDesc string = `A utility for gathering role-specific statistics from Greenhouse.

ghstat provides automation for gather statistics about a given Hiring Lead and the roles
they manage as part of Canonical's hiring process.

This tool is configured using a single file in one of the following locations:

	- ./ghstat.yaml
	- $HOME/.config/ghstat/ghstat.yaml

The configuration file should specify a top-level 'leads' list:

leads:
	# Name of the hiring lead
	- name: Joe Bloggs
	roles:
		# ID of the role in Greenhouse
		- 1234567

By default, ghstat will try to reuse an active Greenhouse session by reading the cookies
from a previous invocation. In the case that this isn't possible, it will prompt
for Ubuntu One credentials. To streamline login, the following environment variables can be set:

  - U1_LOGIN - the username/email for Ubuntu One login
  - U1_PASSWORD - the password for Ubuntu One login

For more information, visit the homepage at: https://github.com/jnsgruk/ghstat
`

var rootCmd = &cobra.Command{
	Use:           "ghstat",
	Version:       fmt.Sprintf("%s (%s)", version, commit),
	Short:         shortDesc,
	Long:          longDesc,
	SilenceErrors: true,
	SilenceUsage:  true,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure the slog logger is set for the correct format/log level
		ghstat.SetupLogger(verbose)

		// Validate the choice of formatter from the command line
		if !slices.Contains(validFormatters, selectedFormatter) {
			return fmt.Errorf("invalid output formatter specified, please choose one of 'pretty', 'markdown' or 'json'")
		}

		switch selectedFormatter {
		case "pretty":
			formatter = &ghstat.PrettyTableFormatter{}
			break
		case "markdown":
			formatter = &ghstat.MarkdownTableFormatter{}
			break
		case "json":
			formatter = &ghstat.JsonFormatter{}
			break
		}

		// Load and validate the configuration file
		conf, err := ghstat.ParseConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to parse configuration: %w", err)
		}

		manager := ghstat.NewManager(conf, formatter)

		// Ensure that we can launch a browser, and login to Greenhouse
		err = manager.Init()
		if err != nil {
			return fmt.Errorf("failed to initialise ghstat: %w", err)
		}

		// Show a spinner, unless verbose logging is switched on, then omit the spinner
		// so that the two don't fight over stdout/stderr
		s, _ := gospinner.NewSpinnerWithColor(gospinner.Dots, gospinner.FgGreen)
		if !verbose {
			s.Start("Processing Greenhouse roles...")
		}

		// Start gathering stats about the configured roles
		manager.Process()

		if !verbose {
			s.Finish()
		}

		// Dump the results to stdout using the specified formatter
		manager.Output()
		return nil
	},
}

func main() {
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.Flags().StringVarP(&selectedFormatter, "format", "f", "pretty", "choose the output format ('pretty', 'markdown' or 'json')")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "path to a specific config file to use")
	err := rootCmd.Execute()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
