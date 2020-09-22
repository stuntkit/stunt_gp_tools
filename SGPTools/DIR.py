#!/usr/bin/python3 
# coding=utf-8

import os
import sys

# files
from pathlib import Path
# from glob import glob

from typing import List, Dict, Optional, BinaryIO

#binary
import struct
# from SGPTools.FileEntry import FileEntry

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
                file_folder = Path(*Path(r).parts[1:])
                # TODO fix this, won't work properly for folders not placed next to the script
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
            file_offsets: Dict[str, int] = dict()
            descriptor_offsets: Dict[str, int] = dict()

            descriptor_table: List[str] = [""] * 1024
            next_file: Dict[str, str] = dict()

            hashes: Dict[str, int] = dict()

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
                        s = next_file[filename]
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
    '''
    
    def to_dir(self, filename: str):

        # 4 bytes for magic, size, table offset
        header_size = 12

        # files_data, file_offsets = self.__prepare_data()
        # directory = self.__prepare_directory(file_offsets, offset)
        # self.files = self.files.sort()
        files_data, directory_table = self.__prepare_dir_data(header_size)
        # offset of directory table
        offset = header_size + len(files_data)
        # 12 bytes for magic, size, table offset, then just file data and table
        size = header_size + len(files_data) + len(directory_table)

        # magic, size, table offset, file data, directory table
        data = b'DIR\x1a' + size.to_bytes(4, 'little') + offset.to_bytes(4, 'little') + files_data + directory_table

        with open(filename, 'wb') as dir_file:
            dir_file.write(data)
    
    def __prepare_dir_data(self, header_size: int):
        # return files data and directory table separately to calculate offsets properly
        # TODO check if would't it be wiser to return one big binary and offset instead

        ''
        Two loops?
        Firstloop
            glue files in files_data
            add data offset to the files_data_offset table

        Second loop

        ''
        files_data_offset = []
        # index in this.files, descriptor offset¸next file offset
        # next file offset is 0 by default ans is set retrospectively by one of the next files
        files_data_meta = []

        # table of bool values if the hash is filled or not
        # TODO make this its own type, with .index and .offset fields?
        # index - index of a file
        # offset - offset of file descriptor
        directory_table = [ {'index': None, 'offset': 0} for i in range(1024)]

        # all files data glued together, with proper padding
        files_data = b''

        # hash array + file descriptors (next file offset, data offset, size, filename)
        directory_table_binary = b'\x0A\x00\x00\x00'
        data_offset = 0xC
        # next - index of next linked file
        # data - data offset
        # offset - descriptor offset
        meta = [{'next': None, 'data': 0, 'offset': 0} for i in self.files]

        descriptor_offset = 0x1004
        # first loop, glue data together and get offsets to meta table
        for i, file_entry in enumerate(self.files, start=0):
            files_data += file_entry.data + file_entry.data_padding
            meta[i]['data'] = data_offset
            data_offset = len(files_data) + 0xC

            # check if hashtable entry is empty
            if directory_table[file_entry.hash]['index'] == None:
                directory_table[file_entry.hash] = {'index': i, 'offset': descriptor_offset}
            else:
                # follow chain utntil next offset is empty
                print('merde')
                index = directory_table[file_entry.hash]['index']
                while meta[index]['next'] != None:
                    print('fuck')
                    print(meta[index])
                    index = meta[index]['next']
                # now theoretically we should have last file in chain, with empty offset, so we can add this file there
                meta[index]['next'] = i

            # next(4), data ofset(4), size(4), filename, trailing \0, padding
            meta[i]['offset'] = descriptor_offset
            descriptor_offset += 12 + len(file_entry.filename) + 1 + len(file_entry.filename_padding)

        # Now we have all required data to create binary directory table
        # TODO cahnge all range(1024) to something more dynamic, maybe somehow coupe with Hahs bits?
        for i in range (1024):
            directory_table_binary += directory_table[i]['offset'].to_bytes(4, 'little') 

        # now write all descriptors
        # convert all next None's to 0
        # TODO make this less ugly, ideally in different class

        for i in range(len(self.files)):
            # next
            if meta[i]['next'] == None:
                directory_table_binary += b'\x00\x00\x00\x00'
            else:
                print(self.files[i].filename)
                print(meta[meta[i]['next']]['offset'])
                directory_table_binary += meta[meta[i]['next']]['offset'].to_bytes(4, 'little') 
            # data offset
            directory_table_binary += meta[i]['data'].to_bytes(4, 'little') 
            # size
            directory_table_binary += self.files[i].size.to_bytes(4, 'little') 
            # filename and padding
            directory_table_binary += self.files[i].filename.encode('ascii') + b'\x00'
            directory_table_binary += self.files[i].filename_padding
        
        return (files_data, directory_table_binary)

    ''

    def __prepare_directory(self, file_offsets, directory_offset) -> bytes:
        directory = b''
        directory_file_entries = b''

        directory_hash_table = b''
        for i, file_entry in enumerate(self.files, start=0):
            if file_entry:
                directory_hash_table = directory_hash_table + (4 + 4096 + len(directory_file_entries)).to_bytes(4, 'little')
                directory_file_entries = directory_file_entries + self.__prepare_directory_file_entry(file_entry, file_offsets[i])

            else:
                directory_hash_table = directory_hash_table + b'\x00\x00\x00\x00'

        directory = b'\x0A\x00\x00\x00' + directory_hash_table + directory_file_entries
        return directory
    
    @staticmethod
    def __prepare_directory_file_entry(file_entry: FileEntry, offset: int) -> bytes:
        entry = file_entry.unknown.to_bytes(4, 'little')
        #offset
        entry = entry + offset.to_bytes(4, 'little')

        entry = entry + file_entry.size.to_bytes(4, 'little')
        name = file_entry.filename.encode('ascii') + b'\x00'
        entry = entry + name
        padding = (4 - (len(name) % 4)) % 4
        if padding:
            entry = entry + b''.join([b'\x00'] * padding)
        return entry'''
