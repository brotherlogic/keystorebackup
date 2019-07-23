package main

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbks "github.com/brotherlogic/keystore/proto"
	pb "github.com/brotherlogic/keystorebackup/proto"
)

func (s *Server) syncKeys(ctx context.Context) error {
	resp, err := s.keystore.getDirectory(ctx)
	if err != nil {
		return err
	}
	s.trackedKeys = resp.Keys
	return nil
}

func (s *Server) readData(ctx context.Context) error {
	allDatums := &pb.AllDatums{Datums: make([]*pb.Datum, 0)}
	for _, key := range s.trackedKeys {
		resp, err := s.keystore.read(ctx, &pbks.ReadRequest{Key: key})
		if err != nil {
			statusCode, ok := status.FromError(err)
			if !ok || statusCode.Code() != codes.OutOfRange {
				return err
			}
		}
		allDatums.Datums = append(allDatums.Datums, &pb.Datum{Key: key, Value: resp.Payload})
	}

	s.Log(fmt.Sprintf("Read in %v worth of data", len(allDatums.String())))
	return nil
}
