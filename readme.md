Stunt GP tools
==
[![a](https://discord.com/api/guilds/749260704447463495/widget.png?style=shield)](https://discord.gg/ykzAWnA)

[![DeepSource](https://deepsource.io/gh/stuntkit/stunt_gp_formats.svg/?label=active+issues&token=xrA_xeEtEOj9PK-TMnSL4ZBU)](https://deepsource.io/gh/stuntkit/stunt_gp_formats/?ref=repository-badge)

These tools will help you understand, unpack and edit Stunt GP files

Original thread: [https://forum.xentax.com/viewtopic.php?f=16&t=16944&p=160266#p160266](https://forum.xentax.com/viewtopic.php?f=16&t=16944&p=160266#p160266)

Check out [the wiki](https://sgp.halamix2.pl) for more information about the game and its file formats.

## Compilation:
```
go build cmd/pc_pack/pc_pack.go
go build cmd/pc_unpack/pc_unpack.go
```
Or grab compiled .exe [here](https://github.com/StuntKit/stunt_gp_formats/releases)

## Usage:
You can also drag and drop multiple files on `pc_pack` or `pc_unpack`

```bash
./pc_pack mini.png

./pc_pack mini.png -o output.pc

# pack Dreamcast texture
./pc_pack mini.png --dc
./pc_pack mini.png -o output.dc --dc

# unpack texture (including Dreamcast)
./pc_unpack mini.pc

./pc_unpack mini.pc -o output.png
```

