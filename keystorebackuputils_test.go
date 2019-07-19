package main

import (
	"fmt"
	"testing"

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

	return &pbks.GetDirectoryResponse{Keys: []string{"key1", "key2"}}, nil
}

func InitTest() *Server {
	s := Init()
	s.SkipLog = true
	s.keystore = &keystoreTest{}

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
