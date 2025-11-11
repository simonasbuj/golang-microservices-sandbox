package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"logger-service/logs"
	"net"

	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, in *logs.LogRequest) (*logs.LogResponse, error) {
	input := in.GetLogEntry()

	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "failed to insert log"}
		return res, fmt.Errorf("failed to insert log: %w", err)
	}

	res := &logs.LogResponse{Result: "logged"}
	return res, nil
}

func (app *App) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen grpc server: %v", err)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	log.Printf("grpc server started on port %s", grpcPort)

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve grpc server: %v", err)
	}
}
