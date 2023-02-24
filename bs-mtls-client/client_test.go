// This conventional test.
// Before his execute run grpc-server  ./tls-service/tls-service

package main

import (
	"context"
	"log"
	"testing"
	"time"

	pb "github.com/blablatov/stream-mtls-grpc/mtls-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// Conventional test that starts a gRPC client test the service with RPC.
// Традиционный тест, который запускает клиент для проверки удаленного метода сервиса.
func TestServer_AddProduct(t *testing.T) {
	tokau := oauth.NewOauthAccess(fetchToken())
	creds, err := credentials.NewClientTLSFromFile(crtFile, hostname)
	if err != nil {
		log.Fatalf("Failed to load credentials: %v", err)
	}
	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(tokau),
		grpc.WithTransportCredentials(creds),
	}
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrderManagementClient(conn)

	// Contact the server and print out its response.
	id := "Sumsung S999"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2029"
	price := float32(777.0)
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

// Тестирование производительности в цикле за указанное колличество итераций
func BenchmarkServer_AddProduct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 25; i++ {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewOrderManagementClient(conn)

		// Contact the server and print out its response.
		id := "Sumsung S999"
		description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2029"
		price := float32(777.0)
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
}
