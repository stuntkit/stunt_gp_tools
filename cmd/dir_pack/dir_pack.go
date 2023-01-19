/*
Dir_pack packs directories to .dir/.wad files
*/
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/halamix2/stunt_gp_tools/pkg/dir"
	_ "github.com/halamix2/stunt_gp_tools/pkg/dir"
	"github.com/spf13/pflag"
)

// flags
var (
	outputName string
	// verbose    bool
)

func parseFlags() {
	pflag.StringVarP(&outputName, "output", "o", "", "name of the output file")
	// verbose
	pflag.Parse()
}

func usage() {
	fmt.Println("Packs dir archive used by Stunt GP and Worms Armageddon.")
	fmt.Println("Usage:\ndir_pack input_dir [input_dirs]")
	fmt.Println("Flags:")
	pflag.PrintDefaults()
}

func main() {
	parseFlags()
	args := pflag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	if outputName != "" && len(args) > 1 {
		fmt.Println("Output name can only be used with one input directory")
		usage()
		os.Exit(1)
	}

	for _, inputName := range args {
		if outputName == "" {
			// I know, Worms uses .dir extension, but this is primarily Stunt GP program after all
			outputName = strings.TrimSuffix(inputName, filepath.Ext(inputName)) + ".wad"
		}

		dir := dir.Dir{}

		walkFunc := getWalkFunc(&dir)
		err := filepath.Walk(inputName, walkFunc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't get list of files: %s\n", outputName)
			os.Exit(1)
		}
		fmt.Printf("packing %d files\n", len(dir))

		outputFile, err := os.Create(outputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't create output image file %s!\n", outputName)
			os.Exit(1)
		}

		err = dir.ToWriter(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't save dir file %s: %s\n", outputName, err)
			os.Exit(1)
		}

		err = outputFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't close output file %s: %s\n", outputName, err)
			os.Exit(1)
		}
	}
}

func getWalkFunc(dir *dir.Dir) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			// file
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			pathNormalized := strings.ReplaceAll(path, "/", "\\")
			split := strings.Split(pathNormalized, "\\")
			pathNormalized = strings.Join(split[1:], "\\")
			(*dir)[pathNormalized] = data
		}
		return nil
	}
}
