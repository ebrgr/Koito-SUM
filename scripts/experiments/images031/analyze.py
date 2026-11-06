#!/usr/bin/env python3
"""Classify the recorded experiment without turning missing coverage into approval."""
import argparse
import json
from pathlib import Path


def analyze(directory):
    """Check admissions, pair budgets and finite corpus geometry/quality."""
    data=json.loads((directory/'results-031.json').read_text());rows=data['rows']
    errors=[x for x in rows if x.get('error')]
    expected_shared={'axis10001.png','pixels40008000.png','bytes+1.jpg','zero.png','invalid.bin','truncated.jpg'}
    expected_pure={'animated.gif','animated.webp','extra.bmp','extra.gif','extra.svg','extra.tiff','extra.avif','extra.heic'}
    problems=[]
    for row in rows:
        expected=row['input']in expected_shared or (row['mode']=='pure' and row['input'] in expected_pure) or (row['mode']=='vips' and row['input']=='extra.bmp')
        if bool(row.get('error'))!=expected:problems.append({'mode':row['mode'],'input':row['input'],'operation':row['operation'],'error':row.get('error'),'expected_rejection':expected})
        if row.get('worker_exit_code')!=0:problems.append({'worker_exit':row.get('worker_exit_code'),'input':row['input']})
    pairs={}
    for row in rows:
        if row.get('p95_ms') and row['operation'] in ('source','derive-from-vips','derive-from-pure','direct-small','all-sizes'):
            pairs.setdefault((row['input'],row['operation'],row['size']),{})[row['mode']]=row
    comparisons=[]
    for key,pair in pairs.items():
        if len(pair)!=2:problems.append({'unpaired':key});continue
        a,b=pair['vips'],pair['pure'];comparisons.append({'input':key[0],'operation':key[1],'size':key[2],'dimensions_equal':(a['width'],a['height'])==(b['width'],b['height']),'bytes_ratio':b['bytes']/a['bytes'],'p95_ratio':b['p95_ms']/a['p95_ms']})
    quality=[{'mode':r['mode'],'input':r['input'],**r['quality']} for r in rows if 'quality'in r]
    summary={'samples':len(rows),'expected_rejections':len(errors),'unexpected_results':problems,'paired_operations':len(comparisons),'dimension_mismatches':[x for x in comparisons if not x['dimensions_equal']],'bytes_gate_failures':[x for x in comparisons if x['bytes_ratio']>1.5],'latency_gate_failures':[x for x in comparisons if x['p95_ratio']>3],'worst_bytes_ratio':max(x['bytes_ratio']for x in comparisons),'worst_p95_ratio':max(x['p95_ratio']for x in comparisons),'quality':quality,'metadata_failures':[x for x in data['metadata']if x['metadata_present']],'resource_cases':[x for x in rows if x['operation']in ('max-pixels','parallel')],'comparisons':comparisons}
    absolute=[]
    for i,row in enumerate(rows):
        if 'quality' not in row:continue
        q=row['quality'];threshold=28 if 'alpha' in row['input'] else 30
        if not q.get('dimension_match') or min(q.get('psnr_black_db',0),q.get('psnr_white_db',0))<threshold:
            absolute.append({'row':i,'mode':row['mode'],'input':row['input'],'operation':row['operation'],'threshold_db':threshold,'quality':q})
    relative=[]
    for key,pair in pairs.items():
        if len(pair)!=2 or not all('quality'in v for v in pair.values()):continue
        a,b=pair['vips']['quality'],pair['pure']['quality']
        difference=min(b.get('psnr_black_db',0),b.get('psnr_white_db',0))-min(a.get('psnr_black_db',0),a.get('psnr_white_db',0))
        if difference<-.5:relative.append({'input':key[0],'operation':key[1],'delta_db':difference})
    summary['quality_absolute_alerts']=absolute
    summary['quality_relative_regressions_over_0_5db']=relative
    summary['quality_coverage']=sum('quality'in row for row in rows)
    summary['resource_gate_failures']=[{'mode':r['mode'],'input':r['input'],'operation':r['operation'],'workers':r['workers'],'rss':r['peak_rss_bytes']} for r in rows if (r['operation']=='max-pixels' and r['peak_rss_bytes']>768*1048576) or (r['operation']=='parallel' and r['workers']==2 and r['peak_rss_bytes']>512*1048576)]
    calibrated=[]
    for key,pair in pairs.items():
        if len(pair)!=2 or not all('quality'in v for v in pair.values()):continue
        a,b=pair['vips']['quality'],pair['pure']['quality'];natural='natural' in key[0]
        baseline=min(a.get('psnr_black_db',0),a.get('psnr_white_db',0));candidate=min(b.get('psnr_black_db',0),b.get('psnr_white_db',0))
        tolerance=1.5 if natural and key[1]=='source' else .5
        absolute_threshold=(40 if key[1]=='source' else 35) if natural else 0
        if candidate<absolute_threshold or candidate<baseline-tolerance:calibrated.append({'input':key[0],'operation':key[1],'baseline_db':baseline,'candidate_db':candidate,'tolerance_db':tolerance,'absolute_threshold_db':absolute_threshold})
    summary['calibrated_quality_gate_failures']=calibrated
    summary['independent_alpha']=data.get('independent_alpha')
    (directory/'summary-031.json').write_text(json.dumps(summary,indent=2)+'\n')
    print(json.dumps({k:v for k,v in summary.items()if k not in ('quality','comparisons','resource_cases','quality_absolute_alerts')},indent=2))
    if problems or summary['dimension_mismatches'] or summary['metadata_failures'] or data.get('independent_alpha',{}).get('alpha_max_error')!=0:raise SystemExit('unexpected correctness result')


if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('directory',type=Path);args=parser.parse_args();analyze(args.directory)
