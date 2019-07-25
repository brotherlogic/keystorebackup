package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbks "github.com/brotherlogic/keystore/proto"
	pb "github.com/brotherlogic/keystorebackup/proto"
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

func (s *Server) readData(ctx context.Context) error {
	allDatums := &pb.AllDatums{Datums: make([]*pb.Datum, 0)}
	for _, key := range s.trackedKeys {
		resp, err := s.keystore.read(ctx, &pbks.ReadRequest{Key: key})
		if err != nil {
			statusCode, ok := status.FromError(err)
			if !ok || statusCode.Code() != codes.OutOfRange {
				return err
			}
		} else {
			allDatums.Datums = append(allDatums.Datums, &pb.Datum{Key: key, Value: resp.Payload})
		}
	}

	s.Log(fmt.Sprintf("Read in %v worth of data", len(allDatums.String())))

	data, _ := proto.Marshal(allDatums)
	today := time.Now()
	err := ioutil.WriteFile(s.saveDirectory+"/"+fmt.Sprintf("%v-%v-%v.backup", today.Year(), today.Month(), today.Day()), data, 0644)
	return err
}
