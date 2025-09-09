package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/github/github-mcp-server/internal/ghmcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// These variables are set by the build process using ldflags.
var version = "version"
var commit = "commit"
var date = "date"

var (
	rootCmd = &cobra.Command{
		Use:     "mcp-prime",
		Short:   "MCP PRIME - Repository to MCP Conversion Tool",
		Long:    `A Model Context Protocol (MCP) server that converts repositories into MCP-compatible tools.`,
		Version: fmt.Sprintf("Version: %s\nCommit: %s\nBuild Date: %s", version, commit, date),
	}

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio MCP server",
		Long:  `Start an MCP server that communicates via standard input/output streams using JSON-RPC messages.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			stdioServerConfig := ghmcp.StdioServerConfig{
				Version:              version,
				EnabledToolsets:      []string{"repository"},
				DynamicToolsets:      false,
				ReadOnly:             false,
				ExportTranslations:   viper.GetBool("export-translations"),
				EnableCommandLogging: viper.GetBool("enable-command-logging"),
				LogFilePath:          viper.GetString("log-file"),
				ContentWindowSize:    viper.GetInt("content-window-size"),
			}
			return ghmcp.RunRepositoryStdioServer(stdioServerConfig)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetGlobalNormalizationFunc(wordSepNormalizeFunc)

	rootCmd.SetVersionTemplate("{{.Short}}\n{{.Version}}\n")

	// Add global flags
	rootCmd.PersistentFlags().String("log-file", "", "Path to log file")
	rootCmd.PersistentFlags().Bool("enable-command-logging", false, "When enabled, the server will log all command requests and responses to the log file")
	rootCmd.PersistentFlags().Bool("export-translations", false, "Save translations to a JSON file")
	rootCmd.PersistentFlags().Int("content-window-size", 5000, "Specify the content window size")

	// Bind flags to viper
	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("enable-command-logging", rootCmd.PersistentFlags().Lookup("enable-command-logging"))
	_ = viper.BindPFlag("export-translations", rootCmd.PersistentFlags().Lookup("export-translations"))
	_ = viper.BindPFlag("content-window-size", rootCmd.PersistentFlags().Lookup("content-window-size"))

	// Add subcommands
	rootCmd.AddCommand(stdioCmd)
}

func initConfig() {
	// Initialize Viper configuration
	viper.SetEnvPrefix("MCP_PRIME")
	viper.AutomaticEnv()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func wordSepNormalizeFunc(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	from := []string{"_"}
	to := "-"
	for _, sep := range from {
		name = strings.ReplaceAll(name, sep, to)
	}
	return pflag.NormalizedName(name)
}
