package mock_bs_mtls_proto

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// rpcMsg implements the gomock.Matcher interface
type rpcMsg struct {
	msg proto.Message
}

func (r *rpcMsg) Matches(msg interface{}) bool {
	m, ok := msg.(proto.Message)
	if !ok {
		return false
	}
	return proto.Equal(m, r.msg)
}

func (r *rpcMsg) String() string {
	return fmt.Sprintf("is %s", r.msg)
}

func TestClient_ProcessOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockOrderManagementClient(ctrl)

	id := "Sumsung S9999"
	description := "Samsung Galaxy S10 is the latest smart phone, launched in February 2039"
	price := float32(7777.0)
	req := &pb.Order{Id: id, Description: description, Price: price}

	mockClient.EXPECT().ProcessOrders(gomock.Any(), &rpcMsg{msg: req}).
		Return(&wrappers.StringValue{Value: "Product_mock" + id}, nil)
	testClient_ProcessOrders(t, mockClient)
}

func testClient_ProcessOrders(t *testing.T, client pb.OrderManagementClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	streamProcOrder, err := client.ProcessOrders(ctx)
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

	channel := make(chan int) // Создаем канал для горутин (goroutines).
	// Вызываем функцию с помощью горутин, чтобы распараллелить чтение сообщений, возвращаемых сервисом.
	//go asncClientBidirectionalRPC(streamProcOrder, channel)
	time.Sleep(time.Millisecond * 1000) // Имитируем задержку при отправке сервису некоторых сообщений

	if err := streamProcOrder.Send(&wrappers.StringValue{Value: "101"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "101", err)
	}
	if err := streamProcOrder.CloseSend(); err != nil { // Сигнализируем о завершении клиентского потока (с ID заказов).
		log.Fatal(err)
	}
	channel <- 1
}

/*func asncClientBidirectionalRPC(streamProcOrder pb.OrderManagementClient, c chan int) {
	for {
		combinedShipment, errProcOrder := streamProcOrder.Recv() // Читаем сообщения сервиса на клиентской стороне.
		if errProcOrder == io.EOF {                              // Условие для обнаружения конца потока.
			break
		}
		log.Println("Combined shipment : ", combinedShipment.OrdersList)
	}
	<-c
}*/
