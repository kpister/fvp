cd ~
wget https://dl.google.com/go/go1.12.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.12.4.linux-amd64.tar.gz
rm go1.12.4.linux-amd64.tar.gz

echo “putting go in path .....”
echo "export PATH=\$PATH:/usr/local/go/bin" >> .bashrc

echo “installing git ...”
sudo yum install -y git

echo “setting GOPATH ...”
mkdir go-workplace
echo "export GOPATH=/home/ec2-user/go-workplace" >> .bashrc

echo “Installing protoc ...”
mkdir protoc
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip
unzip -d protoc protoc-3.7.1-linux-x86_64.zip
rm protoc-3.7.1-linux-x86_64.zip
echo "export PATH=\$PATH:/home/ec2-user/protoc/bin" >> .bashrc
source ~/.bashrc

echo “installing grpc and protoc ...”
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
go get github.com/grpc-ecosystem/go-grpc-middleware

echo "export PATH=$PATH:$GOPATH/bin" >> .bashrc
source ~/.bashrc

echo “FINISHED SETTING UP ...”

echo “pull the repo and checking out experimental branch...”
go get -u github.com/kpister/fvp/server
cd github.com/kpister/fvp/server
go install

cd ~
mkdir cfgs
mkdir logs
rm nohup.out
