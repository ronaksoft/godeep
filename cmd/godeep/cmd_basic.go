package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/ronaksoft/godeep"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

/*
   Creation Time: 2019 - Oct - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func init() {
	RootCmd.AddCommand(CmdAnalyze, CmdPrint, CmdImport, CmdExport)
}

func ResetCommands() {
	CmdPrint.ResetCommands()
	AllPackages.ForEach(func(pkgPath string, pkg *godeep.Package) {
		CmdPrint.AddCommand(&cobra.Command{
			Use: pkgPath,
			Run: func(cmd *cobra.Command, args []string) {
				pkg.Print()
			},
		})
	})

}

var CmdImport = &cobra.Command{
	Use:   "import",
	Short: "import is used to use already exported data from analyze",
	Run: func(cmd *cobra.Command, args []string) {
		inputDir, err := cmd.Flags().GetString(FlagInputDir)
		PrintOnErr(err)

		f, err := os.Open(filepath.Join(inputDir, "all_packages.json"))
		PanicOnErr(err)
		data, err := ioutil.ReadAll(f)
		PanicOnErr(err)
		err = AllPackages.Unmarshal(data)
		PanicOnErr(err)
		err = f.Close()
		PanicOnErr(err)
		ResetCommands()
	},
}

var CmdExport = &cobra.Command{
	Use:   "export",
	Short: "exports the analyzed data as a json file",
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, err := cmd.Flags().GetString(FlagOutputDir)
		PrintOnErr(err)

		exportData := AllPackages.Marshal()
		f, err := os.Create(filepath.Join(outputDir, "all_packages.json"))
		PanicOnErr(err)
		_, err = f.Write(exportData)
		PanicOnErr(err)
		err = f.Close()
		PanicOnErr(err)
	},
}

var CmdAnalyze = &cobra.Command{
	Use: "analyze",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please be patient, this may take a bit longer than you think ...")
		cwd, _ := os.Getwd()
		err := godeep.FindPackages(AllPackages, cwd, func(path string) {
			fmt.Println(fmt.Sprintf("Package '%s' %s",
				color.WhiteString("%s", path),
				color.GreenString("analyzed"),
			))
		})
		PanicOnErr(err)
		color.HiGreen("All Packages have been traversed, now we are building the relation")
		ResetCommands()
	},
}

var CmdPrint = &cobra.Command{
	Use: "print",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			pkg := AllPackages.GetByPath(args[0])
			if pkg != nil {
				pkg.Print()
			} else {
				fmt.Println("Package not found:", args[0])
			}
		} else {
			AllPackages.ForEach(func(pkgPath string, pkg *godeep.Package) {
				pkg.Print()
			})
		}

	},
}
