#!/usr/bin/python3 
# coding=utf-8

import argparse
import sys
from pathlib import Path

from PIL import Image

parser = argparse.ArgumentParser()
parser.add_argument("filename", help="name of the image file")

def to_bytes(number, bytes = 2):
    return number.to_bytes(bytes, 'little')
    
def sixteen(image):
    crappy = b''
    
    pixels = image.load() # create the pixel map

    for j in range(image.width):
        for i in range(image.height):
            #pixels[i,j] = (i, j, 100) # set the colour accordingly
            red = int(pixels[i, j][0] * (31/255) ) << 10
            green = int(pixels[i, j][1] * (31/255) ) << 5
            blue = int(pixels[i, j][2] * (31/255) )
            alpha = 1 << 15
            pixel = to_bytes(red+green+blue+alpha, 2)
            crappy = crappy + pixel

    return crappy
def pack(image):
    packed_data = b''
    
    #magic
    packed_data = packed_data + b'TM!\x1a'
    
    #unknown, probably still needed???
    packed_data = packed_data + to_bytes(3, 2)
    
    packed_data = packed_data + to_bytes(image.width, 2)
    
    packed_data = packed_data + to_bytes(image.height, 2)
    
    image = sixteen(image)
    
    packed_data = packed_data + image
    
    packed_data = packed_data + b'\x00\x00'
    
    return packed_data

def main():
    args = parser.parse_args()
    image = Image.open(args.filename)

    data = pack(image)
    
    out = Path(args.filename).with_suffix('.pc')
    with open(out, 'wb') as pc:
        pc.write(data)

if __name__ == "__main__":
    main()

