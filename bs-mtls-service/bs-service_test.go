// Testing remote functions without using network
// Модульное тестирование бизнес-логики удаленных методов без передачи по сети.
// С запуском стандартного gRPC-сервера поверх HTTP/2 на реальном порту.
// Имитация запуска сервера с использованием буфера.

package main

import (
	"context"
	"io"
	"log"
	"net"
	"testing"
	"time"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
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
	initSampleData()
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

// Initialization of BufConn. Package bufconn provides a net
// Conn implemented by a buffer and related dialing and listening functionality
// Реализует имитацию запуска сервера на реальном порту с использованием буфера
func initGRPCServerBuffConn() {
	listener = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &mserver{})
	initSampleData()
	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

// Conventional test that starts a gRPC server and client test the service with RPC
func TestServer_ProcessOrders(t *testing.T) {
	log.SetPrefix("Client-test event: ")
	log.SetFlags(log.Lshortfile)
	// Starting a conventional gRPC server runs on HTTP2
	// Запускаем стандартный gRPC-сервер поверх HTTP/2
	initGRPCServerHTTP2()
	conn, err := grpc.Dial(address, grpc.WithInsecure()) // Подключаемся к серверному приложению
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Ее экземпляр содержит все удаленные методы, которые можно вызвать на сервере
	client := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Process Order : Bi-distreaming scenario
	// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
	streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}

	// IDs for test. Мапа с тестируемыми ID
	mp := map[string]int{
		"105": 105,
		"102": 102,
		"103": 103,
		"104": 106,
	}

	for k, v := range mp {
		if k != "" && v != 0 {
			// Иммитируем отправку сообщений сервису.
			if err := streamProcOrder.Send(&wrappers.StringValue{Value: k}); err != nil {
				log.Fatalf("%v.Send(%v) = %v", client, v, err)
			}
		} else {
			log.Printf("ID not found(%s) = %b", k, v)
		}
	}

	channel := make(chan int) // Создаем канал для горутин (create chanel for goroutines)
	// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
	go asncClientBidirectionalRPC(streamProcOrder, channel)
	time.Sleep(time.Millisecond * 500) // Имитируем задержку при отправке сервису сообщений. Wait time

	// Сигнализируем о завершении клиентского потока (с ID заказов)
	// Signal about close stream of client
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}
	channel <- 1
}

// Test written using Buffconn
func TestServer_ProcessOrdersBufConn(t *testing.T) {
	ctx := context.Background()
	initGRPCServerBuffConn()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(getBufDialer(listener)), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Передаем соединение и создаем заглушку
	// Ее экземпляр содержит все удаленные методы, которые можно вызвать на сервере
	client := pb.NewOrderManagementClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Process Order : Bi-distreaming scenario
	// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
	streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}

	// IDs for test. Мапа с тестируемыми ID
	mp := map[string]int{
		"105": 105,
		"102": 102,
		"103": 103,
		"104": 106,
	}

	for k, v := range mp {
		if k != "" && v != 0 {
			// Иммитируем отправку сообщений сервису.
			if err := streamProcOrder.Send(&wrappers.StringValue{Value: k}); err != nil {
				log.Fatalf("%v.Send(%v) = %v", client, v, err)
			}
		} else {
			log.Printf("ID not found(%s) = %b", k, v)
		}
	}

	channel := make(chan int) // Создаем канал для горутин (create chanel for goroutines)
	// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
	go asncClientBidirectionalRPC(streamProcOrder, channel)
	time.Sleep(time.Millisecond * 500) // Имитируем задержку при отправке сервису сообщений. Wait time

	// Сигнализируем о завершении клиентского потока (с ID заказов)
	// Signal about close stream of client
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}
	channel <- 1
}

// Benchmark test
// Тестирование производительности в цикле за указанное колличество итераций
func BenchmarkServer_ProcessOrdersBufConn(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 25; i++ {
		ctx := context.Background()
		initGRPCServerBuffConn()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(getBufDialer(listener)), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		// Ее экземпляр содержит все удаленные методы, которые можно вызвать на сервере
		client := pb.NewOrderManagementClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// Process Order : Bi-distreaming scenario
		// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
		streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
		if err != nil {
			log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
		}

		// IDs for test. Мапа с тестируемыми ID
		mp := map[string]int{
			"105": 105,
			"102": 102,
			"103": 103,
			"104": 106,
		}

		for k, v := range mp {
			if k != "" && v != 0 {
				// Отправляем сообщения сервису.
				if err := streamProcOrder.Send(&wrappers.StringValue{Value: k}); err != nil {
					log.Fatalf("%v.Send(%v) = %v", client, v, err)
				}
			} else {
				log.Printf("ID not found(%s) = %b", k, v)
			}
		}

		channel := make(chan int) // Создаем канал для горутин (create chanel for goroutines)
		// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
		go asncClientBidirectionalRPC(streamProcOrder, channel)
		time.Sleep(time.Millisecond * 1000) // Имитируем задержку при отправке сервису сообщений. Wait time

		// Сигнализируем о завершении клиентского потока (с ID заказов)
		// Signal about close stream of client
		if err := streamProcOrder.CloseSend(); err != nil {
			log.Fatal(err)
		}
		channel <- 1
	}
}

func asncClientBidirectionalRPC(streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan int) {
	for {
		// Читаем сообщения сервиса на клиентской стороне
		// Read messages on side of client
		combinedShipment, errProcOrder := streamProcOrder.Recv()

		if errProcOrder != nil {
			log.Printf("Error Receiving messages: %v", errProcOrder)
			break
		} else {
			if errProcOrder == io.EOF { // Обнаружение конца потока. End of stream
				break
			}
			log.Println("Combined shipment : ", combinedShipment.OrdersList)
		}
	}
	<-c
}
