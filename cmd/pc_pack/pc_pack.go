package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/halamix2/stunt_gp_tools/pkg/texture"
	"github.com/spf13/cobra"
)

// flags
var (
	outputName              string
	dreamcast, uncompressed bool
)

var rootCmd = &cobra.Command{
	Use:   "pack",
	Short: "Packs image to texture format used by Stunt GP",
	Long: `Packs image to texture format used by Stunt GP
	It can pack to either PC or Dreamcast versions of the format
	or make uncompressed files if the need arises.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputName := args[0]
		if outputName == "" {
			outputName = strings.TrimSuffix(inputName, filepath.Ext(inputName)) + ".pc"
		}

		fmt.Printf("Hi, today I'll pack %s...\n", inputName)

		file, err := os.Open(inputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open file %s!\n", inputName)
			os.Exit(3)
		}
		defer file.Close()

		img, _, err := image.Decode(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't read image file %s!\n", inputName)
			os.Exit(4)
		}

		outputFile, err := os.Create(outputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't create output image file %s!\n", outputName)
			os.Exit(4)
		}
		defer outputFile.Close()

		fmt.Printf("packing %s...\n", outputName)
		enc := texture.Encoder{Compress: !uncompressed, Dreamcast: dreamcast}
		err = enc.Encode(outputFile, img)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't pack output image %s!\n", outputName)
			os.Exit(5)
		}
		fmt.Println("done")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test_cobra.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&outputName, "output", "o", "", "name of the output file")
	rootCmd.Flags().BoolVarP(&dreamcast, "dreamcast", "d", false, "Packs texture for use in Dreamcast version ")
	rootCmd.Flags().BoolVarP(&uncompressed, "uncompressed", "u", false, "Do not compress texture")
}
