protoc --proto_path ../../../ -I=./proto --go_out=plugins=grpc:./proto proto/keystorebackup.proto
mv proto/github.com/brotherlogic/keystorebackup/proto/* ./proto
