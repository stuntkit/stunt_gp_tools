Stunt GP tools
==
These tools will help you understand, unpack and edit Stunt GP files

Original thread: [https://forum.xentax.com/viewtopic.php?f=16&t=16944&p=160266#p160266](https://forum.xentax.com/viewtopic.php?f=16&t=16944&p=160266#p160266)


Every address and size is in little endian format, many values are 4 bytes aligned/padded

File formats description were moved to [the wiki](https://github.com/Halamix2/stunt_gp_formats/wiki)

# Various

* wads/fonts.wad: graphics24/fonts/index.txt - uses ISO-8859-1 encoding (tested in Polish version, YMMV)
* in StuntGP_D3D.exe all radeon grapics card have blocked best quality settings, including newest ones.
* Resolution is limited to 2048x2048 in binary. Editing binary and using another dll should allow to override this limit.
    * DirectDraw seems to be limiting resolution to 1920x1200 anyway
    * Glide mode + dgVoodoo can override this limitation
