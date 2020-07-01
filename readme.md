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
    * 4 bytes - if two files have the same hash, then hashtable points to first file, which have this 4 bytes pointed to another one with the same hash. If there is no collision then it's 0.
    * 4 bytes - file offset in archive, should be 4 bytes aligned
    * 4 bytes - file size
    * file name with path, 4 bytes aligned, terminated by 0x00 (and padded with 0x00 to 4 bytes)

# PC textures
All textures are in this format, engine supports maximum size of 256x256 pixels.

* 4 bytes - magic `T54 4D 21 1A`, (`TM!` and one unprintable)
* 2 bytes - perhaps format? I only seen this as `03 00`
* 2 bytes width
* 2 bytes height
* compressed image data starts at 0xA, see pc_unpack.py for implementation details
    * Each pixel is written on 2 bytes, little endian, after converting to big endian the algorithm looks like this:
    * If the 15th bit is 1, then stream data as long as each compressed pixel 15th bit is set
        * Else
            * If 14th bit is set copy previously unpacked data
                * bits 11-13 - number of pixels to repeat - 1, so if this number is set to 7, then repeat 8 pixels
                * bits 0-10 - offset - back head for copying that many pixels
                * else - add transparent pixels
                    * bits 0-13 - how many transparent pixels to add
    * `00 00` marks the end of compressed data

### Unpacked image data
Uncompressed image data uses 2 bytes for each pixel, in A1 R5 G5 B5 format. In some places game uses additional file with greyscale alpha channel to compensate for 1bit alpha (5 bits instead of 1).

# PMD files
Stores 3D models, more investigation needed

* Magic - `'PMD V1.83'`
    * some unused files use version 1.82 instead


# Binary config file
Binary config file is described in okteta structures directory

* music volume is always set to 100 on save, other volumes just gets copied from config.cfg file
* place for path is 200 bytes long, it's ASCII string (unicode might work, probably not made with that in mind) terminated by "\0".
    * Windows reads PATH from registy in Unicode 16 and writes it to this place and later converts it in-place to ASCII, leaving some garbage after string terminator
        * Considering that, is maximum path length 99 characters then?

# Various

* wads/fonts.wad: graphics24/fonts/index.txt - uses ISO-8859-1 encoding (tested in Polish version, YMMV)
* in StuntGP_D3D.exe all radeon grapics card have blocked best quality settings, including newest ones.
* Resolution is limited to 2048x2048 in binary. Editing binary and using another dll should allow to override this limit.
    * DirectDraw seems to be limiting resolution to 1920x1200 anyway
    * Glide mode + dgVoodoo can override this limitation