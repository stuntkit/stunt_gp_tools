#!/usr/bin/python3 
# coding=utf-8

class FileEntry:
    def __init__(self, filename: str, data: bytes):
        self.filename = filename
        # self.filename = filename.encode("ascii", errors="ignore").decode()
        # self.__filename_padding = b''.join([b'\x00'] * ((4 - ((len(self.filename) + 1) % 4)) % 4))
        self.data = data
        # self.__padding = b'\x1a\x00' + b''.join([b'\x00'] * ((4 - ((len(self.data) + 2) % 4)) % 4))
    
    def __str__(self) -> str:
        return self.filename+": "+str(self.size)+' bytes'
    
    def __repr__(self) -> str:
        """For array, print just filename"""
        return '\n"'+self.filename+'"'

    def __lt__(self, other):
        return self.filename < other.filename
    
    @property
    def size(self):
        """We don't store file length explicitly, there's no reason (unless we would want to break things"""
        return len(self.__data)
        
    @property
    def data(self) -> bytes:
        return self.__data
        
    @data.setter
    def data(self, value: bytes):
        self.__data = value
        self.__padding = b'\x1a\x00' + b''.join([b'\x00'] * ((4 - ((len(self.data) + 2) % 4)) % 4))
    
    @property
    def data_padding(self):
        return b'\x1a\x00' + b''.join([b'\x00'] * ((4 - ((len(self.data) + 2) % 4)) % 4))

    @property
    def filename_padding(self):
        return b''.join([b'\x00'] * ((4 - ((len(self.filename) + 1) % 4)) % 4))

    @property
    def hash(self):
        HASH_BITS = 10
        hash_size = 1 << HASH_BITS
        
        hash_calculated = 0
        for char in self.filename:
            hash_calculated = ((hash_calculated << 1) % hash_size) | (hash_calculated >> (HASH_BITS - 1) & 1)
            hash_calculated = hash_calculated + ord(char)
            hash_calculated = hash_calculated % hash_size
        return hash_calculated
