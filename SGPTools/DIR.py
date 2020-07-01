#!/usr/bin/python3 
# coding=utf-8

import os
import sys

# files
from pathlib import Path
# from glob import glob

from typing import List, Dict, Optional

#binary
import struct
from SGPTools.FileEntry import FileEntry

class DIR():
    def __init__(self):
        # TODO make class out of this or at least add there aren't two files with the same name (case insensitive)
        self.files = []

    @classmethod
    def from_dir(cls, filename: str):
        new_dir = cls()
        with open(filename, 'rb') as dir_file:
            magic = dir_file.read(4) # magic
            if magic != b'DIR\x1a':
                return None
            size = struct.unpack("<L", dir_file.read(4))[0]
            table_offset = struct.unpack("<L", dir_file.read(4))[0]

            #jump to hash array
            dir_file.seek(table_offset + 4)
            # read hash array
            for i in range(1024):
                # jump back to hash table
                dir_file.seek(table_offset + 4 + (4 * i))
                # read file offset
                descriptor_offset = struct.unpack("<L", dir_file.read(4))[0]
                # if file(s) with that hash exists:
                if descriptor_offset != 0:
                    new_dir.__read_files_from_table(dir_file, table_offset, descriptor_offset)
        
        return new_dir

    # read file descriptor and file itself, as well as all its children
    def __read_files_from_table(self, dir_file, table_offset, descriptor_offset):
        # jump to file
        dir_file.seek(table_offset + descriptor_offset)
        
        # offset of next file with same hash
        file_next = struct.unpack("<L", dir_file.read(4))[0]
        file_offset = struct.unpack("<L", dir_file.read(4))[0]
        file_size = struct.unpack("<L", dir_file.read(4))[0]

        # read file name, it's always aligned to 4 bytes, so we can read it this way
        file_name = b''
        while True:
            file_name = file_name + dir_file.read(4)
            #if last char is 00, then end reading
            if file_name[-1:] == b'\x00': 
                break
        # cleanup name from 00 and convert to string
        file_name = file_name.replace(b'\x00', b'').decode('ascii')

        # jump to file data and read it
        dir_file.seek(file_offset)
        file_data = dir_file.read(file_size)

        self.files.append(FileEntry(file_name, file_data))

        # read another file with that hash if exists
        if file_next != 0:
            self.__read_files_from_table(dir_file, table_offset, file_next)

    @classmethod
    def from_folder(cls, folder: str):
        new_dir = cls()
        
        folder = Path(folder)
        for r, d, f in os.walk(folder):
            for filename in f:
                file_folder = Path(*Path(r).parts[1:])
                filename_in_archive = str(file_folder/filename)
                # Linux, ladies and gentlemen
                filename_in_archive = filename_in_archive.replace('/', '\\')

                with open(Path(r)/filename, 'rb') as data:
                    new_dir.files.append(FileEntry(filename_in_archive, data.read()))
        return new_dir
    
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

        '''
        Two loops?
        Firstloop
            glue files in files_data
            add data offset to the files_data_offset table

        Second loop

        '''
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

    def to_folder(self, folder):
        path = Path(folder)
        path.mkdir(parents=True, exist_ok=True)
        for i, file_entry in enumerate(self.files, start=0):
            if file_entry:
                tmp_filename = Path(path/file_entry.filename.replace('\\', '/'))
                tmp_filename.parent.mkdir(parents=True, exist_ok=True)
                with open(tmp_filename, 'wb') as tmp_file:
                    tmp_file.write(file_entry.data)
   

    '''

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
