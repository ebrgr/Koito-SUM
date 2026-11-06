#!/usr/bin/env python3
"""Evaluate source and second-generation quality against a two-stage PNG oracle."""
import argparse
import hashlib
import json
import os
import subprocess
from pathlib import Path


def evaluate(work, report):
    """Reuse timed outputs; reference generation and metrics are outside benchmarks."""
    filename=report/'results-031.json';data=json.loads(filename.read_text());rows=data['rows'];sources={}
    def output(index,row):return work/(f"{index:03}-{row['mode']}-{Path(row['input']).stem}-{row['operation']}-e{row['effort']}.webp")
    for i,row in enumerate(rows):
        if row['operation']=='source' and not row.get('error'):sources[output(i,row).name]=row['input']
    def dims(path):return tuple(map(int,subprocess.check_output(['magick','identify','-ping','-format','%w %h',str(path)],text=True).split()))
    refs={}
    def source_reference(original):
        if original in refs:return refs[original]
        source=work/'corpus'/original;w,h=dims(source);side=min(w,h,1200);dest=work/('oracle-source-'+original+'.png')
        subprocess.run(['magick',str(source),'-auto-orient','-filter','Lanczos','-resize',f'{side}x{side}^','-gravity','center','-background','none','-extent',f'{side}x{side}','-strip','PNG32:'+str(dest)],check=True,capture_output=True)
        refs[original]=dest;return dest
    for i,row in enumerate(rows):
        if row.get('error') or not row.get('output_sha256') or row['operation']in ('parallel','max-pixels','extra-format'):continue
        out=output(i,row)
        if hashlib.sha256(out.read_bytes()).hexdigest()!=row['output_sha256']:raise AssertionError('output changed after measurement')
        op=row['operation']
        if op in ('source','effort-source'):ref=source_reference(row['input'])
        else:
            src=source_reference(sources[row['input']]) if row['input']in sources else work/'corpus'/row['input']
            w,h=dims(src);side=row['size'];ref=work/('oracle-'+out.stem+'.png')
            args=['magick',str(src),'-auto-orient']
            if not(w<side and h<side):args+=['-filter','Lanczos','-resize',f'{side}x{side}^','-gravity','center','-background','none','-extent',f'{side}x{side}']
            subprocess.run(args+['-strip','PNG32:'+str(ref)],check=True,capture_output=True)
        row['quality']=json.loads(subprocess.check_output([str(work/'metrics'),str(ref),str(out)],text=True))
        row['quality']['oracle']='original → lossless geometry source PNG → lossless geometry derivative PNG; no codec recompression in oracle'
        if row['operation']=='derive-from-vips' and row['size']==300 and sources.get(row['input'])in ('natural.jpg','alpha.png'):
            name=Path(sources[row['input']]).stem
            subprocess.run(['magick',str(out),'PNG32:'+str(report/(row['mode']+'-derived-'+name+'.png'))],check=True)
    env=dict(os.environ,CGO_ENABLED='0',GOPROXY='off',GOCACHE='/private/tmp/koito-go-cache',ALPHA_EVIDENCE_DIR=str(work))
    subprocess.run(['go','test','-count=1','./cmd/pure'],cwd=Path(__file__).resolve().parent,env=env,check=True,capture_output=True)
    subprocess.run(['magick',str(work/'alpha-output.webp'),'PNG32:'+str(work/'alpha-independent.png')],check=True)
    data['independent_alpha']=json.loads(subprocess.check_output([str(work/'metrics'),str(work/'alpha-reference.png'),str(work/'alpha-independent.png')],text=True))
    metadata=[]
    for i,row in enumerate(rows):
        if not row.get('output_sha256'):continue
        raw=output(i,row).read_bytes();offset=12;tags=[]
        while offset+8<=len(raw):
            tag=raw[offset:offset+4].decode('ascii',errors='replace');length=int.from_bytes(raw[offset+4:offset+8],'little');tags.append(tag);offset+=8+length+(length%2)
        metadata.append({'row':i,'mode':row['mode'],'input':row['input'],'operation':row['operation'],'chunks':tags,'metadata_present':any(tag in ('EXIF','ICCP','XMP ')for tag in tags)})
    data['metadata']=metadata
    data['timing_scope']='transform only: decode/config within transform, orientation, geometry, encode; excludes ReadFile and outer inspect/admission. RSS includes all worker stages.'
    data['concurrency_scope']='same input buffer shared in process; excludes independent upload/download buffers and HTTP queue'
    filename.write_text(json.dumps(data,indent=2)+'\n')
    print('Quality measured for',sum('quality'in x for x in rows),'outputs')


if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('--work-dir',type=Path,required=True);parser.add_argument('--report-dir',type=Path,required=True);args=parser.parse_args();evaluate(args.work_dir,args.report_dir)
