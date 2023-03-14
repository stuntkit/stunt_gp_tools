#!/usr/bin/env python3

import struct 

def checksum(data):
    # originally loop was hardcoded to save size, 0x211C bytes, but we can do this instead
    hash = 0
    # data[i] + rotate hash left 3 bits on 32 bits, I hope one crop to 32 bits is enough
    for i in range(0, len(data)):
        hash = ( int(data[i]) + ( hash >> 0x1D | hash << 3 ) ) & 0xffffffff
    return hash
    
def main():
    with open('setup.bin', 'rb') as savefile:
        copied = list(savefile.read())
        expected = struct.unpack("<I",bytes(copied[4:8]))[0]
        # set checksum to 0 
        copied[4] = b'0'
        copied[5] = b'0'
        copied[6] = b'0'
        copied[7] = b'0'

        check = checksum(copied)

        if expected != check:
            print(f"expected {hex(expected)}, got {hex(check)}")
        else:
            print('The checksum is correct')

if __name__ == "__main__":
    main()