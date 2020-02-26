#!/usr/bin/python3 
# coding=utf-8

# LIMITATIONS:
# Can only output .png files, converting 16-bit color to 32bpp

import argparse
import sys
from pathlib import Path

from PIL import Image, ImageDraw
import struct

parser = argparse.ArgumentParser(description='Stunt GP Texture converter', epilog='This script can convert Stunt GP .pc texture files to .png') # and vice versa')
parser.add_argument("filename", help="name of the texture file")
parser.add_argument("-o", "--output", help="name of the output file")
parser.add_argument("-v", "--verbose", help="increase output verbosity", action="store_true")


def to_bytes(number, bytes = 2):
    return number.to_bytes(bytes, 'little')

def decompress(packed_data):
    unpacked_data = b''
    i = 0
    
    # as everything, read data by 2 bytes
    packed_data = memoryview(packed_data).cast('H')
    current_word = packed_data[i]
    
    #copy first pixel to output
    unpacked_data = unpacked_data + to_bytes(current_word)
    i = i + 1
    
    while True:
        current_word = packed_data[i]
        
        # just stream pixels while 15th bit set.
        # It is also used for alpha channel, so we know for a fact that these pixels are opaque
        while (current_word >> 15) & 1 == 1:
            i = i + 1
            unpacked_data = unpacked_data + to_bytes(current_word)
            current_word = packed_data[i]
        
        i = i + 1
        
        # is 14th bit set
        if (current_word >> 14) & 1 == 1:
            repeats = ((current_word>>11) & 7) + 1
            location = ( current_word & 0x7FF ) * 2
            head = len(unpacked_data)
            
            if repeats <= 8:
                for j in range(repeats+1):
                    #repeat two bytes from output
                    unpacked_data = unpacked_data + to_bytes(unpacked_data[head+(2*j)-location-2], 1)+ to_bytes(unpacked_data[head+(2*j)-location-1], 1)

        else:
            # get 13 bits from last pixel
            current_word = (current_word & 0x3FFF) * 4

            if current_word == 0:
                break

            fill = (current_word ) & 3
            current_word = current_word >> 2

            for j in range(current_word):
                unpacked_data = unpacked_data + to_bytes(0)

            if fill == 2:
                unpacked_data = unpacked_data + b'\x00\x00'
    
    return unpacked_data

def unpack(filename):
    with open(filename, 'rb') as pc_file:
        pc_header = pc_file.read(4)
        if pc_header != b'TM!\x1a':
            sys.exit('unknown magic')
        
        pc_unknown = pc_file.read(2)
        
        # width/height has 2 bytes, so max resolution is 65 535 (2^16 - 1)
        # although the game uses way smaller images
        pc_width = pc_file.read(2)
        pc_width = struct.unpack("<H",pc_width)[0]
        
        pc_height = pc_file.read(2)
        pc_height = struct.unpack("<H",pc_height)[0]
        
        #each pixel is stored on 16bits
        pc_data = decompress(pc_file.read())

        image = {'unknown': pc_unknown, 'width': pc_width, 'height': pc_height, 'data': pc_data}
        return image

def convert_color(data):
    pixels = []
    if len(data) % 2 != 0:
        sys.exit('wrong number of data for color converter')

    #each pixel is stored on 2 bytes, in A1 R5 G5 B5  format
    pixels_data = memoryview(data).cast('H')
    for pixel in pixels_data:
        # 5 bits means 2^5-1 = 31
        red =  int((pixel >> 10 &   31) * (255/31))
        green = int((pixel >> 5 &  31) * (255/31))
        blue = int( (pixel &  31) * (255/31))
        alpha = int((pixel >15 & 1) * 255)
        
        pixels.append((red, green, blue, alpha))
    return pixels

def main():
    args = parser.parse_args()

    data = unpack(args.filename)
    
    if args.verbose:
        print('Image size:',str(data['width'])+'x'+str(data['height']))
        if data['width']*data['height'] != len(data['data'])/2:
            print('data length mismatch, got', len(data['data'])/2,'but expected',data['width']*data['height'])
    
    img = Image.new('RGBA', (data['width'], data['height']))
    pixels = convert_color(data['data'])
    img.putdata(pixels)

    output = Path(args.filename).with_suffix('.png')
    if args.output:
        output = args.output
    
    img.save(output, 'PNG')
    if args.verbose:
        print('Saved to',output)

if __name__ == "__main__":
    args = parser.parse_args()
    main()
