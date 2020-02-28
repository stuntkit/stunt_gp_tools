
function pathPaddingUpdate(root) {
    var pad = 4 - (this.parent.path.byteCount % 4);
    if((pad % 4) != 0) {
        this.padding.length = pad;
    }
}

function init() {
    var file = struct({
        data: array(uint8(), function(root) {return this.parent.parent.parent.size.value; }),
    });
    file.name = "file";
    
    //ugly hack, but works
    var pathPadding = struct({
        padding: array(uint8(), 0),
    });
    pathPadding.updateFunc = pathPaddingUpdate;
    
    var fileMeta = struct({
        unknown: uint32(),
        offset: pointer(uint32(), file),
        size: uint32(),
        path: string('ascii'),
        padding: pathPadding,
    });
    fileMeta.name = "file meta";
    
    fileMeta.validationFunc = function() {
        var valid = true;
        if(this.offset.value == 0) {
            this.offset.validationError = "empty offset";
        }
        else if(this.size.value == 0) {
            this.size.validationError = "empty size";
        }
        else if(this.path.byteCount % 4 != 0) {
            this.validationError = "Path is not divisible by 4";
            valid = false;
        }
        
        return valid;
    };
    
    
    var footer = struct({
        footerHeader: uint32(),
        unknown: array(uint32(), 1024),
        //TODO fix this, will show files, but will also show lots of empty entries and EOF error
        //I don't knwo how to make Okteta read until EOF and then just end
        Files: array(fileMeta, 1024), 
    });
    footer.name = "footer";
    
    var wadFile = struct({
        Magic: array(char(),4),
        WADSize: uint32(),
        WADOffset: pointer(uint32(), footer),      
    });

    wadFile.defaultLockOffset = '0';
    wadFile.byteOrder = "little-endian";
    
    wadFile.validationFunc = function() {
        var valid = true;

        if(this.Magic != ['D', 'I', 'R', 0x1A] ) {
            this.Magic.validationError = "Invalid magic";
            valid = false;
        }
        
        else if(this.WADSize.value < this.WADOffset.value) {
            this.WADSize.validationError = "WAD footer offset is beyond end of the file";
            valid = false;
        }
        
        return valid;
    };
    
    wadFile.name = "SGP WAD";
    return wadFile;
}

