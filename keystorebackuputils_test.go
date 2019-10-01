package main

import (
	"fmt"
	"testing"

	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"

	pbks "github.com/brotherlogic/keystore/proto"
)

type keystoreTest struct {
	fail bool
}

func (k *keystoreTest) getDirectory(ctx context.Context) (*pbks.GetDirectoryResponse, error) {
	if k.fail {
		return nil, fmt.Errorf("Built to fail")
	}

	return &pbks.GetDirectoryResponse{Keys: []*pbks.FileMeta{&pbks.FileMeta{Key: "key1"}, &pbks.FileMeta{Key: "key2"}}}, nil
}

func (k *keystoreTest) read(ctx context.Context, req *pbks.ReadRequest) (*pbks.ReadResponse, error) {
	if k.fail {
		return nil, fmt.Errorf("Built to fail")
	}

	return &pbks.ReadResponse{}, nil
}

func InitTest() *Server {
	s := Init()
	s.saveDirectory = ".testdir"
	s.SkipLog = true
	s.keystore = &keystoreTest{}
	s.GoServer.KSclient = *keystoreclient.GetTestClient("./testing")
	return s
}

func TestPullKeys(t *testing.T) {
	s := InitTest()

	err := s.syncKeys(context.Background())
	if err != nil {
		t.Fatalf("Failed to sync keys")
	}

	if len(s.trackedKeys) == 0 {
		t.Errorf("Failed to pull keys")
	}
}

func TestPullKeysFail(t *testing.T) {
	s := InitTest()
	s.keystore = &keystoreTest{fail: true}

	err := s.syncKeys(context.Background())
	if err == nil {
		t.Fatalf("Sync did not fail")
	}
}

func TestPullData(t *testing.T) {
	s := InitTest()
	s.trackedKeys = append(s.trackedKeys, &pbks.FileMeta{Key: "madeup"})

	err := s.readData(context.Background())
	if err != nil {
		t.Fatalf("Failed to read keys")
	}
}

func TestPullDataFail(t *testing.T) {
	s := InitTest()
	s.trackedKeys = append(s.trackedKeys, &pbks.FileMeta{Key: "madeup"})
	s.keystore = &keystoreTest{fail: true}

	err := s.readData(context.Background())
	if err == nil {
		t.Fatalf("Read did not fail")
	}
}
