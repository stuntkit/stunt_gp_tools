#!/usr/bin/python3 
# coding=utf-8

import os
import sys
import argparse

#import pandas as pd
#import numpy as np

# files
#import pathlib
#from glob import glob

#binary
#inport struct

class FileEntry:
    def __init__(self, unknown: int, filename: str, data: bytes):
        self.unknown = unknown
        self.filename = filename
        self.data = data
        self.__padding = b''
    
    def __str__(self):
        return self.filename+": "+str(self.size)+' bytes'
    
    def __repr__(self):
        """For array, print just filename"""
        return '"'+self.filename+'"'
    
    @property
    def size(self):
        """We don't store file length explicitly, there's no reason (unless we would want to break things"""
        return len(self.__data)
        
    @property
    def data(self):
        return self.__data
        
    @data.setter
    def data(self, value):
        self.__data = value
        self.__padding = b''.join([b'\x00'] * ((4 - (len(self.data) % 4)) % 4))
    
    @property
    def padding(self):
        return self.__padding
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
