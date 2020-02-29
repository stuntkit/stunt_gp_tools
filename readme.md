Stunt GP tools
==
These tools will help you understand, unpack and edit Stunt GP files

# WAD packs
every address and size is in little endian format, almost everything is 4bytes padded

* 4 bytes - `44 49 52 1A` - magic ('DIR' and one non-ASCII character)
* 4 bytes - file size
* 4 bytes - address of something
* file data, usually starts at 0xC and spans up to the "address of something"
    * There's `1A 00` at the end of each file
    * There may be `00` padding after that, so the next file starts at offset % 4 == 0
* at the "address of something" there is `0A 00 00 00`
* After this 4096 bytes of something that I was unable to figure out what it is, necessary for game apparently
    * probably in groups of 4 bytes (always 2 bytes of data, smallest gap between data 2 `00` bytes)
    * each entry value + "address of something" is a pointer to corresponding *file descriptor*  
    * bytes never overlap between all .wad files
    * 1024 4bytes entries theory
        * Combined from all .wad files, only 739 of these entries contain data
    * To do
        * Why the data are on these coordinates?
            * Hardcoded table of 1024 entries? There are over 1300 .pc textures in game archives, maybe if we group all tires etc. together it would be closer to 739?
* Later on (the "addres of something" + 4100 bytes) is where the **file descriptors** start, each one have following structure"
    * 4 bytes - unknown, most files have this set to `00 00 00 00`, few have other stuff there, TODO
    * 4 bytes - file offset in archive, should be 4 bytes aligned
    * 4 bytes - file size
    * file name with path, 4 bytes aligned, terminated by 0x00 (and padded with 0x00 to 4 bytes)

# PC files
All textures are in this format

Every address and size is in little endian format

* 4 bytes - magic `T54 4D 21 1A`, (`TM!` and one unprintable, probably still part of magic, as SGP LOVES aligning stuff)
* 2 bytes - perhaps format? I only seen this as `03 00`
* 2 bytes width
* 2 bytes height
* compressed image data starts at 0xA, see pc_unpack.py for implementation details

## Unpacked image data
Uncompressed image data uses 2 bytes for each pixel, in A1 R5 G5 B5 format

## pc_pack.py
Very simple and naive implementation for non-alpha images to see if they will work in-game, no compression or alpha yet.

# Various

* wads/fonts.wad: graphics24/fonts/index.txt - uses ISO-8859-1 encoding
