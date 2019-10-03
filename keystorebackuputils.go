package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/net/context"

	pbks "github.com/brotherlogic/keystore/proto"
	"github.com/golang/protobuf/proto"
)

func (s *Server) syncKeys(ctx context.Context) error {
	resp, err := s.keystore.getDirectory(ctx)
	if err != nil {
		return err
	}
	s.trackedKeys = resp.Keys
	return nil
}

func (s *Server) performSync(ctx context.Context) error {
	resp, err := s.keystore.getDirectory(ctx)
	if err != nil {
		return fmt.Errorf("Error getting directory: %v", err)
	}

	for _, key := range resp.Keys {
		found := false
		for i, stKey := range s.config.LastKeys {
			if stKey.Key == key.Key {
				found = true
				if stKey.Version != key.Version {
					err := s.saveData(ctx, i, key)
					if err != nil {
						return err
					}
				}
			}
		}

		if !found {
			err := s.saveData(ctx, -1, key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func trim(key string) string {
	if strings.HasPrefix(key, "/") {
		return key[1:]
	}
	return key
}

func (s *Server) saveData(ctx context.Context, index int, key *pbks.FileMeta) error {
	resp, err := s.keystore.read(ctx, &pbks.ReadRequest{Key: key.Key})
	if err != nil {
		return fmt.Errorf("Error on read: %v", err)
	}

	pl := resp.Payload
	data, _ := proto.Marshal(pl)
	err = ioutil.WriteFile(s.saveDirectory+fmt.Sprintf("%v.backup-%v", trim(key.Key), key.Version), data, 0644)
	if err == nil {
		s.saves++
		if index >= 0 {
			s.config.LastKeys[index] = key
		} else {
			s.config.LastKeys = append(s.config.LastKeys, key)
		}
		s.save(ctx)
	}

	if err != nil {
		return fmt.Errorf("Error on write: %v", err)
	}

	return err
}
