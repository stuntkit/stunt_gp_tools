package texture

import (
	"encoding/binary"
	"io"
)

// FormatName is the name of the registered texture format
const FormatName = "stuntGP"

const textureHeader = "TM!\x1a"

// TODO there's gotta be a better way
const (
	FIncorrect = iota // we''ll default to FPC
	// TODO check PS2/prototype versions for this
	FUnknown   // we don't have any examples of these textures at all
	FPaletted  // paletted images used by PS2
	FPC        // Windows version
	FDreamcast // Dreamcast swizzled textures
	FPSTwo     // PS2 textures with swapped R B channels
)

func getWord(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf), nil
}
