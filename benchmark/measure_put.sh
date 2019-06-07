#! /bin/bash

pem=$1
n_exp=5
n_req=3
n_term=10

for ((evil=0;evil<=60;evil+=20)); do
    for ((exp=0;exp<=$n_exp;exp++)); do
        cfg=evil"$evil"_"$exp".cfg
        python ../remote/gen_qs_cfg.py 50 ../remote/ips.txt $cfg -m 10 -e $evil
        python ../remote/setup.py $cfg $pem

        ./benchmark -cfgFile $cfg -nRequest $n_req -nTerm $n_term

        sleep 5

        mkdir evil"$evil"_"$exp"
        python ../remote/get_state.py $pem ../remote/ips.txt --output evil"$evil"_"$exp"
    done
done

