#!/bin/bash

rm nohup.out
pkill -f "server --config"

cd /home/ec2-user/go-workplace/src/github.com/kpister/fvp/server/
go install
cd ~

FILES=/home/ec2-user/cfgs/*
for f in $FILES
do
        nohup server --config=$f &
done
