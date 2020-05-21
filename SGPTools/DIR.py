#!/usr/bin/python3 
# coding=utf-8

import os
import sys
import argparse

#import pandas as pd
#import numpy as np

# files
import csv
from pathlib import Path
from glob import glob

#binary
import struct
from SGPTools.FileEntry import FileEntry

class DIR():
    def __init__(self):
        self.__magic = b'DIR\x1a'
        self.files = [None for i in range(1024)] #FileEntry()
    
    @classmethod
    def from_dir(cls, filename: str):
        new_dir = cls()
        with open(filename, 'rb') as dir_file:
            new_dir.__magic = dir_file.read(4)
            size = struct.unpack("<L", dir_file.read(4))[0]
            offset = struct.unpack("<L", dir_file.read(4))[0]
            dir_file.seek(offset + 4)
            j=0
            for i in range(1024):

                descriptor_offset = struct.unpack("<L", dir_file.read(4))[0]
                if descriptor_offset != 0:
                    j = j +1
                    dir_file.seek(offset + descriptor_offset)
                    
                    file_unknown = struct.unpack("<L", dir_file.read(4))[0]
                    file_offset = struct.unpack("<L", dir_file.read(4))[0]
                    file_size = struct.unpack("<L", dir_file.read(4))[0]
                    file_name = b''
                    while True:
                        file_name = file_name + dir_file.read(4)
                        if file_name[-1:] == b'\x00': 
                            break
                    #TODO finish; also does file need offset? Can't we just calculate?
                    #file_data = read
                    #
                    file_name = file_name.replace(b'\x00', b'').decode('ascii') #('utf8')
                    print(str(i)+"\t"+file_name)
                    dir_file.seek(file_offset)
                    file_data = dir_file.read(file_size)

                    new_dir.files[i] = FileEntry(file_unknown, file_name, file_data)

                dir_file.seek(offset + 4 + (4 * (i + 1) ))
            print(j)
        return new_dir
    
    #TODO
    @classmethod
    def from_folder(cls, folder: str):
        new_dir = cls()
        
        folder = Path(folder)
        '''for r, d, f in os.walk(folder):
            for filename in f:
                file_folder = Path(*Path(r).parts[1:])
                print(file_folder/filename)
                new_dir.files[self.__calculate_hash(file_folder/filename)] = FileEntry(int(row['unknown']), row['filename'], tmp_data)'''
        with open(folder/'dir.csv') as descriptor_file:
            csv_reader = csv.DictReader(descriptor_file)#, fieldnames=['id', 'filename', 'unknown'],)
            
            for row in csv_reader:
                tmp_path = Path(row['filename'].replace('\\', '/'))
                tmp_data = b''
                with open(folder.joinpath(tmp_path), 'rb') as tmp_file:
                    tmp_data = tmp_file.read()
                file_hash = new_dir.__calculate_hash(row['filename'])
                new_dir.files[file_hash] = FileEntry(int(row['unknown']), row['filename'], tmp_data)
        
        return new_dir
    
    
    def to_dir(self, filename: str):

        files_data, file_offsets = self.__prepare_data()
        offset = (len(files_data)+12) #self.__calculate_data_size().to_bytes(4, 'little')


        directory = self.__prepare_directory(file_offsets, offset)

        size = 12 + len(files_data) + len(directory)
        data = self.__magic + size.to_bytes(4, 'little') + offset.to_bytes(4, 'little') + files_data + directory
        with open(filename, 'wb') as dir_file:
            dir_file.write(data)
    
    
    def to_folder(self, folder):
        path = Path(folder)
        path.mkdir(parents=True, exist_ok=True)
        descriptors = []
        for i, file_entry in enumerate(self.files, start=0):
            if file_entry:
                tmp_filename = Path(path/file_entry.filename.replace('\\', '/'))
                tmp_filename.parent.mkdir(parents=True, exist_ok=True)
                with open(tmp_filename, 'wb') as tmp_file:
                    tmp_file.write(file_entry.data)
                descriptors.append({'filename': file_entry.filename, 'unknown': file_entry.unknown})

        with open(path/'dir.csv', 'w') as descriptor_file:
            #descriptor_file.write('tmp')
            wr = csv.DictWriter(descriptor_file, fieldnames=['filename', 'unknown'], quoting=csv.QUOTE_NONNUMERIC)
            wr.writeheader()
            wr.writerows(descriptors)
    

    def __prepare_data(self) -> bytes:
        data = b''
        offsets = [None for i in range(1024)]
        for i, file_entry in enumerate(self.files, start=0):
            if file_entry:
                offsets[i] = len(data)+12
                data = data + file_entry.data + b'\x1a\x00'
                padding = (4 - ((file_entry.size + 2) % 4)) % 4
                if padding % 4 != 0:
                    data = data + b''.join([b'\x00' * padding])
        
        return data, offsets

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
        return entry

    @staticmethod
    def __calculate_hash(filename):
        filename = filename.encode("ascii", errors="ignore").decode()
        hash_bits = 10
        hash_size = 1 << hash_bits
        
        hash_calculated = 0
        for char in filename:
            hash_calculated = ((hash_calculated << 1) % hash_size) | (hash_calculated >> (hash_bits - 1) & 1)
            # sum = ((sum << 1) % HASH_SIZE) | (sum >> (HASH_BITS - 1) & 1);
            hash_calculated = hash_calculated + ord(char)
            hash_calculated = hash_calculated % hash_size
        return hash_calculated
