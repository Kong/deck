package cmd

import (
	"strings"

	"github.com/kong/deck/konnect"

	"github.com/kong/deck/file"
	"github.com/kong/deck/utils"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	konnectDumpIncludeConsumers bool
	konnectDumpCmdKongStateFile string
	konnectDumpCmdStateFormat   string
	konnectDumpWithID           bool
)

// konnectDumpCmd represents the dump2 command
var konnectDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Export configuration from Konnect (in alpha)",
	Long: `Dump command reads all entities present in Konnect and exports them to
a file on disk. The file can then be read using the Sync or Diff command to again
configure Konnect.` + konnectAlphaState,
	Args: validateNoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		httpClient := utils.HTTPClient()

		if yes, err := confirmFileOverwrite(konnectDumpCmdKongStateFile, dumpCmdStateFormat, assumeYes); err != nil {
			return err
		} else if !yes {
			return nil
		}

		// get Konnect client
		konnectClient, err := utils.GetKonnectClient(httpClient, konnectConfig.Debug)
		if err != nil {
			return err
		}

		// authenticate with konnect
		_, err = konnectClient.Auth.Login(cmd.Context(),
			konnectConfig.Email,
			konnectConfig.Password)
		if err != nil {
			return errors.Wrap(err, "authenticating with Konnect")
		}

		// get kong control plane ID
		kongCPID, err := fetchKongControlPlaneID(cmd.Context(), konnectClient)
		if err != nil {
			return err
		}

		// initialize kong client
		kongClient, err := utils.GetKongClient(utils.KongClientConfig{
			Address:    konnect.BaseURL() + "/api/control_planes/" + kongCPID,
			HTTPClient: httpClient,
			Debug:      konnectConfig.Debug,
		})
		if err != nil {
			return err
		}

		ks, err := getKonnectState(cmd.Context(), kongClient, konnectClient, kongCPID,
			!konnectDumpIncludeConsumers)
		if err != nil {
			return err
		}

		if err := file.KonnectStateToFile(ks, file.WriteConfig{
			Filename:   konnectDumpCmdKongStateFile,
			FileFormat: file.Format(strings.ToUpper(konnectDumpCmdStateFormat)),
			WithID:     dumpWithID,
		}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	konnectCmd.AddCommand(konnectDumpCmd)
	konnectDumpCmd.Flags().StringVarP(&konnectDumpCmdKongStateFile, "output-file", "o",
		"konnect", "file to which to write Kong's configuration."+
			"Use '-' to write to stdout.")
	konnectDumpCmd.Flags().StringVar(&konnectDumpCmdStateFormat, "format",
		"yaml", "output file format: json or yaml")
	konnectDumpCmd.Flags().BoolVar(&konnectDumpWithID, "with-id",
		false, "write ID of all entities in the output")
	konnectDumpCmd.Flags().BoolVar(&konnectDumpIncludeConsumers, "include-consumers",
		false, "export consumers, associated credentials and any plugins associated "+
			"with consumers")
	konnectDumpCmd.Flags().BoolVar(&assumeYes, "yes",
		false, "Assume 'yes' to prompts and run non-interactively")
}
