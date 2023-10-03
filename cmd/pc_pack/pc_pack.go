/*
Pc_pack converts some popular image formats to Stunt GP textures.

This program accepts jpg or png files.
It can output PC, DC and PS2 textures (except paletted ones).
*/
package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
	"github.com/stuntkit/stunt_gp_tools/pkg/texture"
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
	fmt.Println("Flags:")
	pflag.PrintDefaults()
}

func main() {
	failed := false
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
		f, err := os.Stat(inputName)
		if err != nil {
			fmt.Printf("Failed to get info about %s: %s", inputName, err)
			failed = true
		}

		if f.IsDir() {
			walkFunc := getWalkFunc(&failed)
			err := filepath.Walk(inputName, walkFunc)
			if err != nil {
				fmt.Printf("Failed to pack dir %s: %s", inputName, err)
				failed = true
			}
		} else {

			if outputName == "" || len(args) > 1 {
				outputName = strings.TrimSuffix(inputName, filepath.Ext(inputName)) + ".pc"
			}

			err := packTexture(inputName, outputName)
			if err != nil {
				fmt.Printf("Failed to pack %s: %s", inputName, err)
				failed = true
			}

		}
	}
	if failed {
		os.Exit(1)
	}
}

func getWalkFunc(failed *bool) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			outputName = strings.TrimSuffix(path, filepath.Ext(path)) + ".pc"
			err := packTexture(path, outputName)
			if err != nil {
				fail := true
				failed = &fail
			}
		}
		return nil
	}
}

func packTexture(inputName, outputName string) error {
	// file deepcode ignore PT: This is CLI tool, this is intended to be traversable
	file, err := os.Open(inputName)
	if err != nil {
		return fmt.Errorf("couldn't open file %s", inputName)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("couldn't read image file %s", inputName)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("couldn't close image file %s: %s", inputName, err)
	}

	outputFile, err := os.Create(outputName)
	if err != nil {
		return fmt.Errorf("couldn't create output image file %s", outputName)
	}

	textureFormat := texture.FPC
	if dreamcast {
		textureFormat = texture.FDreamcast
	}
	fmt.Printf("packing %s...\n", outputName)
	enc := texture.Encoder{Compress: !uncompressed, Format: textureFormat}
	err = enc.Encode(outputFile, img)
	if err != nil {
		return fmt.Errorf("couldn't pack output image %s", outputName)
	}

	err = outputFile.Close()
	if err != nil {
		return fmt.Errorf("couldn't close image file %s: %s", outputName, err)
	}

	return nil
}
