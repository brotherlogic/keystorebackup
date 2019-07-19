package main

import "golang.org/x/net/context"

func (s *Server) syncKeys(ctx context.Context) error {
	resp, err := s.keystore.getDirectory(ctx)
	if err != nil {
		return err
	}
	s.trackedKeys = resp.Keys
	return nil
}
