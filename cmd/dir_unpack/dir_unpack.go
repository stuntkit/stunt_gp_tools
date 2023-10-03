/*
Dir_unpack unpacks .dir/.wad files to directories
*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/stuntkit/stunt_gp_tools/pkg/dir"
)

// flags
var (
	outputDirectory string
)

func parseFlags() {
	pflag.StringVarP(&outputDirectory, "output", "o", "", "name of the output directory")
	pflag.Parse()
}

func usage() {
	fmt.Println("Unpacks dir archive used by Stunt GP and Worms Armageddon.")
	fmt.Println("Usage:\ndir_unpack input_file [input_files]")
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

	if outputDirectory != "" && len(args) > 1 {
		fmt.Println("Output name can only be used with one input directory")
		usage()
		os.Exit(1)
	}

	for _, inputName := range args {
		if outputDirectory == "" || len(args) > 1 {
			// I know, Worms uses .dir extension, but this is primarily Stunt GP program after all
			outputDirectory = strings.TrimSuffix(inputName, filepath.Ext(inputName))
		}

		dirFile, err := dir.FromDir(inputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't read Dir file: %s\n", err)
			os.Exit(1)
		}

		err = os.MkdirAll(outputDirectory, 0750)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't create %s path: %s\n", outputDirectory, err)
			os.Exit(1)
		}

		// deepcode ignore PT: this is a CLI tool
		err = saveFiles(dirFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't save files: %s\n", err)
			os.Exit(1)
		}
	}
}

func saveFiles(dirFile dir.Dir) error {
	for archFilename, archData := range dirFile {
		archFilename = strings.ReplaceAll(archFilename, "\\", "/")
		outPath := filepath.Join(outputDirectory, archFilename)

		err := os.MkdirAll(filepath.Dir(outPath), 0750)
		if err != nil {
			return fmt.Errorf("couldn't create %s path: %s", outPath, err)
		}

		// deepcode ignore PT: this is a CLI tool
		err = os.WriteFile(outPath, archData, 0666)
		if err != nil {
			return fmt.Errorf("couldn't save %s file: %s", outPath, err)
		}
	}
	return nil
}
