#!/usr/bin/python3 
# coding=utf-8

import os
import sys
import struct
import glob

import pathlib

def read_data(filename):
    with open(filename, 'rb') as wad_file:
        wad_magic = wad_file.read(4)
        if wad_magic != b'DIR\x1A':
            sys.exit('unknown magic')
        
        wad_size = wad_file.read(4)
        wad_size = struct.unpack("<L",wad_size)[0]
        if wad_size != os.path.getsize(filename):
            sys.exit('size of file is different than size in signature')
        
        #that descriptor or whatever, Unknown block of data that spans for 4100 bytes
        wad_offset = wad_file.read(4)
        wad_offset = struct.unpack("<L",wad_offset)[0]
        
        wad_file.seek(wad_offset)
        tmp = wad_file.read(4)
        if tmp[:1] != b'\x0a':
            print('0x0A not found, no matter as I cannot parse that yet :/')
            
        wad_info = {'size': wad_size, 'offset':wad_offset}
        
        
        #FILE DESCRIPTORS
        #I'm skipping 4100 bytes of something, I cannot determine what
        wad_file.seek(wad_offset+4100)
        files = []
        while True:
            #unknown, for most of the time empty, but see boot.wad, just after boot1d.pc
            file_unknown = wad_file.read(4)
            if file_unknown == b'':
                break
                
            file_offset = wad_file.read(4)

            file_offset = struct.unpack("<L",file_offset)[0]
            
            file_size = wad_file.read(4)
            file_size = struct.unpack("<L",file_size)[0]
            
            #filenames read by 4 bytes at a time, padded with 0x00
            name = b''
            while True:
                name = name + wad_file.read(4)
                if name[-1:] == b'\x00': 
                    break
            name = name.replace(b'\x00', b'')
            name = name.decode('utf8')
            files.append({'name': name, 'offset': file_offset, 'size': file_size, 'unknown': file_unknown})
            #print(name, 'at', str(hex(file_offset))+',', file_size, 'bytes')
        wad_data = {'info':wad_info, 'files':files}
        return wad_data

def extract_files(filename):
    wad_meta = read_data(filename)
    print('got', len(wad_meta['files']), 'files')
    with open(filename, 'rb') as wad_file:
        for packed_file in wad_meta['files']:
            wad_file.seek(packed_file['offset'])
            out = 'out/' + packed_file['name'].replace('\\', '/')
            pathlib.Path(out).parent.mkdir(parents=True, exist_ok=True)
            with open(out, 'wb') as extracted_file:
                extracted_file.write(wad_file.read(packed_file['size']))
            print(packed_file['name'])

def main():
    print ("read files from Stunt GP .wad files, expects wthem in `wads` folder")
    names = glob.glob('wads/*.wad')
    for name in names:
        print(name)
        #files = read_data(name)
        extract_files(name)
        print('')

if __name__ == "__main__":
    main()
