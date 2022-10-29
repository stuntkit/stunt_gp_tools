#!/usr/bin/python3 
# coding=utf-8

import os

# files
from pathlib import Path
# from glob import glob

from typing import List, Dict, Optional, BinaryIO

#binary
import struct

class DIR(dict):

    @classmethod
    def from_dir(cls, filename:str):
        dir = cls()
        with open(filename, 'rb') as dir_file:
            magic = dir_file.read(4) # magic
            if magic != b'DIR\x1a':
                raise RuntimeError("Not a DIR file!")

            archive_size = struct.unpack("<L", dir_file.read(4))[0]
            table_offset = struct.unpack("<L", dir_file.read(4))[0]

            #jump to hash array
            dir_file.seek(table_offset)
            table_header = dir_file.read(4)
            if table_header != b'\x0a\x00\x00\x00':
                raise RuntimeError("Not a valid table header!")

            for i in range(1024):
                 dir_file.seek(table_offset + 4 + (4 * i))

                 element_offset = struct.unpack("<L", dir_file.read(4))[0]
                 if element_offset:
                     dir.__read_file(dir_file, table_offset, element_offset)
        
        return dir

    def __read_file(self, dir_file: BinaryIO, table_offset: int, element_offset: int):
        dir_file.seek(table_offset + element_offset)
        
        next_offset = struct.unpack("<L", dir_file.read(4))[0]
        data_offset = struct.unpack("<L", dir_file.read(4))[0]
        data_size = struct.unpack("<L", dir_file.read(4))[0]

        filename = ''
        byte = dir_file.read(1)
        while byte != b"\x00":
            filename += byte.decode('ascii')
            byte = dir_file.read(1)
        
        dir_file.seek(data_offset)
        data = dir_file.read(data_size)

        self[filename] = data

        # read another file with the same hash if exists
        if next_offset:
            self.__read_file(dir_file, table_offset, next_offset)
        
    def to_folder(self, folder_name: str):
        path = Path(folder_name)
        path.mkdir(parents=True, exist_ok=True)

        for filename, data in self.items():
            # TODO omit slash conversion for Windows
            tmp_filename = Path(path/filename.replace('\\', '/'))
            tmp_filename.parent.mkdir(parents=True, exist_ok=True)
            with open(tmp_filename, 'wb') as tmp_file:
                tmp_file.write(data)
    
    @classmethod
    def from_folder(cls, folder_name: str):
        dir = cls()
        folder = Path(folder_name)
        for r, d, f in os.walk(folder):
            for filename in f:
                file_folder = Path(r).relative_to(folder)
                filename_in_archive = str(file_folder/filename)
                # It's weird how all systems except Windows uses slash
                filename_in_archive = filename_in_archive.replace('/', '\\')

                with open(Path(r)/filename, 'rb') as data:
                    print('FiA {}'.format(filename_in_archive))
                    dir[filename_in_archive] = data.read()
        return dir

    def to_dir(self, filename: str):
        with open(filename, "wb") as dir_file:
            # note - this is just rerite of C++ code that I made later, this code should have improved quality, although I haven't put much thought into polishing it

            # TODO consolidate all these dicts
            file_offsets: Dict[str, int] = {}
            descriptor_offsets: Dict[str, int] = {}

            descriptor_table: List[str] = [""] * 1024
            next_file: Dict[str, str] = {}

            hashes: Dict[str, int] = {}

            # where is directory table header
            table_offset = 12
            descriptor_offset = 4 + (4 * 1024)

            for filename, data in self.items():
                file_offsets[filename] = table_offset
                table_offset += len(data) + 2
                # padding and all that jazz
                table_offset += table_offset % 4

                descriptor_offsets[filename] = descriptor_offset
                descriptor_offset += 12 + len(filename) + 1
                descriptor_offset += descriptor_offset % 4

                hashes[filename] = self.get_hash(filename)

            for filename, data in self.items():
                hash = self.get_hash(filename)
                if descriptor_table[hash] == "":
                    descriptor_table[hash] = filename
                else:
                    s = descriptor_table[hash]
                    while s in next_file:
                        s = next_file[s]
                    next_file[s] = filename

            archive_size = table_offset + descriptor_offset

            #write portion here
            dir_file.write(b"DIR\x1A")
            dir_file.write(archive_size.to_bytes(4, 'little'))
            dir_file.write(table_offset.to_bytes(4, 'little'))

            for filename, data in self.items():
                dir_file.write(data)

                dir_file.write(b"\x1A\x00")
                dir_file.write(b"".join([b"\x00"] * ((file_offsets[filename] + len(data) + 2) % 4)))
            
            # directory table
            dir_file.write(b"\x0A\x00\x00\x00")
            for i in range(1024):
                if descriptor_table[i] != "":
                    dir_file.write(descriptor_offsets[descriptor_table[i]].to_bytes(4, 'little'))
                else:
                    dir_file.write(b"\x00\x00\x00\x00")
            
            for filename, data in self.items():
                if filename in next_file:
                    dir_file.write(descriptor_offsets[next_file[filename]].to_bytes(4, 'little'))
                else:
                    dir_file.write(b"\x00\x00\x00\x00")

                dir_file.write(file_offsets[filename].to_bytes(4, 'little'))
                dir_file.write(len(data).to_bytes(4, 'little'))
                dir_file.write(filename.encode('ascii'))
                dir_file.write(b"\x00")
                dir_file.write(b"".join([b"\x00"] * ((descriptor_offsets[filename] + 12 + len(filename) + 1) % 4)))




    @staticmethod
    def get_hash(filename: str) -> int:
        hash = 0

        HASH_BITS = 10
        hash_size = 1 << HASH_BITS
        
        for char in filename:
            # bitise rotate left on 10 bits
            hash = ((hash << 1) % hash_size) | (hash >> (HASH_BITS - 1) & 1)
            hash = (hash + ord(char)) % hash_size
        return hash
