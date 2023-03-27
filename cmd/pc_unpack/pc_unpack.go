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
	"io/fs"
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
				fmt.Printf("Failed to unpack dir %s: %s", inputName, err)
				failed = true
			}
		} else {
			fmt.Printf("%s, %d", outputName, len(args))
			if outputName == "" || len(args) > 1 {
				outputName = strings.TrimSuffix(inputName, filepath.Ext(inputName)) + ".png"
			}
			err := unpackTexture(inputName, outputName)
			if err != nil {
				fmt.Printf("Failed to unpack %s: %s", inputName, err)
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
			outputName = strings.TrimSuffix(path, filepath.Ext(path)) + ".png"
			err := unpackTexture(path, outputName)
			if err != nil {
				fail := true
				failed = &fail
			}
		}
		return nil
	}
}

func unpackTexture(inputName, outputName string) error {
	// file deepcode ignore PT: This is CLI tool, this is intended to be traversable
	file, err := os.Open(inputName)
	if err != nil {
		return fmt.Errorf("couldn't open file %s: %s", inputName, err)
	}

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("couldn't read image file config %s: %s", inputName, err)
	}
	if format != texture.FormatName {
		return nil
	}

	fmt.Printf("Unpacking %s: %dx%d\n", inputName, config.Width, config.Height)

	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("couldn't seek in image file %s: %s", inputName, err)
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("couldn't read image file %s: %s", inputName, err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("couldn't close image file %s: %s", inputName, err)
	}

	outputFile, err := os.Create(outputName)
	if err != nil {
		return fmt.Errorf("couldn't create output image file %s: %s", outputName, err)
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
		return fmt.Errorf("couldn't pack output image %s: %s", outputName, err)
	}

	err = outputFile.Close()
	if err != nil {
		return fmt.Errorf("couldn't close image file %s: %s", outputName, err)
	}
	return nil
}
