package main

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	pb "github.com/blablatov/stream-mtls-grpc/mtls-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/test/bufconn"
)

const (
	address = "localhost:50051"
	bufSize = 1024 * 1024
)

var listener *bufconn.Listener

func initGRPCServerHTTP2() {
	lis, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &mserver{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

func getBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

// Initialization of BufConn.
// Package bufconn provides a net. Реализует имитацию запуска сервера на реальном порту
// Conn implemented by a buffer and related dialing and listening functionality.
// Реализован с использованием буфера
func initGRPCServerBuffConn() {
	listener = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &mserver{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

// Conventional test that starts a gRPC server and client test the service with RPC
func TestServer_AddProduct(t *testing.T) {
	// Starting a conventional gRPC server runs on HTTP2
	// Запускаем стандартный gRPC-сервер поверх HTTP/2
	initGRPCServerHTTP2()
	conn, err := grpc.Dial(address, grpc.WithInsecure()) // Подключаемся к серверному приложению
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrderManagementClient(conn)

	// Contact the server and print out its response.
	id := "Sumsung S10"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2019"
	price := float32(700.0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Calls remote method of AddProduct
	// Вызываем удаленный метод AddProduct
	r, err := c.AddOrder(ctx, &pb.Order{Id: id, Description: description, Price: price})
	if err != nil { // Checks response. Проверяем ответ
		log.Fatalf("Could not add product: %v", err)
	}
	log.Printf("Res %s", r.Value)
}

// Test written using Buffconn
func TestServer_AddProductBufConn(t *testing.T) {
	ctx := context.Background()
	initGRPCServerBuffConn()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(getBufDialer(listener)), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrderManagementClient(conn)

	// Contact the server and print out its response.
	id := "Sumsung S10"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2019"
	price := float32(700.0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.AddOrder(ctx, &pb.Order{Id: id, Description: description, Price: price})
	if err != nil {
		log.Fatalf("Could not add product: %v", err)
	}
	log.Printf(r.Value)
}

// Тестирование производительности в цикле за указанное колличество итераций
func BenchmarkServer_AddProductBufConn(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 25; i++ {
		ctx := context.Background()
		initGRPCServerBuffConn()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(getBufDialer(listener)), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewOrderManagementClient(conn)

		// Contact the server and print out its response.
		id := "Sumsung S999"
		description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2029"
		price := float32(99999.0)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.AddOrder(ctx, &pb.Order{Id: id, Description: description, Price: price})
		if err != nil {
			log.Fatalf("Could not add product: %v", err)
		}
		log.Printf(r.Value)
	}
}
