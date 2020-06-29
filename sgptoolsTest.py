#!/usr/bin/python3 
# coding=utf-8

import os
import sys
import argparse

#from SGPTools import pc, WAD
import SGPTools

#import pandas as pd
#import numpy as np

# files
#import pathlib
#from glob import glob

#binary
#inport struct

class App:
    def __init__(self):
        pass

    @staticmethod
    def to_folder(wad_name):
        wad = SGPTools.DIR.from_dir(wad_name)
        if wad:
            print('to folder... ', end='')
            wad.to_folder('catal')
            # wad.to_dir('logos_new.wad')
            print('done')
    
    @staticmethod
    def from_folder_to_wad(directory):
        wad = SGPTools.DIR.from_folder(directory)
        if wad:
            print('to dir... ', end='')
            wad.to_dir('test.wad')
            print('done')

    def main(self):
        #pc = SGPTools.PC.from_pc("grs.pc")
        #if pc:
        #    pc.to_png("test.png")
        wad = SGPTools.DIR.from_dir('wads/p_catal.wad')

        for file_entry in wad.files:
            print("{}\t{}".format(file_entry.filename, file_entry.hash))
        wad.to_dir('test.dir')

        #self.to_folder('wads/p_catal.wad')
        #self.to_wad('catal')
        
        #wad = SGPTools.DIR.from_folder('logos')
        
        #if wad:
            #print('saving')
            #wad.to_dir('logos_broken.wad')


if __name__ == "__main__":
    if 'argparse' in sys.modules:
        print('argparser found')
        parser = argparse.ArgumentParser()
        #parser.add_argument("filename", help="name of the image file")
        args = parser.parse_args()
        
    app = App()
    app.main()

