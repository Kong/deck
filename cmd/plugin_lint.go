package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kong/deck/plugin/lua"
	"github.com/spf13/cobra"

)

var (
	pluginLintCode		string
	pluginLintEdition	string
	pluginLintSandbox	string
)

// Executes plugin lint command. 
func executePluginLint(cmd *cobra.Command, args []string) error{
	var luaCode string
	var err error

	if pluginLintCode == "-" {
		luaCode, err = readFromStdin()
	} else {
		content, readErr := os.ReadFile(pluginLintCode)
		if readErr != nil {
			return fmt.Errorf("failed to read file %s: %w", pluginLintCode, readErr)
		}
		luaCode = string(content)

	}

	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	if strings.TrimSpace(luaCode) == "" {
		return errors.New("no Lua code provided. Use --code or pipe code to stdin")
	}


	v, err := lua.NewValidator(pluginLintEdition, "")
	if err != nil {
		return err
	}

	violations, err := v.Validate(luaCode, pluginLintSandbox)
	if err != nil {
		return err
	}

	if len(violations) == 0 {
		fmt.Println("Success: No violations found. Your Lua code is safe for the specified sandbox.")
		return nil
	}

	fmt.Printf("Found %d violations:\n", len(violations))
	for _, vio := range violations {
		lineInfo := ""
		if vio.Line > 0 {
			lineInfo = fmt.Sprintf(" (line %d)", vio.Line)
		}
		fmt.Printf("  - [%s] %s: %s%s\n", vio.Severity, vio.ID, vio.Message, lineInfo)
	}


	return errors.New("lua validation failed")



}
func readFromStdin() (string, error) {
   data, err := io.ReadAll(os.Stdin)
	 if err != nil {
		 return "", fmt.Errorf("error reading from stdin: %w", err)
	 }

   return string(data), err
}


func newPluginLintCmd() *cobra.Command {
	pluginLintCmd := &cobra.Command{
		Use: "lint [flags]",
		Short: "Check custom LUA code for security and sandbox compatibility",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executePluginLint(cmd, args)
		},
	}
	
	pluginLintCmd.Flags().StringVarP(&pluginLintCode, "code", "c", "-", "custom LUA code to validate. Use - to read from stdin.")
	pluginLintCmd.Flags().StringVarP(&pluginLintEdition, "edition", "e", "ee", "Kong Edition [choices: ee (enterprise), oss (open source).")
	pluginLintCmd.Flags().StringVarP(&pluginLintSandbox, "sandbox", "s", "strict", "Kong Edition [choices: ee (enterprise), oss (open source).")
	
	return pluginLintCmd
}
