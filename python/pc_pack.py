#!/usr/bin/python3 
# coding=utf-8

import argparse
# import sys
from pathlib import Path

from SGPTools import PC

parser = argparse.ArgumentParser()
parser.add_argument("filename", help="name of the png file")
parser.add_argument("-o", "--output", help="name of the output file")
parser.add_argument("-u", "--uncompressed", action="store_true", help="Don't compress resulting  pc file")


def main():
    args = parser.parse_args()

    pc = PC.from_png(args.filename)

    if not pc:
        raise Exception('Not a valid pc file')

    output = Path(args.filename).with_suffix('.pc')
    print(output)
    if args.output:
        output = args.output

    compress = True
    if args.uncompressed:
        compress = False
    pc.to_pc(output, compress)

if __name__ == "__main__":
    main()
