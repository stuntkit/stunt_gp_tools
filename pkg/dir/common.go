// Package dir implements functions used to manipulate Worms Armageddon / Stunt GP .dir / .wad files.
package dir

// Dir is a base representation of a Dir archive
type Dir map[string][]byte

// HashTable stores whole hash table
type HashTable struct {
	Header  [4]byte
	Entries [1024]uint32
}

// Header contains all data at the beginning of a file
type Header struct {
	Magic            [4]byte
	Size             uint32
	DirectoryAddress uint32
}

// FileMetadata contains all metadata stored in the file list
type FileMetadata struct {
	NextHash   uint32
	DataOffset uint32
	DataSize   uint32
	Filename   string
}

// FileMetadataSmall contains metadata stored in the file list without the filename
type FileMetadataSmall struct {
	NextHash   uint32
	DataOffset uint32
	DataSize   uint32
}

// GetHash generates hash used by the DIR archives. Filename should be provided in ASCII encoding and without trailing \0
func GetHash(filename string) uint32 {
	var hash uint32

	const hashBits uint32 = 10
	const hashSize uint32 = 1 << hashBits

	for _, char := range filename {
		hash = (uint32(hash<<1) % hashSize) | uint32(hash>>(hashBits-1)&1)
		hash = uint32(hash+uint32(char)) % hashSize
	}

	return hash
}
