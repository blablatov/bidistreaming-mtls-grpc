// This conventional test. Традиционный тест
// Before his execute run grpc-server ./bs-mtls-service/bs-mtls-service
// Перед выполнением запустить grpc-сервер.

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/encoding/gzip"
)

// Conventional test that starts a gRPC client test the service with RPC
// Традиционный тест, который запускает клиент для проверки удаленного метода сервиса
func TestClient_ProcessOrders(t *testing.T) {
	log.SetPrefix("Client-test event: ")
	log.SetFlags(log.Lshortfile)

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

	// Указываем аутентификационные данные для транспортного протокола с помощью DialOption
	opts := []grpc.DialOption{
		// Указываем один и тот же токен OAuth в параметрах всех вызовов в рамках одного соединения
		// Если нужно указывать токен для каждого вызова отдельно, используем CallOption
		grpc.WithPerRPCCredentials(tokau),
		// Указываем транспортные аутентификационные данные в виде параметров соединения
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

	// Finding of Duration. Тестированием определить оптимальное значение для крайнего срока кпд
	clientDeadline := time.Now().Add(time.Duration(2000 * time.Millisecond))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	defer cancel()

	// Process Order : Bi-distreaming scenario
	// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
	streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}

	// IDs for test. Мапа с тестируемыми ID
	mp := map[string]int{
		"10":  10,
		"102": 102,
		"106": 106,
		"104": 104,
		"101": 101,
		"11":  11,
		"103": 103,
	}

	for k, v := range mp {
		if k != "" && v != 0 {
			// Отправляем сообщения сервису
			if err := streamProcOrder.Send(&wrappers.StringValue{Value: k}); err != nil {
				log.Fatalf("%v.Send(%v) = %v", client, k, err)
			}
		} else {
			log.Printf("ID not found(%s) = %b", k, v)
		}
	}

	chs := make(chan struct{}) // Создаем канал для горутин (create chanel for goroutines)

	// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
	go func() {
		asncClientBidirectionalRPC(streamProcOrder, chs)
		chs <- struct{}{}
		close(chs)
	}()

	time.Sleep(time.Millisecond * 500) // Имитируем задержку при отправке сервису сообщений. Wait time

	// Сигнализируем о завершении клиентского потока (с ID заказов)
	// Signal about close stream of client
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}

	chs <- struct{}{}
}

// Тестирование производительности в цикле за указанное колличество итераций
func BenchmarkTestClient_ProcessOrders(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 250; i++ {
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

		// Finding of Duration. Тестированием определить оптимальное значение для крайнего срока кпд
		clientDeadline := time.Now().Add(time.Duration(2000 * time.Millisecond))
		ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

		defer cancel()

		// Process Order : Bi-distreaming scenario
		// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
		streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
		if err != nil {
			log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
		}

		// IDs for test. Мапа с тестируемыми ID
		mp := map[string]int{
			"101": 101,
			"102": 102,
			"106": 106,
			"104": 104,
			"105": 105,
			"11":  11,
			"103": 103,
		}

		for k, v := range mp {
			if k != "" && v != 0 {
				// Отправляем сообщения сервису
				if err := streamProcOrder.Send(&wrappers.StringValue{Value: k}); err != nil {
					log.Fatalf("%v.Send(%v) = %v", client, k, err)
				}
			} else {
				log.Printf("ID not found(%s) = %b", k, v)
			}
		}

		chs := make(chan struct{}) // Создаем канал для горутин (create chanel for goroutines)
		// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
		go asncClientBidirectionalRPC(streamProcOrder, chs)
		time.Sleep(time.Millisecond * 100) // Имитируем задержку при отправке сервису сообщений. Wait time

		// Сигнализируем о завершении клиентского потока (с ID заказов)
		// Signal about close stream of client
		if err := streamProcOrder.CloseSend(); err != nil {
			log.Fatal(err)
		}

		// Cancelling the RPC. Отмена удаленного вызова gRPC на клиентской стороне
		cancel()
		log.Printf("RPC Status : %v", ctx.Err()) // Status of context. Состояние текущего контекста

		chs <- struct{}{}
	}
}

// Provides OAuth2 connection token
// Тест предоставления токена OAuth2
func TestFetchToken(t *testing.T) {
	go func() {
		fetchToken()
		var tok oauth2.Token
		if tok.AccessToken == "blablatok-tokblabla-blablatok" {
			log.Println("Token true")
		}
	}()
}

// Сигнатура для тестирования типа Streamer. Signature of method to test
type strmer interface {
	Streamer(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, ...grpc.CallOption) (grpc.ClientStream, error)
}

type strm struct {
	strmer
	ctx     context.Context
	desc    *grpc.StreamDesc
	cc      *grpc.ClientConn
	method  string
	opts    []grpc.CallOption
	success bool
}

func (d *strm) Streamer(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	if d.success {
		return nil, nil
	}
	return nil, fmt.Errorf("Streamer test error")
}

func TestClientStreamInterceptor(t *testing.T) {

	var su strm
	log.Println("===== [Client Interceptor] ", su.method)
	// Call func streamer. Вызов сигнатуры streamer
	_, err := su.Streamer(su.ctx, su.desc, su.cc, su.method, su.opts...)
	if err != nil {
		log.Println("Error Interceptor", err)
	}
}
