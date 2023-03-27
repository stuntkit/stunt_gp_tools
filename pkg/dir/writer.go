package dir

import (
	"encoding/binary"
	"io"
	"os"
	"sort"
)

// FileData contains physical file data
type FileData struct {
	Data    []byte
	Footer  []byte
	Padding []byte
}

type writer struct {
	*Dir
	keys             []string
	dataSize         uint32
	descriptorsSize  uint32
	descriptorsTable []string
	fh               map[string]fileHelper
}

type fileHelper struct {
	dataOffset       uint32
	descriptorOffset uint32
	padding          int
	filenamePadding  int
	nextFile         string
}

// ToDir creates new Dir file from a Dir
func (d *Dir) ToDir(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	err = d.ToWriter(f)
	if err != nil {
		return err
	}

	return f.Close()
}

// ToWriter saves Dir archive to a file object
func (d *Dir) ToWriter(f io.Writer) error {
	w := writer{Dir: d}
	w.keys = w.getSortedKeys()
	w.prepareMetadata()
	w.prepareDescriptorTable()

	err := w.writeAll(f)
	if err != nil {
		return err
	}

	return nil
}

func (d *Dir) getSortedKeys() []string {
	keys := make([]string, 0)
	for k := range *d {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

func (w *writer) prepareMetadata() {
	dataSize := 0
	descriptorOffset := 4 + (4 * 1024)
	w.fh = make(map[string]fileHelper)

	// data should be padded to nearest 4 bytes
	// luckily the Dir header is 12 bytes
	for _, k := range w.keys {
		fh := w.fh[k]
		fh.dataOffset = uint32(12 + dataSize)

		data := (*w.Dir)[k]
		dataSize += len(data) + 2
		padding := dataSize % 4
		if padding > 0 {
			fh.padding = 4 - padding
			dataSize += fh.padding
		}

		fh.descriptorOffset = uint32(descriptorOffset)
		// +1 for null byte, and then some padding for a neat file
		descriptorOffset += 12 + len(k) + 1
		filenamePadding := descriptorOffset % 4
		if filenamePadding > 0 {
			fh.filenamePadding = 4 - filenamePadding
			descriptorOffset += fh.filenamePadding
		}

		w.fh[k] = fh
	}
	w.dataSize = uint32(dataSize)
	w.descriptorsSize = uint32(descriptorOffset)
}

func (w *writer) prepareDescriptorTable() {
	w.descriptorsTable = make([]string, 1024)

	for _, k := range w.keys {
		hash := GetHash(k)
		if w.descriptorsTable[hash] == "" {
			w.descriptorsTable[hash] = k
		} else {
			s := w.descriptorsTable[hash]
			fh := w.fh[s]

			for fh.nextFile != "" {
				s = fh.nextFile
				fh = w.fh[s]
			}
			fh.nextFile = k
			w.fh[s] = fh
		}
	}
}

func (w *writer) writeAll(f io.Writer) error {
	if err := w.writeHeader(f); err != nil {
		return err
	}

	if err := w.writeData(f); err != nil {
		return err
	}

	if err := w.writeDirectory(f); err != nil {
		return err
	}

	return w.writeEntries(f)
}

func (w *writer) writeHeader(f io.Writer) error {
	magic := []byte("DIR\x1A")
	dirOffset := 12 + w.dataSize
	fileSize := dirOffset + w.descriptorsSize

	if _, err := f.Write(magic); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, fileSize); err != nil {
		return err
	}
	return binary.Write(f, binary.LittleEndian, dirOffset)
}

func (w *writer) writeData(f io.Writer) error {
	fileFooter := []byte("\x1A\x00")
	for _, k := range w.keys {
		fh := w.fh[k]
		if _, err := f.Write((*w.Dir)[k]); err != nil {
			return err
		}
		if _, err := f.Write(fileFooter); err != nil {
			return err
		}
		if err := writePadding(f, fh.padding); err != nil {
			return err
		}

	}
	return nil
}

func (w *writer) writeDirectory(f io.Writer) error {
	directoryHeader := []byte("\x0A\x00\x00\x00")
	if _, err := f.Write(directoryHeader); err != nil {
		return err
	}

	for _, k := range w.descriptorsTable {
		if k != "" {
			fh := w.fh[k]
			if err := binary.Write(f, binary.LittleEndian, fh.descriptorOffset); err != nil {
				return err
			}
		} else {
			if err := binary.Write(f, binary.LittleEndian, uint32(0)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *writer) writeEntries(f io.Writer) error {
	for _, k := range w.keys {
		fh := w.fh[k]
		data := (*w.Dir)[k]

		nextOffset := uint32(0)
		if fh.nextFile != "" {
			fhNext := w.fh[fh.nextFile]
			nextOffset = fhNext.descriptorOffset
		}

		if err := binary.Write(f, binary.LittleEndian, nextOffset); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, fh.dataOffset); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, uint32(len(data))); err != nil {
			return err
		}
		if _, err := f.Write([]byte(k)); err != nil {
			return err
		}
		if _, err := f.Write([]byte{'\x00'}); err != nil {
			return err
		}
		if err := writePadding(f, fh.filenamePadding); err != nil {
			return err
		}
	}
	return nil
}

func writePadding(f io.Writer, padSize int) error {
	var err error
	for i := 0; i < padSize; i++ {
		_, err = f.Write([]byte{'\x00'})
	}
	return err
}
