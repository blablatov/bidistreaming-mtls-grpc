// This conventional test. Традиционный тест
// Before his execute run grpc-server ./ds-mtls-service/ds-mtls-service.
// Перед выполнением запустить сервер.

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"testing"
	"time"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/encoding/gzip"
)

// Conventional test that starts a gRPC client test the service with RPC.
// Традиционный тест, который запускает клиент для проверки удаленного метода сервиса.
func TestClient_ProcessOrders(t *testing.T) {
	tokau := oauth.NewOauthAccess(fetchToken())

	// Load the client certificates from disk
	// Создаем пары ключей X.509 непосредственно из ключа и сертификата сервера
	certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		log.Fatalf("could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	// Генерируем пул сертификатов в нашем локальном удостоверяющем центре
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("could not read ca certificate: %s", err)
	}

	// Append the certificates from the CA
	// Добавляем клиентские сертификаты из локального удостоверяющего центра в сгенерированный пул
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append ca certs")
	}

	// Указываем аутентификационные данные для транспортного протокола с помощью DialOption.
	opts := []grpc.DialOption{
		// Указываем один и тот же токен OAuth в параметрах всех вызовов в рамках одного соединения.
		// Если нужно указывать токен для каждого вызова отдельно, используем CallOption.
		grpc.WithPerRPCCredentials(tokau),
		// Указываем транспортные аутентификационные данные в виде параметров соединения
		// Поле ServerName должно быть равно значению Common Name, указанному в сертификате
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:   hostname, // NOTE: this is required!
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Process Order : Bi-distreaming scenario
	// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
	streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}
	// Отправляем сообщения сервису.
	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "102"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "102", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "103"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "103", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "104"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "104", err)
	}

	channel := make(chan int) // Создаем канал для горутин (create chanel for goroutines)
	// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
	go asncClientBidirectionalRPC(streamProcOrder, channel)
	time.Sleep(time.Millisecond * 1000) // Имитируем задержку при отправке сервису сообщений. Wait time

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "101"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "101", err)
	}

	// Сигнализируем о завершении клиентского потока (с ID заказов)
	// Signal about close stream of client
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}
	channel <- 1
}

// Тестирование производительности в цикле за указанное колличество итераций
func BenchmarkTestClient_ProcessOrders(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 25; i++ {
		tokau := oauth.NewOauthAccess(fetchToken())

		// Load the client certificates from disk
		// Создаем пары ключей X.509 непосредственно из ключа и сертификата сервера
		certificate, err := tls.LoadX509KeyPair(crtFile, keyFile)
		if err != nil {
			log.Fatalf("could not load client key pair: %s", err)
		}

		// Create a certificate pool from the certificate authority
		// Генерируем пул сертификатов в нашем локальном удостоверяющем центре
		certPool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Fatalf("could not read ca certificate: %s", err)
		}

		// Append the certificates from the CA
		// Добавляем клиентские сертификаты из локального удостоверяющего центра в сгенерированный пул
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			log.Fatalf("failed to append ca certs")
		}

		opts := []grpc.DialOption{
			grpc.WithPerRPCCredentials(tokau),
			// Указываем транспортные аутентификационные данные в виде параметров соединения
			// Поле ServerName должно быть равно значению Common Name, указанному в сертификате
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
				ServerName:   hostname, // NOTE: this is required!
				Certificates: []tls.Certificate{certificate},
				RootCAs:      certPool,
			})),
		}

		conn, err := grpc.Dial(address, opts...) // Подключаемся к серверному приложению
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		client := pb.NewOrderManagementClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		// Process Order : Bi-distreaming scenario
		// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
		streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
		if err != nil {
			log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
		}
		// Отправляем сообщения сервису.
		if err := streamProcOrder.Send(&wrappers.StringValue{Value: "102"}); err != nil {
			log.Fatalf("%v.Send(%v) = %v", client, "102", err)
		}

		if err := streamProcOrder.Send(&wrappers.StringValue{Value: "103"}); err != nil {
			log.Fatalf("%v.Send(%v) = %v", client, "103", err)
		}

		if err := streamProcOrder.Send(&wrappers.StringValue{Value: "104"}); err != nil {
			log.Fatalf("%v.Send(%v) = %v", client, "104", err)
		}

		channel := make(chan int) // Создаем канал для горутин (create chanel for goroutines)
		// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
		go asncClientBidirectionalRPC(streamProcOrder, channel)
		time.Sleep(time.Millisecond * 1000) // Имитируем задержку при отправке сервису сообщений. Wait time

		if err := streamProcOrder.Send(&wrappers.StringValue{Value: "101"}); err != nil {
			log.Fatalf("%v.Send(%v) = %v", client, "101", err)
		}

		// Сигнализируем о завершении клиентского потока (с ID заказов)
		// Signal about close stream of client
		if err := streamProcOrder.CloseSend(); err != nil {
			log.Fatal(err)
		}
		channel <- 1
	}
}
