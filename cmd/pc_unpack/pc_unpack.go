package main

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/halamix2/stunt_gp_tools/pkg/texture"
	"github.com/spf13/cobra"
)

// flags
var (
	outputName string
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
			outputName = strings.TrimSuffix(inputName, filepath.Ext(inputName)) + ".png"
		}

		fmt.Printf("Hi, today I'll unpack %s...\n", inputName)

		file, err := os.Open(inputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open file %s!\nReason: %s\n", inputName, err)
			os.Exit(3)
		}
		defer file.Close()

		config, format, err := image.DecodeConfig(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't read image file config %s!\nReason: %s\n", inputName, err)
			os.Exit(4)
		}
		fmt.Println("Width:", config.Width, "Height:", config.Height, "Format:", format)

		file.Seek(0, 0)

		img, _, err := image.Decode(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't read image file %s!\nReason: %s\n", inputName, err)
			os.Exit(4)
		}

		outputFile, err := os.Create(outputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't create output image file %s!\nReason: %s\n", outputName, err)
			os.Exit(4)
		}
		defer outputFile.Close()

		ext := filepath.Ext(strings.ToLower(outputName))
		switch ext {
		case ".jpg":
			fallthrough
		case ".jpeg":
			o := jpeg.Options{Quality: 90}
			err = jpeg.Encode(outputFile, img, &o)
		case ".png":
			err = png.Encode(outputFile, img)
		case ".pc":
			err = texture.Encode(outputFile, img)
		default:
			err = errors.New("unknown output format")
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't pack output image %s!\nReason: %s\n", outputName, err)
			os.Exit(5)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.MousetrapHelpText = ""
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.test_cobra.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&outputName, "output", "o", "", "name of the output file")
}
