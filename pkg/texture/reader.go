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
	r       io.Reader
	version uint16
	width   uint16
	height  uint16
	// image []*image.Image
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

// Decode reads a Stunt GP texture image from r and returns it as an image.Image.
/*func Decode(r io.Reader) (image.Image, error) {
	var d decoder
	d.r = r
	if err := d.decodeConfig(); err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
	unpacked := 0
	var word uint16
	word, _ = getWord(r)

	// unpack first pixel as-is
	var p []byte = rgba.Pix[unpacked*4 : 4*(1+unpacked)]
	setPixel(p, word)
	unpacked++

	for unpacked <= int(d.width)*int(d.height) {
		word, _ = getWord(r)

		for ((word >> 15) & 1) == 1 {
			// stream pixels
			p = rgba.Pix[countIndex(unpacked, d.width)*4 : 4*(1+countIndex(unpacked, d.width))]
			setPixel(p, word)
			word, _ = getWord(r)
			unpacked++
		}

		if ((word >> 14) & 1) == 1 {
			// packed data
			repeats := int(((word >> 11) & 7) + 2)
			location := int(word & 0x7FF)

			for j := 0; j < repeats; j++ {
				p = rgba.Pix[4*(countIndex(unpacked+j, d.width)) : 4*(countIndex(unpacked+j, d.width)+1)]
				p2 := rgba.Pix[4*(countIndex(unpacked-1-location+j, d.width)) : 4*(1+countIndex(unpacked-1-location+j, d.width))]
				copy(p, p2)
			}
			unpacked += repeats

		} else {
			if ((word >> 15) & 1) == 0 {
				// end reading
				if word == 0 {
					return rgba, nil
				}
				// stream transparency
				unpacked += int(word)
			}
		}
	}
	return nil, FormatError("pc: too many bytes")
}*/

/*func untwiddleIndex(l, x, y uint16) int {
		size = 8

		l = math.log(size, 2)

			if l <= 0:
				print(x, y)
				return

			step = math.pow(2, l-1)
			untwiddleIndex(l-1, x, y)
			untwiddleIndex(l-1, x+step, y)
			untwiddleIndex(l-1, x, y+step)
			untwiddleIndex(l-1, x+step, y+step)

}*/

func untwiddle(rgba, rgba2 *image.RGBA, x, y uint16, i int64) {
	pos := (int(y) * rgba.Bounds().Dy()) + int(x)
	p := rgba2.Pix[pos*4 : 4*(pos+1)]
	p2 := rgba.Pix[4*i : 4*(i+1)]
	copy(p, p2)
}

// Decode reads a Stunt GP texture image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	var d decoder
	d.r = r
	if err := d.decodeConfig(); err != nil {
		return nil, err
	}

	rgba := image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
	unpacked := 0
	var word uint16
	word, _ = getWord(r)

	// unpack first pixel as-is
	var p []byte = rgba.Pix[unpacked*4 : 4*(1+unpacked)]
	setPixel(p, word)
	unpacked++

	for unpacked <= int(d.width)*int(d.height) {
		word, _ = getWord(r)

		for ((word >> 15) & 1) == 1 {
			// stream pixels
			p = rgba.Pix[unpacked*4 : 4*(1+unpacked)]
			setPixel(p, word)
			word, _ = getWord(r)
			unpacked++
		}

		if ((word >> 14) & 1) == 1 {
			// packed data
			repeats := int(((word >> 11) & 7) + 2)
			location := int(word & 0x7FF)

			for j := 0; j < repeats; j++ {
				p = rgba.Pix[4*(unpacked+j) : 4*(unpacked+j+1)]
				p2 := rgba.Pix[4*(unpacked-1-location+j) : 4*(unpacked-location+j)]
				copy(p, p2)
			}
			unpacked += repeats

		} else {
			if ((word >> 15) & 1) == 0 {
				// end reading
				if word == 0 {

					// untwiddle
					if d.version == 4 {
						rgba2 := image.NewRGBA(image.Rect(0, 0, int(d.width), int(d.height)))
						// TODO check if image is square and size is power of two
						m := morton.Make64(2, uint64(math.Log2(float64(d.width))))

						for y := uint16(0); y < d.height; y++ {
							for x := uint16(0); x < d.width; x++ {
								// DZ uzez more of a n-curve than z-curve, so we need to swap x and y axes
								i := m.Pack(uint64(y), uint64(x))
								untwiddle(rgba, rgba2, x, y, i)
							}
						}
						return rgba2, nil
					}

					return rgba, nil
				}
				// stream transparency
				unpacked += int(word)
			}
		}
	}
	return nil, FormatError("pc: too many bytes")
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
