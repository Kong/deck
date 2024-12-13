package cmd

import (
	"fmt"
	"log"

	"github.com/kong/deck/stats"
	"github.com/kong/go-apiops/logbasics"
	"github.com/kong/go-database-reconciler/pkg/file"
	"github.com/spf13/cobra"
)

var (
	cmdStatsInputFilename  string
	cmdStatsOutputFilename string
	cmdStatsOutputFormat   string
	cmdStatsStyle          string
	cmdStatsIncludeTags    bool
	cmdStatsSelectorTags   []string
)

func executeStats(cmd *cobra.Command, _ []string) error {
	var (
		outputContent *file.Content
		err           error
		// outputFileFormat file.Format
	)

	verbosity, _ := cmd.Flags().GetInt("verbose")
	logbasics.Initialize(log.LstdFlags, verbosity)

	validFormats := map[string]bool{
		string(stats.TextFormat):     true,
		string(stats.CSVFormat):      true,
		string(stats.HTMLFormat):     true,
		string(stats.MarkdownFormat): true,
	}

	if !validFormats[cmdStatsOutputFormat] {
		return fmt.Errorf("invalid output format '%s'; must be one of: text, csv, html, markdown", cmdStatsOutputFormat)
	}

	inputContent, err := file.GetContentFromFiles([]string{cmdStatsInputFilename}, false)
	if err != nil {
		return fmt.Errorf("failed reding input file '%s'; %w", cmdStatsInputFilename, err)
	}

	outputContent = inputContent.DeepCopy()

	buffer, err := stats.PrintContentStatistics(outputContent, cmdStatsStyle,
		cmdStatsOutputFormat, cmdStatsIncludeTags, cmdStatsSelectorTags)
	if err != nil {
		return fmt.Errorf("failed writing output file '%s'; %w", cmdStatsOutputFilename, err)
	}

	err = stats.WriteStatsToFile(buffer, cmdStatsOutputFilename, cmdStatsOutputFormat)

	if err != nil {
		return fmt.Errorf("failed generating stats '%s'; %w", cmdStatsInputFilename, err)
	}

	return nil
}

func Format(cmdStatsOutputFormat string) {
	panic("unimplemented")
}

//
// Define the CLI data for the Stats command
//

func newStatsCmd() *cobra.Command {
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Obtain entity count and statistics from a decK file",
		Long: `Obtain entity count and statistics from a decK file.
		
This command calculates 
- Number of entities and the percentage of each type in a decK file. 
- Number of instances for each plugin.
- Optionally, entity counts by tag

Output in text, csv, html or markdown.`,
		RunE: executeStats,
		Args: cobra.NoArgs,
	}

	statsCmd.Flags().StringVarP(&cmdStatsInputFilename, "state", "s", "-",
		"decK file to process. Use - to read from stdin.")
	statsCmd.Flags().StringVarP(&cmdStatsOutputFilename, "output-file", "o", "-",
		"Output file to write. Use - to write to stdout.")
	statsCmd.Flags().StringVarP(&cmdStatsOutputFormat, "render", "r", "txt",
		"Render as txt, csv, html or md. Default is to render as txt.")
	statsCmd.Flags().StringSliceVar(&cmdStatsSelectorTags, "select-tag", []string{},
		"only entities matching specified tags are shown.\n"+
			"When this setting has multiple tag values, entities must match every tag. Example --select-tag=\"tag1,tag2\"")
	statsCmd.Flags().BoolVar(&cmdStatsIncludeTags, "with-tags", false,
		"Include entity count by tag.")
	statsCmd.Flags().StringVarP(&cmdStatsStyle, "style", "t", "StyleDefault",
		"Style to use when rendering. See https://github.com/jedib0t/go-pretty/blob/main/table/style.go.")

	return statsCmd
}
