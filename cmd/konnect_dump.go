package cmd

import (
	"fmt"
	"strings"

	"github.com/kong/deck/file"
	"github.com/kong/deck/utils"
	"github.com/spf13/cobra"
)

var (
	konnectDumpIncludeConsumers bool
	konnectDumpCmdKongStateFile string
	konnectDumpCmdStateFormat   string
	konnectDumpWithID           bool
)

// newKonnectDumpCmd represents the dump2 command
func newKonnectDumpCmd() *cobra.Command {
	konnectDumpCmd := &cobra.Command{
		Use:   "dump",
		Short: "Export configuration from Konnect (in alpha)",
		Long: `The konnect dump command reads all entities present in Konnect
	and writes them to a local file.

	The file can then be read using the 'deck konnect sync' command or 'deck konnect diff' command to
	configure Konnect.` + konnectAlphaState,
		Args: validateNoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			httpClient := utils.HTTPClient()
			_ = sendAnalytics("konnect-dump", "", modeKonnect)

			if yes, err := utils.ConfirmFileOverwrite(konnectDumpCmdKongStateFile, dumpCmdStateFormat, assumeYes); err != nil {
				return err
			} else if !yes {
				return nil
			}

			// get Konnect client
			konnectClient, err := utils.GetKonnectClient(httpClient, konnectConfig)
			if err != nil {
				return err
			}

			// authenticate with konnect
			_, err = konnectClient.Auth.Login(cmd.Context(),
				konnectConfig.Email,
				konnectConfig.Password)
			if err != nil {
				return fmt.Errorf("authenticating with Konnect: %w", err)
			}

			// get kong control plane ID
			kongCPID, err := fetchKongControlPlaneID(cmd.Context(), konnectClient)
			if err != nil {
				return err
			}

			// initialize kong client
			kongClient, err := utils.GetKongClient(utils.KongClientConfig{
				Address:    konnectConfig.Address + "/api/control_planes/" + kongCPID,
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

			return file.KonnectStateToFile(ks, file.WriteConfig{
				Filename:   konnectDumpCmdKongStateFile,
				FileFormat: file.Format(strings.ToUpper(konnectDumpCmdStateFormat)),
				WithID:     dumpWithID,
			})
		},
	}

	konnectDumpCmd.Flags().StringVarP(&konnectDumpCmdKongStateFile, "output-file", "o",
		"konnect", "file to which to write Kong's configuration.")
	konnectDumpCmd.Flags().StringVar(&konnectDumpCmdStateFormat, "format",
		"yaml", "output file format: json or yaml.")
	konnectDumpCmd.Flags().BoolVar(&konnectDumpWithID, "with-id",
		false, "write ID of all entities in the output.")
	konnectDumpCmd.Flags().BoolVar(&konnectDumpIncludeConsumers, "include-consumers",
		false, "export consumers, associated credentials and any plugins associated "+
			"with consumers.")
	konnectDumpCmd.Flags().BoolVar(&assumeYes, "yes",
		false, "Assume `yes` to prompts and run non-interactively.")
	return konnectDumpCmd
}
