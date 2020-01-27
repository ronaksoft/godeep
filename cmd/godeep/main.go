package main

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/ronaksoft/godeep"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

var (
	AllPackages = godeep.InitPackages()
)

func main() {
	preRootCmd := &cobra.Command{
		Use: "GoDeep",
		Run: func(cmd *cobra.Command, args []string) {
			if v, _ := cmd.Flags().GetBool(FlagInteractive); v {
				p := prompt.New(executor, completer)
				p.Run()
			} else {
				RootCmd.SetArgs(args)
				_ = RootCmd.Execute()
			}
		},
	}
	preRootCmd.Flags().Bool(FlagInteractive, false, "enable interactive mode")
	_ = preRootCmd.Execute()
}

func executor(s string) {
	RootCmd.SetArgs(strings.Fields(s))
	_ = RootCmd.Execute()
}

func completer(d prompt.Document) []prompt.Suggest {
	suggests := make([]prompt.Suggest, 0, 10)
	cols := d.TextBeforeCursor()
	currCmd := RootCmd
	for _, col := range strings.Fields(cols) {
		for _, cmd := range currCmd.Commands() {
			if cmd.Name() == col {
				currCmd = cmd
				break
			}
		}
	}

	currWord := d.GetWordBeforeCursor()
	if strings.HasPrefix(currWord, "--") {
		// Search in Flags
		RootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if strings.HasPrefix(flag.Name, currWord[2:]) {
				suggests = append(suggests, prompt.Suggest{
					Text:        fmt.Sprintf("--%s", flag.Name),
					Description: flag.Usage,
				})
			}
		})
		currCmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if strings.HasPrefix(flag.Name, currWord[2:]) {
				suggests = append(suggests, prompt.Suggest{
					Text:        fmt.Sprintf("--%s", flag.Name),
					Description: flag.Usage,
				})
			}
		})

	} else {
		for _, cmd := range currCmd.Commands() {
			if strings.Contains(cmd.Name(), currWord) {
				suggests = append(suggests, prompt.Suggest{
					Text:        cmd.Name(),
					Description: cmd.Short,
				})
			}
		}
	}

	return suggests
}

var RootCmd = &cobra.Command{
	Use: "Root",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	fs := RootCmd.PersistentFlags()
	fs.Bool(FlagSkipStandardLib, false, "skip go standard packages")
	fs.Bool(FlagSkipVendor, false, "skip vendor packages")
	fs.String(FlagOutputDir, "./", "generated file will be stored here")
	fs.String(FlagInputDir, "./", "default place to look for files")

}
