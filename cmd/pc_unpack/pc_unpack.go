/*
Pc_unpack converts Stunt GP textures to one of the more popular formats.

This program accepts PC, DC and PS2 textures.
It can output jpg or png files.
*/
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
	"github.com/spf13/pflag"
)

// flags
var (
	outputName string
)

func parseFlags() {
	pflag.StringVarP(&outputName, "output", "o", "", "name of the output file")
	pflag.Parse()
}

func usage() {
	fmt.Println("Unpacks image from texture format used by Stunt GP")
	// TODO
	//fmt.Println("Usage:\npc_unpack ")
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

	for _, inputName := range args {

		if outputName == "" || len(args) > 1 {
			outputName = strings.TrimSuffix(inputName, filepath.Ext(inputName)) + ".png"
		}

		fmt.Printf("Hi, today I'll unpack %s...\n", inputName)

		// deepcode ignore PT: this is a CLI tool
		file, err := os.Open(inputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open file %s!\nReason: %s\n", inputName, err)
			os.Exit(3)
		}

		config, format, err := image.DecodeConfig(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't read image file config %s!\nReason: %s\n", inputName, err)
			os.Exit(4)
		}
		fmt.Println("Width:", config.Width, "Height:", config.Height, "Format:", format)

		_, err = file.Seek(0, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't seek in image file %s: %s\n", inputName, err)
			os.Exit(4)
		}

		img, _, err := image.Decode(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't read image file %s!\nReason: %s\n", inputName, err)
			os.Exit(4)
		}

		err = file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't close image file %s: %s\n", inputName, err)
			os.Exit(4)
		}

		outputFile, err := os.Create(outputName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't create output image file %s!\nReason: %s\n", outputName, err)
			os.Exit(4)
		}

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
		case ".dc":
			encoder := texture.Encoder{
				Compress: true,
				Format:   texture.FDreamcast,
			}
			err = encoder.Encode(outputFile, img)
		default:
			err = errors.New("unknown output format")
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't pack output image %s!\nReason: %s\n", outputName, err)
			os.Exit(5)
		}

		err = outputFile.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't close image file %s: %s\n", outputName, err)
			os.Exit(4)
		}
	}
}
