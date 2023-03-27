// Package texture handles Stunt GP texture conversion
package texture

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"

	morton "github.com/gojuno/go.morton"
)

// A FormatError reports that the input is not a valid StuntGP texture.
type FormatError string

func (e FormatError) Error() string { return "Stunt GP texture: invalid format: " + string(e) }

type decoder struct {
	r       io.Reader
	version uint16
	width   uint16
	height  uint16
	// rgba     *image.RGBA
	// paletted *image.Paletted
	palette color.Palette
}

type imageRGBA struct {
	*image.RGBA
}

func setPixel(p []byte, c uint16) {

	p[0] = uint8(math.Round(float64(c>>10&31) * (255.0 / 31.0)))
	// p[1] = uint8((c >> 5 & 31) * (255 / 31))
	p[1] = uint8(math.Round(float64(c>>5&31) * (255.0 / 31.0)))
	p[2] = uint8(math.Round(float64(c&31) * (255.0 / 31.0)))
	p[3] = 0xFF
}

func (d *decoder) decodeImage() (imageRGBA, error) {
	im := image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
	unpacked := 0
	var word uint16
	word, _ = getWord(d.r)

	// unpack first pixel as-is
	p := im.Pix[unpacked*4 : 4*(1+unpacked)]
	setPixel(p, word)
	unpacked++

	for unpacked <= int(d.width)*int(d.height) {
		word, _ = getWord(d.r)

		for ((word >> 15) & 1) == 1 {
			// stream pixels
			p = im.Pix[unpacked*4 : 4*(1+unpacked)]
			setPixel(p, word)
			word, _ = getWord(d.r)
			unpacked++
		}

		if ((word >> 14) & 1) == 1 {
			// packed data
			repeats := int(((word >> 11) & 7) + 2)
			location := int(word & 0x7FF)

			for j := 0; j < repeats; j++ {
				p = im.Pix[4*(unpacked+j) : 4*(unpacked+j+1)]
				p2 := im.Pix[4*(unpacked-1-location+j) : 4*(unpacked-location+j)]
				copy(p, p2)
			}
			unpacked += repeats

		} else if ((word >> 15) & 1) == 0 {
			// end reading
			if word == 0 {
				return imageRGBA{RGBA: im}, nil
			}
			// stream transparency
			unpacked += int(word)
		}
	}
	return imageRGBA{}, FormatError("Stunt GP texture: too many bytes")
}

func (im imageRGBA) swapRedBlue() {
	width := im.Rect.Max.X - im.Rect.Min.X
	height := im.Rect.Max.Y - im.Rect.Min.Y
	for i := 0; i < int(width)*int(height); i++ {
		r := im.Pix[4*i : (4*i)+1]
		b := im.Pix[(4*i)+2 : (4*i)+3]
		tmp := make([]byte, len(b))
		copy(tmp, b)
		copy(b, r)
		copy(r, tmp)
	}
}

func (im imageRGBA) deswizzle() {
	// Dreamcast, untwiddle
	width := im.Rect.Max.X - im.Rect.Min.X
	height := im.Rect.Max.Y - im.Rect.Min.Y
	rgba2 := image.NewRGBA(image.Rect(0, 0, width, height))
	// TODO check if image is square and size is power of two
	m := morton.Make64(2, uint64(math.Log2(float64(width))))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// DZ uzez more of a n-curve than z-curve, so we need to swap x and y axes
			i := m.Pack(uint64(y), uint64(x))
			untwiddle(im.RGBA, rgba2, x, y, i)
		}
	}
	copy(im.Pix, rgba2.Pix)
}

func untwiddle(rgba, rgba2 *image.RGBA, x, y int, i int64) {
	pos := (y * rgba.Bounds().Dy()) + x
	p := rgba2.Pix[pos*4 : 4*(pos+1)]
	p2 := rgba.Pix[4*i : 4*(i+1)]
	copy(p, p2)
}

// Decode reads a Stunt GP texture image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	var d decoder
	var err error
	d.r = r
	if err := d.decodeConfig(); err != nil {
		return nil, err
	}

	switch d.version {
	case FPaletted:
		d.palette, err = d.readColorPalette()
		if err != nil {
			return nil, err
		}
		return d.decodePalleted()
	case FPC:
		return d.decodeImage()
	case FDreamcast:
		rgbaImage, err := d.decodeImage()
		if err != nil {
			return nil, err
		}
		rgbaImage.deswizzle()
		return rgbaImage, nil
	case FPSTwo:
		rgbaImage, err := d.decodeImage()
		if err != nil {
			return nil, err
		}
		rgbaImage.swapRedBlue()
		return rgbaImage, nil
	default:
		return nil, FormatError(fmt.Sprintf("Unknown file format version! This program supports versions 2-5, but got version %d instead\n", d.version))
	}
}

func (d *decoder) decodeConfig() error {
	var err error
	buf := make([]byte, len(textureHeader)+6)
	if _, err = io.ReadFull(d.r, buf); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	if string(buf[:len(textureHeader)]) != textureHeader {
		return FormatError("Stunt GP texture: invalid format")
	}

	d.version = binary.LittleEndian.Uint16(buf[4:6])
	d.width = binary.LittleEndian.Uint16(buf[6:8])
	d.height = binary.LittleEndian.Uint16(buf[8:10])

	if d.width == 0 || d.height == 0 {
		return FormatError(fmt.Sprintf("Stunt GP texture: unsupported size: %dx%d\n", d.width, d.height))
	}

	if d.version == FPaletted {
		// read number of colors
		if d.palette, err = d.readColorPalette(); err != nil {
			return err
		}
	}

	return nil
}

func (d *decoder) readColorPalette() (color.Palette, error) {
	colorCount, err := getWord(d.r)
	if err != nil {
		return nil, err
	}
	palette := make([]color.Color, colorCount)
	// ARGB8888
	buf := make([]byte, colorCount*4)
	for i := 0; i < int(colorCount); i++ {

		currentColor := color.RGBA{
			A: buf[4*i],
			R: buf[(4*i)+1],
			G: buf[(4*i)+2],
			B: buf[(4*i)+3],
		}
		palette[i] = currentColor
	}

	return palette, nil
}

func (d *decoder) decodePalleted() (*image.Paletted, error) {
	palettedImage := image.NewPaletted(image.Rect(0, 0, int(d.width), int(d.height)), d.palette)
	pixels := palettedImage.Pix[0 : int(d.width)*int(d.height)]
	_, err := io.ReadFull(d.r, pixels)
	return palettedImage, err
}

// DecodeConfig returns the color model and dimensions of a Stunt GP texture without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	var d decoder
	d.r = r
	if err := d.decodeConfig(); err != nil {
		return image.Config{}, err
	}
	colorModel := color.RGBAModel

	if d.version == FPaletted {
		colorModel = d.palette
	}

	return image.Config{
		ColorModel: colorModel,
		Width:      int(d.width),
		Height:     int(d.height),
	}, nil
}

func init() {
	image.RegisterFormat(FormatName, textureHeader, Decode, DecodeConfig)
}
