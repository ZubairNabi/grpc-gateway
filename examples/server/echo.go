package main

import (
	examples "github.com/ZubairNabi/grpc-gateway/examples/examplepb"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

// Implements of EchoServiceServer

type echoServer struct{}

func newEchoServer() examples.EchoServiceServer {
	return new(echoServer)
}

func (s *echoServer) Echo(ctx context.Context, msg *examples.SimpleMessage) (*examples.SimpleMessage, error) {
	glog.Info(msg)
	return msg, nil
}

func (s *echoServer) EchoBody(ctx context.Context, msg *examples.SimpleMessage) (*examples.SimpleMessage, error) {
	glog.Info(msg)
	return msg, nil
}
