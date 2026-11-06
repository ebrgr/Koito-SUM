#!/usr/bin/env python3
"""Generate deterministic fictional images; no user files or database are read."""
import binascii
import hashlib
import json
import math
import struct
import subprocess
import zlib
from pathlib import Path


def chunk(kind, data):
    """Build a CRC-protected PNG chunk."""
    return struct.pack('>I', len(data))+kind+data+struct.pack('>I', binascii.crc32(kind+data)&0xffffffff)


def png(path, width, height, alpha=False, profile=None, natural=False):
    """Stream a repeating textured pattern with asymmetric corner markers."""
    channels=4 if alpha else 3
    compressor=zlib.compressobj(6)
    packed=[]
    rows={}
    for y in range(height):
        key=(y if natural else y%128, y<32, y>=height-32)
        if key not in rows:
            row=bytearray([0])
            for x in range(width):
                r=(x*13+y*7)%256;g=(x*3+y*11)%256;b=(x^y)%256
                if natural:
                    n=math.sin(x/37)+math.sin(y/31)+.5*math.sin((x+y)/17);noise=((x*73856093)^(y*19349663))&15
                    r=int(100+25*n+noise);g=int(130+30*n+noise);b=int(70+15*n+noise)
                if x<32 and y<32:r,g,b=255,0,0
                elif x>=width-32 and y<32:r,g,b=0,255,0
                elif x<32 and y>=height-32:r,g,b=0,0,255
                elif x>=width-32 and y>=height-32:r,g,b=255,255,0
                row.extend((r,g,b))
                if alpha:row.append((x*255)//max(1,width-1))
            rows[key]=bytes(row)
        packed.append(compressor.compress(rows[key]))
    packed.append(compressor.flush())
    ihdr=struct.pack('>IIBBBBB',width,height,8,6 if alpha else 2,0,0,0)
    result=b'\x89PNG\r\n\x1a\n'+chunk(b'IHDR',ihdr)
    if profile:result+=chunk(b'iCCP',b'Koito synthetic RGB\0\0'+zlib.compress(profile))
    result+=chunk(b'tEXt',b'Comment\0Fictional deterministic Koito issue031')+chunk(b'IDAT',b''.join(packed))+chunk(b'IEND',b'')
    path.write_bytes(result)


def icc():
    """Create a small RGB matrix/TRC test profile rather than copying a system profile."""
    def xyz(values):return b'XYZ '+b'\0'*4+b''.join(struct.pack('>i',round(v*65536)) for v in values)
    tags={b'wtpt':xyz((.9642,1,.8249)),b'rXYZ':xyz((.4360747,.2225045,.0139322)),b'gXYZ':xyz((.3850649,.7168786,.0971045)),b'bXYZ':xyz((.1430804,.0606169,.7141733))}
    for name in (b'rTRC',b'gTRC',b'bTRC'):tags[name]=b'curv'+b'\0'*4+struct.pack('>IH',1,563)+b'\0\0'
    header=bytearray(128);header[8:12]=bytes.fromhex('02100000');header[12:24]=b'mntrRGB XYZ ';header[24:36]=struct.pack('>6H',2026,1,1,0,0,0);header[36:40]=b'acsp';header[68:80]=b''.join(struct.pack('>i',round(v*65536)) for v in (.9642,1,.8249));header[80:84]=b'KOTO'
    offset=128+4+12*len(tags);table=[];body=b''
    for name,value in tags.items():table.append(name+struct.pack('>II',offset+len(body),len(value)));body+=value+b'\0'*((-len(value))%4)
    result=header+struct.pack('>I',len(tags))+b''.join(table)+body;result[:4]=struct.pack('>I',len(result));return bytes(result)


def generate(root):
    """Create the full corpus and a manifest with content hashes and provenance."""
    root.mkdir(parents=True,exist_ok=False)
    for name,w,h,a in [('texture',640,480,False),('alpha',400,240,True),('small',80,48,False),('thin',100,500,False),('large',2400,1600,False),('limit40mp',8000,5000,False)]:png(root/(name+'.png'),w,h,a)
    png(root/'icc.png',320,240,profile=icc())
    png(root/'natural.png',640,480,natural=True)
    def magick(*args):subprocess.run(['magick',*map(str,args)],check=True,capture_output=True)
    magick(root/'texture.png','-quality','95',root/'texture.jpg')
    magick(root/'natural.png','-quality','95',root/'natural.jpg')
    magick(root/'texture.png','-colorspace','CMYK',root/'cmyk.jpg')
    for ext in ('bmp','tiff','gif','avif','heic'):magick(root/'small.png',root/('extra.'+ext))
    magick(root/'small.png',root/'alpha.png','-delay','10','-loop','0',root/'animated.gif')
    magick(root/'small.png',root/'alpha.png','-delay','10','-loop','0',root/'animated.webp')
    root.joinpath('extra.svg').write_text('<svg xmlns="http://www.w3.org/2000/svg" width="120" height="80"><rect width="120" height="80" fill="#56ab32"/></svg>')
    jpeg=(root/'texture.jpg').read_bytes()
    for orientation in range(1,9):
        tiff=b'II'+struct.pack('<HIH',42,8,1)+struct.pack('<HHIHHI',274,3,1,orientation,0,0)
        app=b'Exif\0\0'+tiff
        (root/f'exif-{orientation}.jpg').write_bytes(jpeg[:2]+b'\xff\xe1'+struct.pack('>H',len(app)+2)+app+jpeg[2:])
    (root/'truncated.jpg').write_bytes(jpeg[:40]);(root/'invalid.bin').write_bytes(b'not an image')
    for label,w,h in [('axis9999',9999,1),('axis10000',10000,1),('axis10001',10001,1),('pixels39992000',8000,4999),('pixels40008000',8000,5001),('zero',0,1)]:
        if w and h: png(root/(label+'.png'),w,h)
        else: (root/(label+'.png')).write_bytes(b'\x89PNG\r\n\x1a\n'+chunk(b'IHDR',struct.pack('>IIBBBBB',w,h,8,2,0,0,0))+chunk(b'IEND',b''))
    for delta in (-1,0,1):(root/f'bytes{delta:+}.jpg').write_bytes(jpeg+b'\0'*((10<<20)+delta-len(jpeg)))
    manifest={'generator':'corpus.py: deterministic procedural texture, no personal data; axis/pixel probes are complete PNGs; zero.png has only an invalid header','icc':'generated matrix/TRC profile, not a production color calibration','magick':subprocess.check_output(['magick','-version'],text=True).splitlines()[0],'files':[{'file':p.name,'kind':('invalid' if p.name in ('invalid.bin','truncated.jpg','zero.png') else 'synthetic'), 'orientation':(int(p.stem.split('-')[1]) if p.name.startswith('exif-') else None),'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted(root.iterdir())]}
    (root/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
    return manifest
