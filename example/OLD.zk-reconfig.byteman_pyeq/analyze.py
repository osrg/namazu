#!/usr/bin/python

def find_name_files():
    import glob
    return sorted(glob.glob('/tmp/eq_data/*/name'))

def make_csv(name_files):
    csv_str = "#exp_count, pattern_count\n"
    seen_names = []
    exps = patterns = 0
    for name_file in name_files:
        exps += 1
        name_str = ''
        with open(name_file, 'r') as f:
            name_str = f.read()
        if not name_str in seen_names:
            seen_names.append(name_str)
            patterns += 1
        csv_str += "%d, %d\n"%(exps, patterns)
    return csv_str
   
print make_csv(find_name_files())
