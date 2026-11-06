#!/usr/bin/env python3
"""Run the isolated issue031 corpus and preserve sanitized reproducible evidence."""
import argparse
import hashlib
import json
import os
import platform
import statistics
import subprocess
import tempfile
import time
from pathlib import Path
from corpus import generate

HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[2]


def main():
    """Build independent workers, then measure full paths and same-process concurrency."""
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--work-dir',type=Path,required=True)
    parser.add_argument('--report-dir',type=Path,required=True)
    args=parser.parse_args()
    work=args.work_dir.resolve();work.mkdir(parents=True,exist_ok=False)
    report=args.report_dir.resolve();report.mkdir(parents=True,exist_ok=True)
    env=dict(os.environ,GOPROXY='off',GOCACHE='/private/tmp/koito-go-cache',GOMAXPROCS='2',VIPS_CONCURRENCY='2')
    binaries={}
    for name,cgo in [('pure','0'),('vips','1'),('metrics','0'),('inventory','1')]:
        binary=work/name
        subprocess.run(['go','build','-o',str(binary),'./cmd/'+name],cwd=HERE,env=dict(env,CGO_ENABLED=cgo),check=True)
        binaries[name]=binary
    inventory=json.loads(subprocess.check_output([str(binaries['inventory'])],env=env,text=True))
    manifest=generate(work/'corpus')
    (report/'corpus-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
    rows=[]
    def run(mode,path,op,size=0,workers=1,reps=3,effort=-1,probe=False):
        name=f'{len(rows):03}-{mode}-{path.stem}-{op}-e{effort}'
        output=work/(name+'.webp')
        command=[str(binaries[mode]),'-input',str(path),'-size',str(size),'-workers',str(workers),'-reps',str(reps),'-effort',str(effort),'-output',str(output)]
        if probe:command.append('-probe')
        start=time.monotonic()
        result=None
        try:
            result=subprocess.run(command,env=env,capture_output=True,text=True,timeout=120)
            value=json.loads(result.stdout) if result.stdout.strip() else {'error':'worker-exit-'+str(result.returncode),'stderr':result.stderr[-400:]}
        except subprocess.TimeoutExpired:value={'error':'timeout-120s'}
        value['worker_exit_code']=result.returncode if result is not None else None
        value.update(mode=mode,input=path.name,operation=op,size=size,effort=effort,reps=reps,process_wall_ms=round((time.monotonic()-start)*1000,3))
        times=value.get('times_ms') or []
        if times:
            value['p50_ms']=statistics.median(times);value['p95_ms']=sorted(times)[max(0,__import__('math').ceil(len(times)*.95)-1)]
            value['operations_per_second']=value['completed']/(value['wall_ms']/1000)
        if output.exists():value['output_sha256']=hashlib.sha256(output.read_bytes()).hexdigest()
        rows.append(value)
        return value,output
    corpus=work/'corpus'
    sources={}
    maincases=['natural.jpg','natural.png','texture.jpg','texture.png','alpha.png','small.png','thin.png','large.png','icc.png','cmyk.jpg']+[f'exif-{i}.jpg' for i in range(1,9)]
    for filename in maincases:
        path=corpus/filename
        print('Comparing '+filename,flush=True)
        for mode in ('vips','pure'):
            row,out=run(mode,path,'source')
            if row.get('error'):continue
            sources[(mode,filename)]=out
            # An independent geometry-only PNG oracle; no JPEG/WebP re-encode in the reference.
            reference=work/(out.stem+'-reference.png')
            px=min(int(subprocess.check_output(['magick','identify','-ping','-format','%w %h',str(path)],text=True).split()[0]),int(subprocess.check_output(['magick','identify','-ping','-format','%w %h',str(path)],text=True).split()[1]),1200)
            subprocess.run(['magick',str(path),'-auto-orient','-filter','Lanczos','-resize',f'{px}x{px}^','-gravity','center','-background','none','-extent',f'{px}x{px}','-strip','PNG32:'+str(reference)],check=True,capture_output=True)
            row['quality']=json.loads(subprocess.check_output([str(binaries['metrics']),str(reference),str(out)],env=env,text=True))
            if filename in ('natural.jpg','texture.jpg','alpha.png','exif-6.jpg','icc.png','cmyk.jpg'):
                subprocess.run(['magick',str(out),'-resize','320x320>','PNG32:'+str(report/(mode+'-'+path.stem+'.png'))],check=True)
        for origin in ('vips','pure'):
            old=sources.get((origin,filename))
            if old:
                for mode in ('vips','pure'):run(mode,old,'derive-from-'+origin,300)
    for filename in ('small.png','thin.png'):
        for mode in ('vips','pure'):run(mode,corpus/filename,'direct-small',300)
    for size in (64,128,300,640,1000):
        for mode in ('vips','pure'):run(mode,sources[('vips','texture.jpg')],'all-sizes',size)
    for effort in (2,4):
        for filename in ('natural.jpg','texture.jpg','alpha.png','large.png'):
            run('pure',corpus/filename,'effort-source',effort=effort)
            run('pure',sources[('vips',filename)],'effort-derived',size=300,effort=effort)
    for mode in ('vips','pure'):
        run(mode,corpus/'limit40mp.png','max-pixels',reps=1)
        for count in (1,2,4):run(mode,corpus/'large.png','parallel',workers=count,reps=2)
    for path in sorted(corpus.iterdir()):
        if path.name in maincases or path.name in ('manifest.json','limit40mp.png'):continue
        probe=path.name.startswith(('axis','pixels','zero','bytes'))
        for mode in ('vips','pure'):run(mode,path,'admission-only' if probe else 'extra-format',probe=probe,reps=1)
    # Confirm metadata chunks do not propagate from EXIF/ICC input to generated containers.
    metadata=[]
    for (mode,filename),path in sources.items():
        data=path.read_bytes();tags=[];offset=12
        while offset+8<=len(data):
            tag=data[offset:offset+4].decode('ascii',errors='replace');n=int.from_bytes(data[offset+4:offset+8],'little');tags.append(tag);offset+=8+n+(n%2)
        metadata.append({'mode':mode,'input':filename,'chunks':tags,'metadata_present':any(tag in ('EXIF','ICCP','XMP ') for tag in tags)})
    versions={name:subprocess.check_output(cmd,text=True).strip() for name,cmd in {'go':['go','version'],'vips':['vips','--version'],'magick':['magick','-version']}.items()}
    results={'created':'2026-11-06','platform':platform.platform(),'gomaxprocs':2,'vips_concurrency':2,'versions':versions,'native_capabilities':inventory,'rows':rows,'metadata':metadata,'scope':'synthetic corpus only; independent subprocess per sample, same-process worker concurrency; source and legacy/new-derived paths; no database'}
    (report/'results-031.json').write_text(json.dumps(results,indent=2)+'\n')
    from quality import evaluate
    evaluate(work,report)
    print(json.dumps({'cases':len(rows),'errors':sum(bool(r.get('error')) for r in rows),'report':str(report/'results-031.json')}))


if __name__=='__main__':main()
