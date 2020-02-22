Stunt GP tools
==
These tools will help you understand, unpack and edit Stunt GP files

# WAD packs
every address and size is in little endian format, almost everything is 4bytes padded

* 4 bytes - `44 49 52 1A` - magic ('DIR' and one non-ASCII character)
* 4 bytes - file size
* 4 bytes - address of something
* file data, usually starts at 0xC and spans up to the "address of something"
* at the "address of something" there is something I was unable to figure out what it is, and it spans for 4100 bytes
* the "addres of something" + 4100 bytes is where file descriptors start, each one have following structure:

## File descriptor

* 4 bytes - unknown, most files have this set to `00 00 00 00`, few have other stuff there, TODO
* 4 bytes - file offset in archive
* 4 bytes - file size
* file name with path, 4 bytes aligned, terminated by 0x00 (and padded with 0x00 to 4 bytes)

# PC files
All textures are in this format

Every address and size is in little endian format

* 4 bytes - magic `T54 4D 21 1A`, (`TM!` and one unprintable, probably still part of magic, as T17 LOVES aligning stuff)
* 2 bytes???? perhaps format? I only seen this as `03 00`
* 2 bytes width, theroretically up to 65 535
* 2 bytes height
* compressed image data starts at 0xA

## compression
Unknown, but there's working decompressor from Martin here:  [https://forum.re-volt.io/viewtopic.php?t=391](https://forum.re-volt.io/viewtopic.php?t=391).  
It only outputs BMP files, so alpha is lost. And sadly, it's just assemby with no comments or idea how it works.

### What I know
Compressed data seems to always have even number of bytes, and probably is read by 2 bytes

## Unpacked image data
Uncompressed image data uses 2 bytes for each pixel, probably RGBA 4444

# Various

* wads/fonts.wad: graphics24/fonts/index.txt - uses ISO-8859-1 encoding
