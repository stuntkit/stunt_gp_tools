#!/usr/bin/python3 
# coding=utf-8

import os
import sys

#import pandas as pd
#import numpy as np

# files
import pathlib
#from glob import glob

from PIL import Image # type: ignore

#binary
import struct

class PC():
    def __init__(self):
        self.__magic = b'TM!\x1A'
        self.__unknown = b'\x03\x00'
        self.__width = 0
        self.__height = 0
        self.__data = 0
    
    @classmethod
    def from_pc(cls, filename: str):
        new_pc = cls()
        with open(filename, 'rb') as pc_file:
            new_pc.__magic = pc_file.read(4)
            new_pc.__unknown = pc_file.read(2)
            new_pc.__width = struct.unpack("<H",pc_file.read(2))[0]
            new_pc.__height = struct.unpack("<H",pc_file.read(2))[0]
            
            new_pc.__data = new_pc.__unpack(pc_file.read())
            
        if new_pc.check():
            print('ok')
            return new_pc
    
    def from_png(self, filename: str):
        pass
   
    
    def to_pc(self, filename: str):
        pass
    
    def to_png(self, filename: str):
        print('saving to', filename)
        image = Image.new('RGBA', (self.__width, self.__height))
        #Image.new('RGBA', (data['width'], data['height']))
        image.putdata(self.__convert_color(self.__data))
        image.save(filename, 'PNG')
    
    def check(self) -> bool:
    
        try:
            assert self.__magic == b'TM!\x1A', 'wrong magic'

            assert self.__width > 0, 'width <= 0'
            assert self.__width <= 65536, 'width > 65536'

            assert self.__height > 0, 'height <= 0'
            assert self.__height <= 65536, 'height > 65536'

        except AssertionError as error:
            print('assertion error:', error)
            return False
        
        try:
            assert self.__unknown == b'\x03\x00', 'wrong unknown'
        except AssertionError as error:
            print('assertion suggestion:', error)
        
        return True
    
    def __unpack(self, packed_data: bytes):
        unpacked_data = b''
        i = 0
        
        # as everything, read data by 2 bytes
        packed_data = memoryview(packed_data).cast('H')
        current_word = packed_data[i]
        
        #copy first pixel to output
        unpacked_data = unpacked_data + current_word.to_bytes(2, 'little')
        i = i + 1
        
        while True:
            current_word = packed_data[i]
            
            # just stream pixels while 15th bit set.
            # It is also used for alpha channel, so we know for a fact that these pixels are opaque
            while (current_word >> 15) & 1 == 1:
                i = i + 1
                unpacked_data = unpacked_data + current_word.to_bytes(2, 'little')
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
                        unpacked_data = unpacked_data + unpacked_data[head+(2*j)-location-2].to_bytes(1, 'little') + unpacked_data[head+(2*j)-location-1].to_bytes(1, 'little')

            else:
                # get 13 bits from last pixel
                current_word = (current_word & 0x3FFF) * 4

                if current_word == 0:
                    break

                fill = (current_word ) & 3
                current_word = current_word >> 2

                for j in range(current_word):
                    unpacked_data = unpacked_data + b'\x00\x00'

                if fill == 2:
                    unpacked_data = unpacked_data + b'\x00\x00'
        
        return unpacked_data
        
    def __pack(self, data: bytes):
        pass

    @staticmethod
    def __convert_color(data):
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
'''
class C:
    def __init__(self):
        self._x = None

    @property
    def x(self):
        """I'm the 'x' property."""
        return self._x

    @x.setter
    def x(self, value):
        self._x = value

    @x.deleter
    def x(self):
        del self._x
'''
