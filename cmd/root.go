package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  kongClientConfig
	verbose int
	noColor bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deck",
	Short: "Administer your Kong declaritively",
	Long: `decK helps you manage your Kong clusters in a declaritive fashion.

You can export your existing Kong configuration, reset your Kong clusters.

It is also possible to use deck in your CI/CD pipeline to manage your Kong
configuration via GitOps.`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := url.ParseRequestURI(config.Address); err != nil {
			return errors.WithStack(errors.Wrap(err, "invalid URL"))
		}
		if noColor {
			color.NoColor = true
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets
// sflags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is $HOME/.deck.yaml)")

	rootCmd.PersistentFlags().String("kong-addr", "http://localhost:8001",
		"HTTP Address of Kong's Admin API.\n"+
			"This value can also be set using DECK_KONG_ADDR"+
			" environment variable.")
	viper.BindPFlag("kong-addr",
		rootCmd.PersistentFlags().Lookup("kong-addr"))

	rootCmd.PersistentFlags().Bool("tls-skip-verify", false,
		"Disable verification of Kong's Admin TLS certificate.\n"+
			"This value can also be set using DECK_TLS_SKIP_VERIFY "+
			"environment variable.")
	viper.BindPFlag("tls-skip-verify",
		rootCmd.PersistentFlags().Lookup("tls-skip-verify"))

	rootCmd.PersistentFlags().String("tls-server-name", "",
		"Custom CA certificate to use to verify"+
			"Kong's Admin TLS certificate.\n"+
			"This value can also be set using DECK_TLS_SERVER_NAME"+
			" environment variable.")
	viper.BindPFlag("tls-server-name",
		rootCmd.PersistentFlags().Lookup("tls-server-name"))

	rootCmd.PersistentFlags().String("ca-cert", "",
		"Custom CA certificate to use to verify"+
			"Kong's Admin TLS certificate.\n"+
			"This value can also be set using DECK_CA_CERT"+
			" environment variable.")
	viper.BindPFlag("ca-cert",
		rootCmd.PersistentFlags().Lookup("ca-cert"))

	rootCmd.PersistentFlags().Int("verbose", 0,
		"enable verbose verbose logging")
	viper.BindPFlag("verbose",
		rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.PersistentFlags().Bool("no-color", false,
		"disable colorized output")
	viper.BindPFlag("no-color",
		rootCmd.PersistentFlags().Lookup("no-color"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".deck"(without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".deck")
	}
	viper.SetEnvPrefix("deck")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	viper.ReadInConfig()
	config.Address = viper.GetString("kong-addr")
	config.TLSServerName = viper.GetString("tls-server-name")
	config.TLSSkipVerify = viper.GetBool("tls-skip-verify")
	config.TLSCACert = viper.GetString("ca-cert")
	verbose = viper.GetInt("verbose")
	noColor = viper.GetBool("no-color")

	if verbose < 1 {
		// supress output of terraform DAG if verbose logging is not enabled
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	if verbose >= 2 {
		config.Debug = true
	}
}
