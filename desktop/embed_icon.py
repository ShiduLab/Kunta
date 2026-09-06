from pathlib import Path
import struct, subprocess, sys, shutil, os

exe=Path(sys.argv[1])
ico_path=Path(sys.argv[2])
out=Path(sys.argv[3])

def align(x,a): return (x+a-1)//a*a

def parse_pe(path):
    b=bytearray(path.read_bytes())
    pe=struct.unpack_from('<I',b,0x3c)[0]
    assert b[pe:pe+4]==b'PE\0\0'
    coff=pe+4
    nsects=struct.unpack_from('<H',b,coff+2)[0]
    optsz=struct.unpack_from('<H',b,coff+16)[0]
    opt=coff+20
    magic=struct.unpack_from('<H',b,opt)[0]
    assert magic==0x20b
    image_base=struct.unpack_from('<Q',b,opt+24)[0]
    sect_align=struct.unpack_from('<I',b,opt+32)[0]
    file_align=struct.unpack_from('<I',b,opt+36)[0]
    sec_table=opt+optsz
    sections=[]
    for i in range(nsects):
        o=sec_table+i*40
        name=b[o:o+8].split(b'\0',1)[0].decode('ascii','ignore')
        vsize,vaddr,rsize,roff=struct.unpack_from('<IIII',b,o+8)
        ch=struct.unpack_from('<I',b,o+36)[0]
        sections.append((name,vsize,vaddr,rsize,roff,ch))
    max_end=max(vaddr+max(vsize,rsize) for name,vsize,vaddr,rsize,roff,ch in sections if ch & 0x20000000 or ch & 0x40000000 or ch & 0x80000000)
    next_rva=align(max_end,sect_align)
    return dict(pe=pe,coff=coff,opt=opt,optsz=optsz,image_base=image_base,sect_align=sect_align,file_align=file_align,nsects=nsects,next_rva=next_rva)

def build_rsrc(ico_path, section_rva, lang=1033):
    ico=ico_path.read_bytes()
    reserved, typ, count=struct.unpack_from('<HHH',ico,0)
    if (reserved,typ)!=(0,1): raise ValueError('Not an icon ICO')
    entries=[]
    for i in range(count):
        off=6+i*16
        bW,bH,bC,bR,wP,wB,size,imgOff=struct.unpack_from('<BBBBHHII',ico,off)
        entries.append((bW,bH,bC,bR,wP,wB,size,imgOff,ico[imgOff:imgOff+size]))
    off_root=0
    off_icon_type=32
    off_group_type=off_icon_type+16+count*8
    off_icon_names=off_group_type+24
    icon_name_offsets=[off_icon_names+i*24 for i in range(count)]
    off_group_name=off_icon_names+count*24
    off_icon_data_entries=off_group_name+24
    icon_data_entry_offsets=[off_icon_data_entries+i*16 for i in range(count)]
    off_group_data_entry=off_icon_data_entries+count*16
    off_blobs=align(off_group_data_entry+16,4)
    grp=bytearray(struct.pack('<HHH',0,1,count))
    for i,(bW,bH,bC,bR,wP,wB,size,imgOff,imgData) in enumerate(entries,1):
        grp += struct.pack('<BBBBHHIH',bW,bH,bC,bR,wP,wB,size,i)
    off_group_blob=off_blobs
    cursor=align(off_group_blob+len(grp),4)
    icon_blob_offsets=[]
    for e in entries:
        icon_blob_offsets.append(cursor)
        cursor=align(cursor+e[6],4)
    buf=bytearray(cursor)
    def write_dir(off, items):
        struct.pack_into('<IIHHHH',buf,off,0,0,0,0,0,len(items))
        for j,(rid,target,isdir) in enumerate(items):
            struct.pack_into('<II',buf,off+16+j*8,rid,(0x80000000 if isdir else 0)|target)
    write_dir(off_root,[(3,off_icon_type,True),(14,off_group_type,True)])
    write_dir(off_icon_type,[(i+1,icon_name_offsets[i],True) for i in range(count)])
    write_dir(off_group_type,[(1,off_group_name,True)])
    for i,o in enumerate(icon_name_offsets): write_dir(o,[(lang,icon_data_entry_offsets[i],False)])
    write_dir(off_group_name,[(lang,off_group_data_entry,False)])
    for i,e in enumerate(entries): struct.pack_into('<IIII',buf,icon_data_entry_offsets[i],section_rva+icon_blob_offsets[i],e[6],0,0)
    struct.pack_into('<IIII',buf,off_group_data_entry,section_rva+off_group_blob,len(grp),0,0)
    buf[off_group_blob:off_group_blob+len(grp)]=grp
    for i,e in enumerate(entries):
        o=icon_blob_offsets[i]; buf[o:o+e[6]]=e[8]
    return bytes(buf)

info=parse_pe(exe)
rsrc=build_rsrc(ico_path,info['next_rva'])
rsrcfile=out.with_suffix('.rsrcbin')
rsrcfile.write_bytes(rsrc)
# add .rsrc at explicit VMA
subprocess.check_call(['objcopy','--add-section',f'.rsrc={rsrcfile}','--set-section-flags','.rsrc=alloc,load,readonly,data','--change-section-vma',f'.rsrc=0x{info["image_base"]+info["next_rva"]:x}',str(exe),str(out)])
# Patch resource data directory
b=bytearray(out.read_bytes())
pe=struct.unpack_from('<I',b,0x3c)[0]
opt=pe+4+20
assert struct.unpack_from('<H',b,opt)[0]==0x20b
resdir_off=opt+112+2*8
struct.pack_into('<II',b,resdir_off,info['next_rva'],len(rsrc))
out.write_bytes(b)
print(f'icon resource embedded: RVA=0x{info["next_rva"]:x}, size={len(rsrc)}')
