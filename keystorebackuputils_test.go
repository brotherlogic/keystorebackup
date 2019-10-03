package main

import (
	"fmt"
	"testing"

	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"

	pbks "github.com/brotherlogic/keystore/proto"
)

type keystoreTest struct {
	fail     bool
	failRead bool
	results  []*pbks.FileMeta
}

func (k *keystoreTest) getDirectory(ctx context.Context) (*pbks.GetDirectoryResponse, error) {
	if k.fail {
		return nil, fmt.Errorf("Built to fail")
	}

	return &pbks.GetDirectoryResponse{Keys: k.results}, nil
}

func (k *keystoreTest) read(ctx context.Context, req *pbks.ReadRequest) (*pbks.ReadResponse, error) {
	if k.failRead {
		return nil, fmt.Errorf("Built to fail")
	}

	return &pbks.ReadResponse{}, nil
}

func InitTest() *Server {
	s := Init()
	s.saveDirectory = ".testdir/"
	s.SkipLog = true
	s.keystore = &keystoreTest{results: []*pbks.FileMeta{}}
	s.GoServer.KSclient = *keystoreclient.GetTestClient("./testing")
	return s
}

func TestPullKeys(t *testing.T) {
	s := InitTest()
	s.keystore = &keystoreTest{results: []*pbks.FileMeta{&pbks.FileMeta{Key: "key1", Version: 1}}}

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

func TestSyncFailOnRead(t *testing.T) {
	s := InitTest()
	s.keystore = &keystoreTest{fail: true}

	err := s.performSync(context.Background())

	if err == nil {
		t.Fatalf("Bad save did not fail")
	}
}

func TestSaveDataFail(t *testing.T) {
	s := InitTest()
	s.keystore = &keystoreTest{failRead: true}

	err := s.saveData(context.Background(), -1, &pbks.FileMeta{Key: "key1", Version: 1})

	if err == nil {
		t.Fatalf("Bad save did not fail")
	}
}

func TestSaveNewFail(t *testing.T) {
	s := InitTest()
	s.keystore = &keystoreTest{results: []*pbks.FileMeta{&pbks.FileMeta{Key: "key1", Version: 1}}, failRead: true}

	err := s.performSync(context.Background())
	if err == nil {
		t.Errorf("Save did not fail")
	}
}

func TestSaveOldFail(t *testing.T) {
	s := InitTest()
	s.keystore = &keystoreTest{results: []*pbks.FileMeta{&pbks.FileMeta{Key: "key1/sub1", Version: 1}}}

	err := s.performSync(context.Background())
	if err != nil {
		t.Errorf("Error syncing %v", err)
	}

	if s.saves != 1 {
		t.Errorf("Key has not been saved")
	}

	s.keystore = &keystoreTest{results: []*pbks.FileMeta{&pbks.FileMeta{Key: "key1/sub1", Version: 2}}, failRead: true}
	err = s.performSync(context.Background())

	if err == nil {
		t.Errorf("bad save did not fail")
	}
}

func TestTrim(t *testing.T) {
	if trim("/blah") != "blah" {
		t.Errorf("Bad trim %v", trim("/blah"))
	}
}
