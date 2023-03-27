#!/usr/bin/env python3

import struct


def checksum(data):
    """calculate Stunt GP savefile checksum"""
    hash_val: int = 0
    # i + rotate hash left 3 bits on 32 bits
    # I hope one crop to 32 bits is enough
    for i in data:
        hash_val = (int(i) + (hash_val >> 0x1D | hash_val << 3)) & 0xFFFFFFFF
    return hash_val


def main():
    """opens setup.bin, calculates checksum and compares against saved one"""
    with open("setup.bin", "rb") as savefile:
        copied = list(savefile.read())
        expected = struct.unpack("<I", bytes(copied[4:8]))[0]
        # set checksum bytes to 0
        copied[4] = b"0"
        copied[5] = b"0"
        copied[6] = b"0"
        copied[7] = b"0"

        check = checksum(copied)

        if expected != check:
            print(f"expected {hex(expected)}, got {hex(check)}")
        else:
            print("The checksum is correct")


if __name__ == "__main__":
    main()
