#!/usr/bin/python3 
# coding=utf-8

import argparse
from pathlib import Path

from SGPTools import DIR


parser = argparse.ArgumentParser()
parser.add_argument("filename", help="name of the dir file")
parser.add_argument("-o", "--output", help="name of the output folder")


def main():
    args = parser.parse_args()

    dir = DIR.from_dir(args.filename)

    if not dir:
        raise Exception('Not a valid dir archive')

    output = Path(args.filename).parent/Path(args.filename).stem
    if args.output:
        output = args.output
    dir.to_folder(output)

if __name__ == "__main__":
    main()
