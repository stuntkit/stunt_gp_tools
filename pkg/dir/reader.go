package dir

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// FromDir creates new Dir map from a dir file
func FromDir(filename string) (Dir, error) {
	dir := Dir{}

	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	header, err := getHeader(r)
	if err != nil {
		return nil, err
	}

	if string(header.Magic[:]) != "DIR\x1a" {
		return nil, fmt.Errorf("Incorrect magic, expected DIR\x1a, got %s", header.Magic)
	}

	// jump to hash array
	_, err = r.Seek(int64(header.DirectoryAddress), 0)
	if err != nil {
		return nil, err
	}

	table, err := readTable(r)
	if err != nil {
		return nil, err
	}

	for tableElement := range table.Entries {
		_, err = r.Seek(int64(header.DirectoryAddress)+4+(int64(tableElement)*4), 0)
		if err != nil {
			return nil, err
		}
		var elementOffset uint32
		err := binary.Read(r, binary.LittleEndian, &elementOffset)
		if err != nil {
			return nil, err
		}
		err = dir.readElements(r, header.DirectoryAddress, elementOffset)
		if err != nil {
			return nil, err
		}
	}
	err = r.Close()
	if err != nil {
		return nil, err
	}
	return dir, nil
}

func getHeader(r io.Reader) (Header, error) {
	header := Header{}
	err := binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return header, err
	}
	return header, nil
}

func readTable(r io.Reader) (HashTable, error) {
	table := HashTable{}
	err := binary.Read(r, binary.LittleEndian, &table)
	if err != nil {
		return table, err
	}
	if string(table.Header[:]) != "\x0a\x00\x00\x00" {
		return table, fmt.Errorf("Incorrect magtable header, expected \x0a, got %s", table.Header)
	}
	return table, nil
}

// parses one whole hash value chain
func (d *Dir) readElements(r *os.File, tableOffset, elementOffset uint32) error {
	// skip empty hashes
	if elementOffset == 0 {
		return nil
	}

	_, err := r.Seek(int64(tableOffset)+int64(elementOffset), 0)
	if err != nil {
		return err
	}

	var meta FileMetadata
	var metaSmall FileMetadataSmall
	err = binary.Read(r, binary.LittleEndian, &metaSmall)
	if err != nil {
		return err
	}
	filename := ""
	// TODO read name
	buf := make([]byte, 1)

	for {
		_, err := r.Read(buf)
		if err != nil {
			return fmt.Errorf("couldn't parse filename %s", err)
		}
		if buf[0] != '\x00' {
			filename += string(buf)
		} else {
			break
		}
	}

	meta = FileMetadata{
		NextHash:   metaSmall.NextHash,
		DataOffset: metaSmall.DataOffset,
		DataSize:   metaSmall.DataSize,
		Filename:   filename,
	}

	_, err = r.Seek(int64(meta.DataOffset), 0)
	if err != nil {
		return err
	}

	fileData := make([]byte, meta.DataSize)
	// TODO check if this reader type is correct here
	_, err = r.Read(fileData)
	if err != nil {
		return err
	}

	(*d)[meta.Filename] = fileData
	if meta.NextHash > 0 {
		err = d.readElements(r, tableOffset, meta.NextHash)
		if err != nil {
			return err
		}
	}
	return nil
}
