rm -rf /home/ec2-user/go-workplace/src/github.com/kpister/raft

echo “pull the repo and checking out master branch...”
go get -u github.com/kpister/fvp/server
cd github.com/kpister/fvp/server
go install

cd ~
mkdir cfgs
mkdir logs
rm nohup.out
