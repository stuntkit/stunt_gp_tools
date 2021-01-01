#!/usr/bin/python3 
# coding=utf-8

import argparse
# import sys
from pathlib import Path

from SGPTools import PC

parser = argparse.ArgumentParser()
parser.add_argument("filename", help="name of the pc file")
parser.add_argument("-o", "--output", help="name of the output file")


def main():
    args = parser.parse_args()

    pc = PC.from_pc(args.filename)

    if not pc:
        raise Exception('Not a valid pc file')

    output = Path(args.filename).with_suffix('.png')
    print(output)
    if args.output:
        output = args.output
    pc.to_png(output)

if __name__ == "__main__":
    main()
