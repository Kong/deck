package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kong/deck/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func genMarkdownTree(cmd *cobra.Command, dir string) error {
	identity := func(s string) string { return s }
	emptyStr := func(_ string) string { return "" }
	return genMarkdownTreeCustom(cmd, dir, emptyStr, identity)
}

func genMarkdownTreeCustom(cmd *cobra.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := genMarkdownTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}
	return genMarkdownCustom(cmd, f, linkHandler)
}

func flagUsagesWrapped(f *pflag.FlagSet) string {
	buf := new(bytes.Buffer)

	lines := make([]string, 0)

	maxlen := 0
	f.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		var line string
		if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
			line = fmt.Sprintf("`-%s`, `--%s`\n:", flag.Shorthand, flag.Name)
		} else {
			line = fmt.Sprintf("`--%s`\n:", flag.Name)
		}
		usage := flag.Usage
		if flag.NoOptDefVal != "" {
			switch flag.Value.Type() {
			case "string":
				line += fmt.Sprintf("[=\"%s\"]", flag.NoOptDefVal)
			case "bool":
				if flag.NoOptDefVal != "true" {
					line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			case "count":
				if flag.NoOptDefVal != "+1" {
					line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
				}
			default:
				line += fmt.Sprintf("[=%s]", flag.NoOptDefVal)
			}
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += usage
		if flag.DefValue != "" && flag.DefValue != "[]" {
			if flag.Value.Type() == "string" {
				line += fmt.Sprintf(" (Default: `%q`)", flag.DefValue)
			} else {
				line += fmt.Sprintf(" (Default: `%s`)", flag.DefValue)
			}
			if len(flag.Deprecated) != 0 {
				line += fmt.Sprintf(" (DEPRECATED: %s)", flag.Deprecated)
			}
		}
		line += "\n"

		lines = append(lines, line)
	})

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		fmt.Fprintln(buf, line[:sidx], "", line[sidx+1:])
	}

	return buf.String()
}

func printFlags(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("## Flags\n\n")
		buf.WriteString(flagUsagesWrapped(flags))
		buf.WriteString("\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("## Global flags\n\n")
		buf.WriteString(flagUsagesWrapped(parentFlags))
		buf.WriteString("\n\n")
	}
	return nil
}

func genMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	buf.WriteString(fmt.Sprintf("---\ntitle: %s\nsource_url: https://github.com/Kong/deck/tree/main/cmd\n---\n\n", name))
	buf.WriteString(cmd.Long + "\n\n")

	if cmd.Runnable() {
		buf.WriteString("## Syntax\n\n")
		buf.WriteString(
			fmt.Sprintf("```\n%s [command-specific flags] [global flags]\n```\n\n", name),
		)
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("## Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if err := printFlags(buf, cmd); err != nil {
		return err
	}

	if hasSeeAlso(cmd) {
		buf.WriteString("## See also\n\n")
		prefix := "/deck/{{page.kong_version}}/reference/"
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			link := strings.Replace(pname, " ", "_", -1)
			buf.WriteString(
				fmt.Sprintf("* [%s](%s%s)\t - %s\n", pname, prefix, linkHandler(link), parent.Short),
			)
			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			link := strings.Replace(cname, " ", "_", -1)
			buf.WriteString(
				fmt.Sprintf("* [%s](%s%s)\t - %s\n", cname, prefix, linkHandler(link), child.Short),
			)
		}
		buf.WriteString("\n")
	}
	_, err := buf.WriteTo(w)
	return err
}

func main() {
	var outputPath string
	flag.StringVar(&outputPath, "output-path", ".", "path to output directory")
	flag.Parse()

	err := genMarkdownTree(cmd.NewRootCmd(), outputPath)
	if err != nil {
		log.Fatal(err)
	}
}
