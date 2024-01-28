package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	crtFile = filepath.Join("..", "bs-mcerts", "client.crt")
	keyFile = filepath.Join("..", "bs-mcerts", "client.key")
	caFile  = filepath.Join("..", "bs-mcerts", "ca.crt")
)

const (
	address  = "localhost:50051"
	hostname = "localhost"
)

func main() {
	log.SetPrefix("Client event: ")
	log.SetFlags(log.Lshortfile)

	// Set up the credentials for the connection
	// Значение токена OAuth2. Используем строку, прописанную в коде
	autok := oauth.NewOauthAccess(fetchToken())

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
		// Register interceptor of stream. Регистрация потокового перехватчика
		grpc.WithStreamInterceptor(clientStreamInterceptor),
		// Указываем один и тот же токен OAuth в параметрах всех вызовов в рамках одного соединения
		// Если нужно указывать токен для каждого вызова отдельно, используем CallOption
		grpc.WithPerRPCCredentials(autok),
		// Transport credentials.
		// Указываем транспортные аутентификационные данные в виде параметров соединения
		// Поле ServerName должно быть равно значению Common Name, указанному в сертификате
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName:   hostname,
			Certificates: []tls.Certificate{certificate},
			RootCAs:      certPool,
		})),
	}

	// Set up a connection to the server
	// Устанавливаем безопасное соединение с сервером, передаем параметры аутентификации
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	// Передаем соединение и создаем заглушку.
	// Ее экземпляр содержит все удаленные методы, которые можно вызвать на сервере
	client := pb.NewOrderManagementClient(conn)

	// Finding of Duration. Тестированием определить оптимальное значение для крайнего срока кпд
	clientDeadline := time.Now().Add(time.Duration(5000 * time.Millisecond))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	// Process Order : Bi-distreaming scenario
	// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
	streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}

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
		if k != "" && k != "0" {
			// Sends IDs. Отправляем сообщения с ID сервису.
			if err := streamProcOrder.Send(&wrappers.StringValue{Value: k}); err != nil {
				log.Fatalf("%v.Send(%v) = %v", client, k, err)
			}
		} else {
			log.Printf("ID not found(%s) = %b", k, v)
		}
	}

	// if err := streamProcOrder.Send(&wrappers.StringValue{Value: "103"}); err != nil {
	// 	log.Fatalf("%v.Send(%v) = %v", client, "103", err)
	// }

	chs := make(chan struct{}) // Создаем канал для горутин (create chanel for goroutines)
	//chs := make(chan int, 1)
	// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
	go func() {
		asncClientBidirectionalRPC(streamProcOrder, chs)
		chs <- struct{}{}
	}()

	time.Sleep(time.Millisecond * 500) //  Wait time. Имитируем задержку при отправке сервису сообщений.

	// Сигнализируем о завершении клиентского потока (с ID заказов)
	// Signal about close stream of client
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}

	// Cancelling the RPC. Отмена удаленного вызова gRPC на клиентской стороне
	cancel()
	log.Printf("RPC Status : %v", ctx.Err()) // Status of context. Состояние текущего контекста

	//chs <- struct{}{}
	<-chs
}

func asncClientBidirectionalRPC(streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan struct{}) {
	for {
		// Read messages on side of client
		// Читаем сообщения сервиса на клиентской стороне
		combinedShipment, errProcOrder := streamProcOrder.Recv()

		if errProcOrder != nil {
			log.Printf("Error Receiving messages: %v", errProcOrder)
			break
		} else {
			if errProcOrder == io.EOF { // End of stream. Обнаружение конца потока.
				break
			}
			log.Println("Combined shipment : ", combinedShipment.Status, combinedShipment.OrdersList)
		}
	}
	//c <- struct{}{} // break
	<-c // loop
}

// Provides OAuth2 connection token
// Учетные данные для соединения. Предоставление токена OAuth2
func fetchToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: "blablatok-tokblabla-blablatok",
	}
}

// Client stream interceptor in gRPC
// Клиентский потоковый перехватчик в gRPC
func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
	method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

	// Preprocessing stage, haves access to RPC request before sent to server
	// Этап предобработки, есть доступ к RPC-запросу перед его отправкой на сервер
	log.Println("===== [Client Interceptor] ", method)

	s, err := streamer(ctx, desc, cc, method, opts...) // Call func streamer. Вызов функции streamer
	if err != nil {
		return nil, err
	}
	// Creating wrapper around Client Stream interface, with intercept and go back to app
	// Создание обертки вокруг интерфейса ClientStream, с перехватом и возвращением приложению
	return newWrappedStream(s), nil
}

// Wrapper for interface of rpc.ClientStream
// Обертка для интерфейса grpc.ClientStream
type wrappedStream struct {
	grpc.ClientStream
}

// Func for intercepting received messages of streaming gRPC
// Функция для перехвата принимаемых сообщений потокового gRPC
func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("===== [Client Stream Interceptor] Receive a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

// Func for intercepting sended messages of streaming gRPC
// Функция для перехвата отправляемых сообщений потокового gRPC
func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("===== [Client Stream Interceptor] Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}
