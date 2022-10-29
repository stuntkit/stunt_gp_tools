#!/usr/bin/python3 
# coding=utf-8

import argparse
from pathlib import Path

from SGPTools import DIR


parser = argparse.ArgumentParser()
parser.add_argument("filename", help="name of the folder")
parser.add_argument("-o", "--output", help="name of the output archive")


def main():
    args = parser.parse_args()

    dir = DIR.from_folder(args.filename)

    if not dir:
        raise Exception('Not a folder apparently')
    
    output = str(Path(args.filename).parent/Path(args.filename).stem)+'.dir'
    if args.output:
        output = args.output
    dir.to_dir(output)

if __name__ == "__main__":
    main()
