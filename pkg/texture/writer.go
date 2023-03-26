package texture

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"math"
	"strconv"

	// "github.com/Nykakin/quantize"
	morton "github.com/gojuno/go.morton"
)

// func dummy() {
// 	quantizer := quantize.NewHierarhicalQuantizer()
// }

// Encoder configures encoding Stunt GP textures
type Encoder struct {
	Compress bool
	Format   int
}

type encoder struct {
	enc     *Encoder
	w       io.Writer
	m       image.Image
	err     error
	d       []uint16
	version uint16
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func uint16ToBytes(i uint16) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, i)
	return bytes
}

func argb1555(c color.Color) uint16 {
	r, g, b, a := c.RGBA()
	// TODO control alpha treshold through Encoder?
	var col16 uint16
	var a1 uint16 = 0
	if a >= 128 {
		a1 = 1
	}

	r5 := uint16(math.Round((float64(r) / 65536.0) * 31.0))
	g5 := uint16(math.Round((float64(g) / 65536.0) * 31.0))
	b5 := uint16(math.Round((float64(b) / 65536.0) * 31.0))

	col16 = a1<<15 | r5<<10 | g5<<5 | b5

	return col16
}

func (e *encoder) convertData() {
	e.d = make([]uint16, e.m.Bounds().Dx()*e.m.Bounds().Dy()) // initialize a slice of dy slices

	if e.version == 4 {
		m := morton.Make64(2, uint64(math.Log2(float64(e.m.Bounds().Dy()))))

		for y := 0; y < e.m.Bounds().Dy(); y++ {
			for x := 0; x < e.m.Bounds().Dx(); x++ {
				mCoord := m.Unpack(int64((y * e.m.Bounds().Dy()) + x))
				e.d[x+(y*e.m.Bounds().Dy())] = argb1555(e.m.At(int(mCoord[1]), int(mCoord[0])))
			}
		}
	} else {

		for y := 0; y < e.m.Bounds().Dy(); y++ {
			for x := 0; x < e.m.Bounds().Dx(); x++ {
				e.d[x+(y*e.m.Bounds().Dx())] = argb1555(e.m.At(x, y))
			}
		}
	}

}

func (e *encoder) writeHeader() {
	_, e.err = io.WriteString(e.w, textureHeader)
}

func (e *encoder) writeVersion() {
	w := make([]byte, 2)
	// TODO allow to select Dreamcast version
	binary.LittleEndian.PutUint16(w, e.version)
	e.w.Write(w)
}

func (e *encoder) writeWidth() {
	w := make([]byte, 2)
	binary.LittleEndian.PutUint16(w, uint16(e.m.Bounds().Dx()))
	e.w.Write(w)
}

func (e *encoder) writeHeight() {
	h := make([]byte, 2)
	binary.LittleEndian.PutUint16(h, uint16(e.m.Bounds().Dy()))
	e.w.Write(h)
}

func (e *encoder) packLoop(len, j, u int) int {
	currentCount := 0
	for u+currentCount+1 < len && e.d[u-j+currentCount] == e.d[u+currentCount] {
		currentCount++
	}
	return currentCount
}

func (e *encoder) writeData() {
	// magic here I guess
	//write first pixel as-is
	length := len(e.d)

	e.w.Write(uint16ToBytes(uint16(e.d[0])))

	u := 1
	for u < length {
		// check if pixlex is trasparent (15th bit is unset)
		if e.d[u] < 0x8000 {
			//transparent pixels
			count := 1
			for u+count+1 < length && e.d[u+count] == 0 && count < 16383 {
				count++
			}
			e.w.Write(uint16ToBytes(uint16(count)))
			u += count
		} else {
			if e.enc.Compress {
				count := 0
				location := 0
				j := min(u, 2048)
				for j > 0 {
					currentCount := e.packLoop(length, j, u)

					if currentCount > count {
						count = currentCount
						location = j
						if count >= 9 {
							break
						}
					}

					j -= currentCount + 1
				}

				if count > 1 {
					for count > 1 {
						currentlyCounted := min(9, count)
						pack := 0b0100000000000000 | ((currentlyCounted - 2) << 11) | (location - 1)
						e.w.Write(uint16ToBytes(uint16(pack)))
						u += currentlyCounted
						count -= currentlyCounted
					}
				} else {
					e.w.Write(uint16ToBytes(uint16(e.d[u])))
					u++
				}
			} else {
				e.w.Write(uint16ToBytes(uint16(e.d[u])))
				u++
			}
		}
	}

}

func (e *encoder) writeEnd() {
	io.WriteString(e.w, "\x00\x00")
}

// Encode writes the Image m to w in Stunt GP texture format. Any Image may be
// encoded, but images that are not ARGB1555 might be encoded lossily.
func Encode(w io.Writer, m image.Image) error {
	var e Encoder
	e.Compress = true
	e.Format = FPC
	return e.Encode(w, m)
}

// Encode writes the Image m to w in Stunt GP texture format.
func (enc *Encoder) Encode(w io.Writer, m image.Image) error {
	// limit width and height to be above zero and fit in uint16
	mw, mh := int64(m.Bounds().Dx()), int64(m.Bounds().Dy())
	if mw <= 0 || mh <= 0 || mw >= 1<<16 || mh >= 1<<16 {
		return FormatError("invalid image size: " + strconv.FormatInt(mw, 10) + "x" + strconv.FormatInt(mh, 10))
	}

	var e *encoder
	if e == nil {
		e = &encoder{}
	}
	e.enc = enc
	e.w = w
	e.m = m
	if enc.Format == FDreamcast {
		e.version = 4
	} else {
		e.version = 3
	}

	e.convertData()

	e.writeHeader()
	e.writeVersion()
	e.writeWidth()
	e.writeHeight()
	e.writeData()
	e.writeEnd()

	return e.err
}
