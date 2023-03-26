/*
Pc_pack converts some popular image formats to Stunt GP textures.

This program acceptsjpg or png files.
It can output PC, DC and PS2 textures (except paletted ones).
*/
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
	"github.com/spf13/pflag"
)

// flags
var (
	outputName              string
	dreamcast, uncompressed bool
)

func parseFlags() {
	pflag.StringVarP(&outputName, "output", "o", "", "name of the output file")
	pflag.BoolVarP(&dreamcast, "dreamcast", "d", false, "Packs texture for use in Dreamcast version ")
	pflag.BoolVarP(&uncompressed, "uncompressed", "u", false, "Do not compress texture")
	pflag.Parse()
}

func usage() {
	fmt.Println("Packs image to texture format used by Stunt GP")
	// TODO
	//fmt.Println("Usage:\npc_pack ")
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
		fmt.Println("Output name can only be used with one input file")
		usage()
		os.Exit(1)
	}

	for _, inputName := range args {
		if outputName == "" || len(args) > 1 {
			outputName = strings.TrimSuffix(inputName, filepath.Ext(inputName)) + ".pc"
		}

		fmt.Printf("Hi, today I'll pack %s...\n", inputName)

		// deepcode ignore PT: this is a CLI tool
		file, err := os.Open(inputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open file %s!\n", inputName)
			os.Exit(3)
		}

		img, _, err := image.Decode(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't read image file %s!\n", inputName)
			os.Exit(4)
		}

		err = file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't close image file %s: %s\n", inputName, err)
			os.Exit(4)
		}

		outputFile, err := os.Create(outputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't create output image file %s!\n", outputName)
			os.Exit(4)
		}

		textureFormat := texture.FPC
		if dreamcast {
			textureFormat = texture.FDreamcast
		}
		fmt.Printf("packing %s...\n", outputName)
		enc := texture.Encoder{Compress: !uncompressed, Format: textureFormat}
		err = enc.Encode(outputFile, img)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't pack output image %s!\n", outputName)
			os.Exit(5)
		}

		err = outputFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't close image file %s: %s\n", outputName, err)
			os.Exit(4)
		}
	}
}
