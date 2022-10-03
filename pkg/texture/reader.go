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

const pcHeader = "TM!\x1a"

// A FormatError reports that the input is not a valid StuntGP texture.
type FormatError string

func (e FormatError) Error() string { return "Stunt GP texture: invalid format: " + string(e) }

type decoder struct {
	r        io.Reader
	version  uint16
	width    uint16
	height   uint16
	rgba     *image.RGBA
	paletted *image.Paletted
	palette  []color.Color
}

func setPixel(p []byte, c uint16) {

	p[0] = uint8(math.Round(float64(c>>10&31) * (255.0 / 31.0)))
	// p[1] = uint8((c >> 5 & 31) * (255 / 31))
	p[1] = uint8(math.Round(float64(c>>5&31) * (255.0 / 31.0)))
	p[2] = uint8(math.Round(float64(c&31) * (255.0 / 31.0)))
	p[3] = 0xFF
}

func getWord(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf), nil
}

func countIndex(i int, width uint16) int {
	return int(width)*(i%int(width)) + int(i/int(width))
}

func untwiddle(rgba, rgba2 *image.RGBA, x, y uint16, i int64) {
	pos := (int(y) * rgba.Bounds().Dy()) + int(x)
	p := rgba2.Pix[pos*4 : 4*(pos+1)]
	p2 := rgba.Pix[4*i : 4*(i+1)]
	copy(p, p2)
}

func (d decoder) readPalette() error {
	for i := 0; i < len(d.palette); i++ {
		colour := color.RGBA{}
		tmp := make([]byte, 4)
		d.r.Read(tmp)
		colour.R = tmp[0]
		colour.G = tmp[1]
		colour.B = tmp[2]
		colour.A = tmp[3]
		d.palette[i] = colour
	}
	return nil
}

func (d decoder) decodePalleted() error {
	pixels := d.paletted.Pix[0 : int(d.width)*int(d.height)]
	io.ReadFull(d.r, pixels)
	return nil
}

func (d decoder) decode() error {
	unpacked := 0
	var word uint16
	word, _ = getWord(d.r)

	// unpack first pixel as-is
	var p []byte = d.rgba.Pix[unpacked*4 : 4*(1+unpacked)]
	setPixel(p, word)
	unpacked++

	for unpacked <= int(d.width)*int(d.height) {
		word, _ = getWord(d.r)

		for ((word >> 15) & 1) == 1 {
			// stream pixels
			p = d.rgba.Pix[unpacked*4 : 4*(1+unpacked)]
			setPixel(p, word)
			word, _ = getWord(d.r)
			unpacked++
		}

		if ((word >> 14) & 1) == 1 {
			// packed data
			repeats := int(((word >> 11) & 7) + 2)
			location := int(word & 0x7FF)

			for j := 0; j < repeats; j++ {
				p = d.rgba.Pix[4*(unpacked+j) : 4*(unpacked+j+1)]
				p2 := d.rgba.Pix[4*(unpacked-1-location+j) : 4*(unpacked-location+j)]
				copy(p, p2)
			}
			unpacked += repeats

		} else {
			if ((word >> 15) & 1) == 0 {
				// end reading
				if word == 0 {
					return nil
				}
				// stream transparency
				unpacked += int(word)
			}
		}
	}
	return FormatError("pc: too many bytes")
}

func (d *decoder) swapRedBlue() {
	for i := 0; i < int(d.height)*int(d.width); i++ {
		r := d.rgba.Pix[4*i : (4*i)+1]
		b := d.rgba.Pix[(4*i)+2 : (4*i)+3]
		tmp := make([]byte, len(b))
		copy(tmp, b)
		copy(b, r)
		copy(r, tmp)
	}
}

func (d decoder) deswizzle() {
	// Dreamcast, untwiddle
	rgba2 := image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
	// TODO check if image is square and size is power of two
	m := morton.Make64(2, uint64(math.Log2(float64(d.width))))

	for y := uint16(0); y < d.height; y++ {
		for x := uint16(0); x < d.width; x++ {
			// DZ uzez more of a n-curve than z-curve, so we need to swap x and y axes
			i := m.Pack(uint64(y), uint64(x))
			untwiddle(d.rgba, rgba2, x, y, i)
		}
	}
	copy(d.rgba.Pix, rgba2.Pix)
}

// Decode reads a Stunt GP texture image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	var d decoder
	d.r = r
	if err := d.decodeConfig(); err != nil {
		return nil, err
	}

	switch d.version {
	case 2:
		colours, _ := getWord(d.r)
		d.palette = make([]color.Color, colours)
		d.readPalette()
		d.paletted = image.NewPaletted(image.Rect(0, 0, int(d.width), int(d.height)), d.palette)
		d.decodePalleted()
		return d.paletted, nil
	case 3:
		d.rgba = image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
		d.decode()
		return d.rgba, nil
	case 4:
		d.rgba = image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
		d.decode()
		d.deswizzle()
		return d.rgba, nil
	case 5:
		d.rgba = image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
		d.decode()
		d.swapRedBlue()
		return d.rgba, nil
	default:
		return nil, FormatError(fmt.Sprintf("Unknown file format version! This program supports versions 2-5, but got version %d instead\n", d.version))
	}
}

func (d *decoder) decodeConfig() error {
	buf := make([]byte, len(pcHeader)+6)
	if _, err := io.ReadFull(d.r, buf); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	if string(buf[:len(pcHeader)]) != pcHeader {
		return FormatError("pc: invalid format")
	}

	d.version = binary.LittleEndian.Uint16(buf[4:6])

	d.width = binary.LittleEndian.Uint16(buf[6:8])
	d.height = binary.LittleEndian.Uint16(buf[8:10])

	if d.width == 0 || d.height == 0 {
		return FormatError(fmt.Sprintf("pc: unsupported size: %dx%d\n", d.width, d.height))
	}

	return nil
}

// DecodeConfig returns the color model and dimensions of a Stunt GP texture without
// decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	var d decoder
	d.r = r
	if err := d.decodeConfig(); err != nil {
		return image.Config{}, err
	}

	return image.Config{
		ColorModel: color.NRGBAModel,
		Width:      int(d.width),
		Height:     int(d.height),
	}, nil
}

func init() {
	image.RegisterFormat("pc", pcHeader, Decode, DecodeConfig)
	image.RegisterFormat("dc", pcHeader, Decode, DecodeConfig)
}
