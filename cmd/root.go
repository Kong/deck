package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/kong/deck/utils"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	rootConfig    utils.KongClientConfig
	konnectConfig utils.KonnectConfig

	disableAnalytics bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deck",
	Short: "Administer your Kong clusters declaratively",
	Long: `The deck tool helps you manage Kong clusters with a declarative
configuration file.

It can be used to export, import, or sync entities to Kong.`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := url.ParseRequestURI(rootConfig.Address); err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		return nil
	},
}

// RootCmdOnlyForDocs is used to generate makrdown documentation.
var RootCmdOnlyForDocs = rootCmd

// Execute adds all child commands to the root command and sets
// flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		// do not print error because cobra already prints it
		os.Exit(1)
	}
}

//nolint:errcheck
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Config file (default is $HOME/.deck.yaml).")

	rootCmd.PersistentFlags().String("kong-addr", "http://localhost:8001",
		"HTTP address of Kong's Admin API.\n"+
			"This value can also be set using the environment variable DECK_KONG_ADDR\n"+
			" environment variable.")
	viper.BindPFlag("kong-addr",
		rootCmd.PersistentFlags().Lookup("kong-addr"))

	rootCmd.PersistentFlags().StringSlice("headers", []string{},
		"HTTP headers (key:value) to inject in all requests to Kong's Admin API.\n"+
			"This flag can be specified multiple times to inject multiple headers.")
	viper.BindPFlag("headers",
		rootCmd.PersistentFlags().Lookup("headers"))

	rootCmd.PersistentFlags().Bool("tls-skip-verify", false,
		"Disable verification of Kong's Admin TLS certificate.\n"+
			"This value can also be set using DECK_TLS_SKIP_VERIFY "+
			"environment variable.")
	viper.BindPFlag("tls-skip-verify",
		rootCmd.PersistentFlags().Lookup("tls-skip-verify"))

	rootCmd.PersistentFlags().String("tls-server-name", "",
		"Name to use to verify the hostname in "+
			"Kong's Admin TLS certificate.\n"+
			"This value can also be set using DECK_TLS_SERVER_NAME"+
			" environment variable.")
	viper.BindPFlag("tls-server-name",
		rootCmd.PersistentFlags().Lookup("tls-server-name"))

	rootCmd.PersistentFlags().String("ca-cert", "",
		"Custom CA certificate (raw contents) to use to "+
			"verify Kong's Admin TLS certificate.\n"+
			"This value can also be set using DECK_CA_CERT"+
			" environment variable.\n"+
			"This takes precedence over --ca-cert-file flag.")
	viper.BindPFlag("ca-cert",
		rootCmd.PersistentFlags().Lookup("ca-cert"))

	rootCmd.PersistentFlags().String("ca-cert-file", "",
		"Path to a custom CA certificate to use "+
			"to verify Kong's Admin TLS certificate.\n"+
			"This value can also be set using DECK_CA_CERT_FILE"+
			" environment variable.")
	viper.BindPFlag("ca-cert-file",
		rootCmd.PersistentFlags().Lookup("ca-cert-file"))

	rootCmd.PersistentFlags().Int("verbose", 0,
		"Enable verbose logging levels\n"+
			"Setting this value to 2 outputs all HTTP requests/responses\n"+
			"between decK and Kong.")
	viper.BindPFlag("verbose",
		rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.PersistentFlags().Bool("no-color", false,
		"Disable colorized output")
	viper.BindPFlag("no-color",
		rootCmd.PersistentFlags().Lookup("no-color"))

	rootCmd.PersistentFlags().Bool("skip-workspace-crud", false,
		"Skip API calls related to Workspaces (Kong Enterprise only).")
	viper.BindPFlag("skip-workspace-crud",
		rootCmd.PersistentFlags().Lookup("skip-workspace-crud"))

	// konnect-specific flags
	rootCmd.PersistentFlags().String("konnect-email", "",
		"Email address associated with your Konnect account.")
	viper.BindPFlag("konnect-email",
		rootCmd.PersistentFlags().Lookup("konnect-email"))

	rootCmd.PersistentFlags().String("konnect-password", "",
		"Password associated with your Konnect account, "+
			"this takes precedence over --konnect-password-file flag.")
	viper.BindPFlag("konnect-password",
		rootCmd.PersistentFlags().Lookup("konnect-password"))

	rootCmd.PersistentFlags().String("konnect-password-file", "",
		"File containing the password to your Konnect account.")
	viper.BindPFlag("konnect-password-file",
		rootCmd.PersistentFlags().Lookup("konnect-password-file"))

	rootCmd.PersistentFlags().String("konnect-addr", "https://konnect.konghq.com",
		"Address of the Konnect endpoint.")
	viper.BindPFlag("konnect-addr",
		rootCmd.PersistentFlags().Lookup("konnect-addr"))

	rootCmd.PersistentFlags().Bool("analytics", true,
		"Share anonymized data to help improve decK.")
	viper.BindPFlag("analytics",
		rootCmd.PersistentFlags().Lookup("analytics"))

	rootCmd.PersistentFlags().Int("timeout", 10,
		"Set a request timeout for the client to connect with Kong (in seconds).")
	viper.BindPFlag("timeout",
		rootCmd.PersistentFlags().Lookup("timeout"))

	rootCmd.PersistentFlags().String("tls-client-cert", "",
		"Set the Kong's Admin TLS client certificate.\n"+
			"This value can also be set using DECK_TLS_CLIENT_CERT "+
			"environment variable. Must be used in conjunction with tls-client-key")
	viper.BindPFlag("tls-client-cert",
		rootCmd.PersistentFlags().Lookup("tls-client-cert"))

	rootCmd.PersistentFlags().String("tls-client-key", "",
		"Set the Kong's Admin TLS client key.\n"+
			"This value can also be set using DECK_TLS_CLIENT_KEY "+
			"environment variable. Must be used in conjunction with tls-client-cert")
	viper.BindPFlag("tls-client-key",
		rootCmd.PersistentFlags().Lookup("tls-client-key"))
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
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	caCertContent := ""

	if viper.GetString("ca-cert") != "" {
		caCertContent = viper.GetString("ca-cert")
	} else if viper.GetString("ca-cert-file") != "" {
		fileContent, err := ioutil.ReadFile(viper.GetString("ca-cert-file"))
		if err != nil {
			fmt.Printf("read file %q: %s", viper.GetString("ca-cert-file"), err)
			os.Exit(1)
		}
		caCertContent = string(fileContent)
		caCertContent = strings.TrimRight(caCertContent, "\n")
	}

	rootConfig.Address = viper.GetString("kong-addr")
	rootConfig.TLSServerName = viper.GetString("tls-server-name")
	rootConfig.TLSSkipVerify = viper.GetBool("tls-skip-verify")
	rootConfig.TLSCACert = caCertContent
	rootConfig.Headers = viper.GetStringSlice("headers")
	rootConfig.SkipWorkspaceCrud = viper.GetBool("skip-workspace-crud")
	rootConfig.Debug = (viper.GetInt("verbose") >= 1)
	rootConfig.Timeout = (viper.GetInt("timeout"))
	rootConfig.TLSClientCert = viper.GetString("tls-client-cert")
	rootConfig.TLSClientKey = viper.GetString("tls-client-key")

	color.NoColor = (color.NoColor || viper.GetBool("no-color"))

	if err := initKonnectConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initKonnectConfig() error {
	password := viper.GetString("konnect-password")
	passwordFile := viper.GetString("konnect-password-file")
	// read from password file only if password is not supplied using an
	// environment variable or flag
	if password == "" && passwordFile != "" {
		fileContent, err := ioutil.ReadFile(passwordFile)
		if err != nil {
			return fmt.Errorf("read file %q: %w", passwordFile, err)
		}
		if len(fileContent) == 0 {
			return fmt.Errorf("file %q: empty", passwordFile)
		}
		password = string(fileContent)
		password = strings.TrimRight(password, "\n")
	}

	disableAnalytics = !viper.GetBool("analytics")
	konnectConfig.Email = viper.GetString("konnect-email")
	konnectConfig.Password = password
	konnectConfig.Debug = (viper.GetInt("verbose") >= 1)
	konnectConfig.Address = viper.GetString("konnect-addr")
	return nil
}
