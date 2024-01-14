package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	//"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"google.golang.org/grpc/codes"

	//"google.golang.org/genproto/googleapis/rpc/status"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/oauth2"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"
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
	clientDeadline := time.Now().Add(time.Duration(15000 * time.Millisecond))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	// Process Order : Bi-distreaming scenario
	// Вызываем удаленный метод и получаем ссылку на поток записи и чтения на клиентской стороне
	streamProcOrder, err := client.ProcessOrders(ctx, grpc.UseCompressor(gzip.Name))
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}
	//for i := 0; i < 10; i++ {
	// Sends IDs. Отправляем сообщения с ID сервису.
	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "10"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "10", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "102"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "102", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "104"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "103", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "101"}); err != nil {
		log.Printf("%v.Send(%v) = %v", client, "101", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "105"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "105", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "103"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "103", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "106"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "106", err)
	}

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "-1"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "-1", err)
	}

	channel := make(chan bool, 1) // Создаем канал для горутин (create chanel for goroutines)
	// Вызываем функцию с помощью горутин, распараллеливаем чтение сообщений, возвращаемых сервисом
	go asncClientBidirectionalRPC(streamProcOrder, channel)
	time.Sleep(time.Millisecond * 1000) //  Wait time. Имитируем задержку при отправке сервису сообщений.

	// Сигнализируем о завершении клиентского потока (с ID заказов)
	// Signal about close stream of client
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}

	// // // Cancelling the RPC. Отмена удаленного вызова gRPC на клиентской стороне
	// cancel()
	log.Printf("RPC Status : %v", ctx.Err()) // Status of context. Состояние текущего контекста

	//channel <- true // for test
	<-channel
	//}
}

func asncClientBidirectionalRPC(streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan bool) {
	for {
		// Read messages on side of client
		// Читаем сообщения сервиса на клиентской стороне
		combinedShipment, errProcOrder := streamProcOrder.Recv()

		if errProcOrder != nil {
			errCode := status.Code(errProcOrder)
			if errCode == codes.InvalidArgument {
				log.Printf("Invalid argument error : %s", errCode)
				errStatus := status.Convert(errProcOrder)
				for _, ds := range errStatus.Details() {
					switch info := ds.(type) {
					case *epb.BadRequest_FieldViolation:
						log.Printf("Request field invalid: %s", info)
					default:
						log.Printf("Unexpected error type: %s", info)
					}
				}
			} else {
				log.Printf("Unhandled error: %s", errCode)
			}
		} else {
			log.Println("Combined shipment : ", combinedShipment.Status, combinedShipment.OrdersList)
		}

		// if errProcOrder != nil {
		// 	log.Printf("Error Receiving messages: %v", errProcOrder)
		// 	//break //!!!! TODO - выход не нужен, после отправки ошибки, клиент продолжает запросы
		// 	switch errProcOrder.Error() {
		// 	case "Order ID received is not found - Invalid information":
		// 		<-c
		// 		break
		// 	case "Order ID received is not valid - Invalid information":
		// 		<-c
		// 		break
		// 	default:
		// 		<-c
		// 	}
		// } else {
		// 	if errProcOrder == io.EOF { // End of stream. Обнаружение конца потока.
		// 		break
		// 	}
		// 	log.Println("Combined shipment : ", combinedShipment.Status, combinedShipment.OrdersList)
		// }
	}
	//c <- true // break
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
	s, err := streamer(ctx, desc, cc, method, opts...) // Call func streamer. Вызов функции streamer.
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
// Функция для перехвата отпрвляемых сообщений потокового gRPC
func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("===== [Client Stream Interceptor] Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}
