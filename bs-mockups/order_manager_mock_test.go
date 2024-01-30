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

	id := "104"
	items := []string{"Google Home Mini", "Google Nest Hub"}
	destination := "San Jose, CA"
	price := float32(400.00)
	req := &pb.Order{Id: id, Items: items, Destination: destination, Price: price}

	mockClient.EXPECT().ProcessOrders(gomock.Any(), &rpcMsg{msg: req}).
		Return(&pb.CombinedShipment{Id: "Id_mock" + id}, nil)

	testClient_ProcessOrders(t, mockClient)
}

func testClient_ProcessOrders(t *testing.T, client pb.OrderManagementClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	streamProcOrder, err := client.ProcessOrders(ctx)
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}

	// IDs for test. Мапа с тестируемыми ID
	mp := map[string]int{
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
