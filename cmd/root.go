package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/kong/deck/utils"
	"github.com/kong/go-apiops/deckformat"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultKongURL    = "http://localhost:8001"
	defaultKonnectURL = "https://us.api.konghq.com"
)

var (
	cfgFile       string
	rootConfig    utils.KongClientConfig
	konnectConfig utils.KonnectConfig

	disableAnalytics bool

	konnectRuntimeGroup string
	konnectControlPlane string
)

// NewRootCmd represents the base command when called without any subcommands
//
//nolint:errcheck
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"Config file (default is $HOME/.deck.yaml).")

	rootCmd.PersistentFlags().String("kong-addr", defaultKongURL,
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
			"This takes precedence over `--ca-cert-file` flag.")
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
			"Sets the verbosity level of log output (higher is more verbose).")
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

	// Support for Session Cookie
	rootCmd.PersistentFlags().String("kong-cookie-jar-path", "",
		"Absolute path to a cookie-jar file in the Netscape cookie format for auth with Admin Server.\n"+
			"You may also need to pass in as header the User-Agent that was used to create the cookie-jar.")
	viper.BindPFlag("kong-cookie-jar-path",
		rootCmd.PersistentFlags().Lookup("kong-cookie-jar-path"))

	rootCmd.PersistentFlags().Bool("analytics", true,
		"Share anonymized data to help improve decK.\n"+
			"Use `--analytics=false` to disable this.")
	viper.BindPFlag("analytics",
		rootCmd.PersistentFlags().Lookup("analytics"))

	rootCmd.PersistentFlags().Int("timeout", 10,
		"Set a request timeout for the client to connect with Kong (in seconds).")
	viper.BindPFlag("timeout",
		rootCmd.PersistentFlags().Lookup("timeout"))

	rootCmd.PersistentFlags().String("tls-client-cert", "",
		"PEM-encoded TLS client certificate to use for authentication with Kong's Admin API.\n"+
			"This value can also be set using DECK_TLS_CLIENT_CERT "+
			"environment variable. Must be used in conjunction with tls-client-key")
	viper.BindPFlag("tls-client-cert",
		rootCmd.PersistentFlags().Lookup("tls-client-cert"))

	rootCmd.PersistentFlags().String("tls-client-cert-file", "",
		"Path to the file containing TLS client certificate to use for authentication with Kong's Admin API.\n"+
			"This value can also be set using DECK_TLS_CLIENT_CERT_FILE "+
			"environment variable. Must be used in conjunction with tls-client-key-file")
	viper.BindPFlag("tls-client-cert",
		rootCmd.PersistentFlags().Lookup("tls-client-cert-file"))

	rootCmd.PersistentFlags().String("tls-client-key", "",
		"PEM-encoded private key for the corresponding client certificate .\n"+
			"This value can also be set using DECK_TLS_CLIENT_KEY "+
			"environment variable. Must be used in conjunction with tls-client-cert")
	viper.BindPFlag("tls-client-key",
		rootCmd.PersistentFlags().Lookup("tls-client-key"))

	rootCmd.PersistentFlags().String("tls-client-key-file", "",
		"Path to file containing the private key for the corresponding client certificate.\n"+
			"This value can also be set using DECK_TLS_CLIENT_KEY_FILE "+
			"environment variable. Must be used in conjunction with tls-client-cert-file")
	viper.BindPFlag("tls-client-key",
		rootCmd.PersistentFlags().Lookup("tls-client-key-file"))

	// konnect-specific flags
	rootCmd.PersistentFlags().String("konnect-email", "",
		"Email address associated with your Konnect account.")
	viper.BindPFlag("konnect-email",
		rootCmd.PersistentFlags().Lookup("konnect-email"))

	rootCmd.PersistentFlags().String("konnect-password", "",
		"Password associated with your Konnect account, "+
			"this takes precedence over `--konnect-password-file` flag.")
	viper.BindPFlag("konnect-password",
		rootCmd.PersistentFlags().Lookup("konnect-password"))

	rootCmd.PersistentFlags().String("konnect-password-file", "",
		"File containing the password to your Konnect account.")
	viper.BindPFlag("konnect-password-file",
		rootCmd.PersistentFlags().Lookup("konnect-password-file"))

	rootCmd.PersistentFlags().String("konnect-token", "",
		"Personal Access Token associated with your Konnect account, "+
			"this takes precedence over `--konnect-token-file` flag.")
	viper.BindPFlag("konnect-token",
		rootCmd.PersistentFlags().Lookup("konnect-token"))

	rootCmd.PersistentFlags().String("konnect-token-file", "",
		"File containing the Personal Access Token to your Konnect account.")
	viper.BindPFlag("konnect-token-file",
		rootCmd.PersistentFlags().Lookup("konnect-token-file"))

	// user must provide at most one token to authenticate to Konnect
	rootCmd.MarkFlagsMutuallyExclusive("konnect-token-file", "konnect-token")

	rootCmd.PersistentFlags().String("konnect-addr", defaultKonnectURL,
		"Address of the Konnect endpoint.")
	viper.BindPFlag("konnect-addr",
		rootCmd.PersistentFlags().Lookup("konnect-addr"))

	rootCmd.PersistentFlags().String("konnect-runtime-group-name", "",
		"Konnect Runtime group name.")
	rootCmd.PersistentFlags().MarkDeprecated(
		"konnect-runtime-group-name", "use --konnect-control-plane-name instead")
	viper.BindPFlag("konnect-runtime-group-name",
		rootCmd.PersistentFlags().Lookup("konnect-runtime-group-name"))

	rootCmd.PersistentFlags().String("konnect-control-plane-name", "",
		"Konnect Control Plane name.")
	viper.BindPFlag("konnect-control-plane-name",
		rootCmd.PersistentFlags().Lookup("konnect-control-plane-name"))

	rootCmd.MarkFlagsMutuallyExclusive("konnect-runtime-group-name", "konnect-control-plane-name")

	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newResetCmd())
	rootCmd.AddCommand(newPingCmd())
	rootCmd.AddCommand(newDumpCmd())
	rootCmd.AddCommand(newDiffCmd())
	rootCmd.AddCommand(newConvertCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newKonnectCmd())
	{
		fileCmd := newAddFileCmd()
		rootCmd.AddCommand(fileCmd)
		fileCmd.AddCommand(newAddPluginsCmd())
		fileCmd.AddCommand(newAddTagsCmd())
		fileCmd.AddCommand(newListTagsCmd())
		fileCmd.AddCommand(newRemoveTagsCmd())
		fileCmd.AddCommand(newMergeCmd())
		fileCmd.AddCommand(newPatchCmd())
		fileCmd.AddCommand(newOpenapi2KongCmd())
		fileCmd.AddCommand(newFileRenderCmd())
		fileCmd.AddCommand(newKong2KicCmd())
	}
	return rootCmd
}

// Execute adds all child commands to the root command and sets
// flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	rootCmd := NewRootCmd()
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		// do not print error because cobra already prints it
		os.Exit(1)
	}
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
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	caCertContent := viper.GetString("ca-cert")

	if caCertContent == "" {
		caCertFileContent := viper.GetString("ca-cert-file")
		if caCertFileContent != "" {
			fileContent, err := ioutil.ReadFile(caCertFileContent)
			if err != nil {
				fmt.Printf("read file %q: %s", caCertFileContent, err)
				os.Exit(1)
			}
			caCertContent = string(fileContent)
			caCertContent = strings.TrimRight(caCertContent, "\n")
		}
	}

	rootConfig.Address = viper.GetString("kong-addr")
	rootConfig.TLSServerName = viper.GetString("tls-server-name")
	rootConfig.TLSSkipVerify = viper.GetBool("tls-skip-verify")
	rootConfig.TLSCACert = caCertContent
	rootConfig.Headers = extendHeaders(viper.GetStringSlice("headers"))
	rootConfig.SkipWorkspaceCrud = viper.GetBool("skip-workspace-crud")
	rootConfig.Debug = (viper.GetInt("verbose") >= 1)
	rootConfig.Timeout = (viper.GetInt("timeout"))

	clientCertContent := viper.GetString("tls-client-cert")

	if clientCertContent == "" {
		clientCertFileContent := viper.GetString("tls-client-cert-file")
		if clientCertFileContent != "" {
			fileContent, err := ioutil.ReadFile(clientCertFileContent)
			if err != nil {
				fmt.Printf("read file %q: %s", clientCertFileContent, err)
				os.Exit(1)
			}
			clientCertContent = string(fileContent)
			clientCertContent = strings.TrimRight(clientCertContent, "\n")
		}
	}
	rootConfig.TLSClientCert = clientCertContent

	clientKeyContent := viper.GetString("tls-client-key")

	if clientKeyContent == "" {
		clientKeyFileContent := viper.GetString("tls-client-key-file")
		if clientKeyFileContent != "" {
			fileContent, err := ioutil.ReadFile(clientKeyFileContent)
			if err != nil {
				fmt.Printf("read file %q: %s", clientKeyFileContent, err)
				os.Exit(1)
			}
			clientKeyContent = string(fileContent)
			clientKeyContent = strings.TrimRight(clientKeyContent, "\n")
		}
	}
	rootConfig.TLSClientKey = clientKeyContent

	if (rootConfig.TLSClientKey == "" && rootConfig.TLSClientCert != "") ||
		(rootConfig.TLSClientKey != "" && rootConfig.TLSClientCert == "") {
		fmt.Printf("tls-client-cert and tls-client-key / tls-client-cert-file and tls-client-key-file " +
			"must be used in conjunction but only one was provided")
		os.Exit(1)
	}

	// cookie-jar support
	rootConfig.CookieJarPath = viper.GetString("kong-cookie-jar-path")

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

	token := viper.GetString("konnect-token")
	tokenFile := viper.GetString("konnect-token-file")
	// read from token file only if token is not supplied using an
	// environment variable or flag
	if token == "" && tokenFile != "" {
		fileContent, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return fmt.Errorf("read file %q: %w", tokenFile, err)
		}
		if len(fileContent) == 0 {
			return fmt.Errorf("file %q: empty", tokenFile)
		}
		token = string(fileContent)
		token = strings.TrimRight(token, "\n")
	}

	disableAnalytics = !viper.GetBool("analytics")
	konnectConfig.Email = viper.GetString("konnect-email")
	konnectConfig.Password = password
	konnectConfig.Token = token
	konnectConfig.Debug = (viper.GetInt("verbose") >= 1)
	konnectConfig.Address = viper.GetString("konnect-addr")
	konnectConfig.Headers = extendHeaders(viper.GetStringSlice("headers"))
	konnectControlPlane = viper.GetString("konnect-control-plane-name")
	konnectRuntimeGroup = viper.GetString("konnect-runtime-group-name")
	return nil
}

func extendHeaders(headers []string) []string {
	userAgentHeader := fmt.Sprintf("User-Agent:decK/%s", VERSION)
	headers = append(headers, userAgentHeader)
	return headers
}

func init() {
	// set version and commit hash to report in the go-apiops library
	deckformat.ToolVersionSet("decK", VERSION, COMMIT)
}
