package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbg "github.com/brotherlogic/goserver/proto"
	pbks "github.com/brotherlogic/keystore/proto"
	pb "github.com/brotherlogic/keystorebackup/proto"
)

const (
	// KEY - where we store sale info
	KEY = "/github.com/brotherlogic/keystorebackup/config"
)

type keystore interface {
	getDirectory(ctx context.Context) (*pbks.GetDirectoryResponse, error)
	read(ctx context.Context, req *pbks.ReadRequest) (*pbks.ReadResponse, error)
}

type keystoreProd struct {
	dial func(server string) (*grpc.ClientConn, error)
}

func (k *keystoreProd) getDirectory(ctx context.Context) (*pbks.GetDirectoryResponse, error) {
	conn, err := k.dial("keystore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbks.NewKeyStoreServiceClient(conn)
	return client.GetDirectory(ctx, &pbks.GetDirectoryRequest{})
}

func (k *keystoreProd) read(ctx context.Context, req *pbks.ReadRequest) (*pbks.ReadResponse, error) {
	conn, err := k.dial("keystore")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbks.NewKeyStoreServiceClient(conn)
	return client.Read(ctx, req)
}

//Server main server type
type Server struct {
	*goserver.GoServer
	config      *pb.Config
	trackedKeys []string
	keystore    keystore
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer:    &goserver.GoServer{},
		config:      &pb.Config{},
		trackedKeys: []string{},
	}
	s.keystore = &keystoreProd{dial: s.DialMaster}
	return s
}

func (s *Server) save(ctx context.Context) {
	s.KSclient.Save(ctx, KEY, s.config)
}

func (s *Server) load(ctx context.Context) error {
	config := &pb.Config{}
	data, _, err := s.KSclient.Read(ctx, KEY, config)

	if err != nil {
		return err
	}

	s.config = data.(*pb.Config)
	return nil
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	// Pass
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.save(ctx)
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	if master {
		err := s.load(ctx)
		return err
	}

	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "last_run", TimeValue: s.config.LastRun},
		&pbg.State{Key: "tracked_keys", Value: int64(len(s.trackedKeys))},
	}
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	var init = flag.Bool("init", false, "Prep server")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.GoServer.KSclient = *keystoreclient.GetClient(server.DialMaster)
	server.PrepServer()
	server.Register = server
	server.RegisterServer("keystorebackup", false)

	if *init {
		ctx, cancel := utils.BuildContext("keystorebackup", "keystorebackup")
		defer cancel()
		server.config.LastRun = time.Now().Unix()
		server.save(ctx)
		return
	}

	server.RegisterRepeatingTask(server.syncKeys, "sync_keys", time.Minute*5)
	server.RegisterRepeatingTask(server.readData, "read_data", time.Minute*4)

	fmt.Printf("%v", server.Serve())
}
