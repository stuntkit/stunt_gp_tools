def __prepare_dir_data_olde(self, header_size: int):
        files_data = b''
        directory_table: List[List[int]] = [ [0, 0] for i in range (1024)] # how I am supposed to do typing here??
        directory_table_binary = b'\x0A\x00\x00\x00'
        # [descriptor offset (hash), data offset, ]
        # directory_offsets = []
        descriptors_length = 0
        files_descriptors = [[0, 0, 0] for i in range(len(self.files))]

        for i, file_entry in enumerate(self.files, start=0):
            files_data = files_data + file_entry.data +  file_entry.padding #b'\x1a\x00' +

            # [data offset, descriptor offset]
            # directory_offsets.append([header_size + len(files_data), 4100 + descriptors_length])
            # next file, data offset, descriptor offset, next file id
            files_descriptors[i] = [0, 0, header_size + len(files_data), descriptors_length]

            file_descriptor_length = 12 + len(file_entry.filename) + 1 + len(file_entry.filename_padding)
            if directory_table[file_entry.hash] == [0, 0]:
                directory_table[file_entry.hash] = [descriptors_length, i]
            else:
                if file_entry.hash == 65:
                    print('fuck me ')
                print(file_entry.hash)
                # TODO error is here
                # find file pointed by directory table
                id = directory_table[file_entry.hash][1]

                while True:
                    id = files_descriptors[id][1]
                    if files_descriptors[id][0] == 0:
                        break
                pass

                files_descriptors[id][0] = descriptors_length
                files_descriptors[id][1] = i
            
            descriptors_length = descriptors_length + file_descriptor_length
        print("table")
        print(directory_table)
        for directory_logic in directory_table:
            if directory_logic != [0, 0]:
                directory_table_binary = directory_table_binary + (4100 + directory_logic[0]).to_bytes(4, 'little')
            else:
                directory_table_binary = directory_table_binary + b'\x00\x00\x00\x00'
        
        print("fuckery")
        for i, file_entry in enumerate(self.files, start=0):
                        # ??                     ????
            # next, offset, size, name
            descriptor = b''
            if files_descriptors[i][0] != 0:
                print(file_entry.filename+'\t'+str(file_entry.hash)+'\t'+ str(4100 + files_descriptors[i][0]))
                descriptor = descriptor + (4100 + files_descriptors[i][0]).to_bytes(4, 'little') 
            else:
                descriptor = descriptor + b'\x00\x00\x00\x00'
            descriptor = descriptor + files_descriptors[i][2].to_bytes(4, 'little') #data offset
            descriptor = descriptor + file_entry.size.to_bytes(4, 'little')
            descriptor = descriptor + file_entry.filename.encode('ascii') + b'\x00' + file_entry.filename_padding

            directory_table_binary = directory_table_binary + descriptor

        return (files_data, directory_table_binary)