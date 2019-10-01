package main

import (
	"testing"

	pbks "github.com/brotherlogic/keystore/proto"
	"golang.org/x/net/context"
)

func TestRunThrough(t *testing.T) {
	s := InitTest()
	s.keystore = &keystoreTest{results: []*pbks.FileMeta{&pbks.FileMeta{Key: "key1", Version: 1}}}

	err := s.performSync(context.Background())
	if err != nil {
		t.Errorf("Error syncing %v", err)
	}

	if s.saves != 1 {
		t.Errorf("Key has not been saved")
	}

	err = s.performSync(context.Background())
	if err != nil {
		t.Errorf("Error syncing %v", err)
	}

	if s.saves != 1 {
		t.Errorf("Mulitple saves, despite no change (%v saves)", s.saves)
	}

	s.keystore = &keystoreTest{results: []*pbks.FileMeta{&pbks.FileMeta{Key: "key1", Version: 2}, &pbks.FileMeta{Key: "newkeey", Version: 1}}}

	err = s.performSync(context.Background())
	if err != nil {
		t.Errorf("Error syncing %v", err)
	}

	if s.saves != 3 {
		t.Errorf("Wrong number of saves: %v", s.saves)
	}
}
