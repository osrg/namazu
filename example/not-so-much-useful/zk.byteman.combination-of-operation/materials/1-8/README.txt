mkdir /disk2/eq_test

# dumb
bin/earthquake init example/zk.byteman.add_node/config_dumb.json example/zk.byteman.add_node/materials /disk2/eq_test
# random
bin/earthquake init example/zk.byteman.add_node/config_random.json example/zk.byteman.add_node/materials /disk2/eq_test

# once
bin/earthquake run /disk2/eq_test
# loop
for i in `seq 1 1000`; do bin/earthquake run /disk2/eq_test; done

bin/earthquake tools summary /disk2/eq_test
bin/earthquake tools visualize --mode gnuplot /disk2/eq_test

bin/earthquake tools dump-trace -trace-path /disk2/eq_test/00000000/result
