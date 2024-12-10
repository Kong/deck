package stats

import (
	"fmt"
	"os"
	"strings"

	"github.com/kong/go-database-reconciler/pkg/utils"
)

func WriteStatsToFile(buffer []byte, fileName string, fileFormat string) error {

	if fileName == "-" {
		if _, err := fmt.Print(string(buffer)); err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}
	} else {
		fileName = utils.AddExtToFilename(fileName, strings.ToLower(string(fileFormat)))
		if err := os.WriteFile(fileName, buffer, 0o600); err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
	}
	return nil
}
