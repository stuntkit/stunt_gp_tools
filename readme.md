Stunt GP tools
==
These tools will help you understand, unpack and edit Stunt GP files

# WAD packs
every address is in little endian format, almost everything is 4bytes padded

`44 49 52 1A` - magic ('DIR' and one non-ASCII character)
next 8 bytes:
4 bytes - file size
4 bytes - address of something

Then there's file data

If you go to the addres of something + 4100 bytes you will be at the descriptor of the first file

## File descriptor

* 4 bytes - unknown, most files have this set to `00 00 00 00`
* 4 bytes - file offset
* 4 bytes - file size
* name, terminated by 0x00 (and padded with 0x00 to 4 bytes)


# Various

wads/fonts.wad: graphics24/fonts/index.txt - ISO-8859-1 encoding
