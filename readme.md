Stunt GP tools
==
These tools will help you understand, unpack and edit Stunt GP files

Every address and size is in little endian format, many values are 4bytes aligned/padded

# WAD packs
This file is exactly the same as Worms .dir file, onyl extension is different.  

* 4 bytes - `44 49 52 1A` - magic ('DIR' and one non-ASCII character)
* 4 bytes - file size
* 4 bytes - address of directory
* file data, usually starts at 0xC and spans up to the "address of directory"
    * There's `1A 00` at the end of each file
    * There may be `00` padding after that, so the next file/section starts at offset % 4 == 0
* at the "address of directory" there is `0A 00 00 00`
* After this 1024 4byte entries, this is hash table for all files in archive
    * each entry value + "address of directory" is a pointer to corresponding *file descriptor*
    * each position is hardcoded and should always corespond to the same file in archive
* each **file descriptor** have following structure:
    * 4 bytes - unknown, most files have this set to `00 00 00 00`, few have other stuff there, TODO
    * 4 bytes - file offset in archive, should be 4 bytes aligned
    * 4 bytes - file size
    * file name with path, 4 bytes aligned, terminated by 0x00 (and padded with 0x00 to 4 bytes)

# PC files
All textures are in this format, biggest textures used in game have 256x256 pixels.

* 4 bytes - magic `T54 4D 21 1A`, (`TM!` and one unprintable)
* 2 bytes - perhaps format? I only seen this as `03 00`
* 2 bytes width
* 2 bytes height
* compressed image data starts at 0xA, see pc_unpack.py for implementation details

### Unpacked image data
Uncompressed image data uses 2 bytes for each pixel, in A1 R5 G5 B5 format. In some places game uses additional file with greyscale alpha channel to compensate for 1bit alpha.

# Config files
Binary config files are described in okteta structures directory, text ones are self-explanatory

# Various

* wads/fonts.wad: graphics24/fonts/index.txt - uses ISO-8859-1 encoding
