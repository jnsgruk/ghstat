package main

import (
	"fmt"
	"log/slog"
	"os"

	"jnsgruk/ghstat/internal/ghstat"

	"github.com/spf13/cobra"
)

var (
	version string = "dev"
	commit  string = "dev"
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
		flags := cmd.PersistentFlags()
		verbose, _ := flags.GetBool("verbose")
		output, _ := flags.GetString("output")
		configFile, _ := flags.GetString("config")
		leads, _ := flags.GetStringSlice("leads")

		// Ensure the slog logger is set for the correct format/log level
		setupLogging(verbose)

		// Load and validate the configuration file
		conf, err := ghstat.ParseConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to parse configuration: %w", err)
		}

		conf.Filter = leads
		conf.Verbose = verbose
		conf.Formatter = output

		mgr, err := ghstat.NewManager(conf, os.Stdout)
		if err != nil {
			return err
		}
		return mgr.Execute()
	},
}

func setupLogging(verbose bool) {
	logLevel := new(slog.LevelVar)

	// Set the default log level to "INFO", and "DEBUG" if verbose is specified.
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	// Setup the TextHandler and ensure our configured logger is the default.
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	slog.SetDefault(logger)
	logLevel.Set(level)
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.BoolP("verbose", "v", false, "enable verbose logging")
	flags.StringP("output", "o", "pretty", "choose the output format ('pretty', 'markdown' or 'json')")
	flags.StringP("config", "c", "", "path to a specific config file to use")
	flags.StringSliceP("leads", "l", []string{}, "filter results to specific hiring leads from the config")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
