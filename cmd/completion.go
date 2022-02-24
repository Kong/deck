package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(deck completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ deck completion bash > /etc/bash_completion.d/deck
  # macOS:
  $ deck completion bash > /usr/local/etc/bash_completion.d/deck

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ deck completion zsh > "${fpath[1]}/_yourprogram"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ deck completion fish | source

  # To load completions for each session, execute once:
  $ deck completion fish > ~/.config/fish/completions/deck.fish

PowerShell:

  PS> deck completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> deck completion powershell > deck.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("invalid shell: %q", args[0])
			}
		},
	}
}
