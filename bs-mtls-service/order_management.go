package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc/codes"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

var orderMap = make(map[string]pb.Order)

// mСервер используется для реализации order_management
type mserver struct {
	orderMap map[string]*pb.Order
}

// Bi-directional Streaming RPC
// Двунаправленный потоковый RPC
func (s *mserver) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {

	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		// Checks whether current context is cancelled by the client or Deadline was exceeded
		// Сервер проверяет, отменен ли текущий контекст клиентом или превышен крайний срок
		// if stream.Context().Err() == context.Canceled {
		// 	log.Printf("Context Cacelled for this stream: -> %s", stream.Context().Err())
		// 	log.Printf("Stopped processing any more order of this stream!")
		// 	return stream.Context().Err()
		// }

		switch {
		case stream.Context().Err() == context.Canceled:
			log.Printf("Context Cacelled for this stream: -> %s", stream.Context().Err())
			log.Printf("Stopped processing any more order of this stream!")
			return stream.Context().Err()

		case stream.Context().Err() == context.DeadlineExceeded:
			log.Printf("Deadline was exceeded for this stream: -> %s", stream.Context().Err())
			log.Printf("Stopped processing any more order of this stream!")
			return stream.Context().Err()

		default:
			// Err of ID. Проверка ID
			orderId, err := stream.Recv() // Reads IDs. Читаем ID заказов из входящего потока
			log.Printf("Reading Proc order : %s", orderId)

			for k, _ := range orderMap {
				if k != orderId.Value {
					log.Printf("Order ID is invalid! -> Received Order ID %s", orderId)
				}
			}

			if orderId.String() == `value:"-1"` {
				log.Printf("Order ID is invalid! -> Received Order ID %s", orderId)

				errorStatus := status.New(codes.InvalidArgument, "Invalid information received")
				ds, err := errorStatus.WithDetails(
					&epb.BadRequest_FieldViolation{
						Field:       "ID",
						Description: fmt.Sprintf("Order ID received is not valid %s : %s", orderId, orderId.Value),
					},
				)
				if err == nil {
					return errorStatus.Err()
				}

				return ds.Err()
			}

			// Err EOF
			if err == io.EOF { // Reads IDs to EOF. Продолжаем читать, пока не обнаружим конец потока
				// Client has sent all the messages. Send remaining shipments
				log.Printf("EOF : %s", orderId)
				for _, shipment := range combinedShipmentMap {
					// If EOF sends all data of groups
					// При обнаружении конца потока отправляем клиенту все сгруппированные оставшиеся данные
					if err := stream.Send(&shipment); err != nil {
						return err
					}
				}
				return nil //Closes stream. Сервер завершает поток, возвращая nil
			}
			if err != nil {
				log.Println(err)
				return err
			}
			// Logic makes group of orders. Логика для объединения заказов в партии на основе адреса доставки
			destination := orderMap[orderId.GetValue()].Destination
			shipment, found := combinedShipmentMap[destination]

			if found {
				ord := orderMap[orderId.GetValue()]
				shipment.OrdersList = append(shipment.OrdersList, &ord)
				combinedShipmentMap[destination] = shipment
			} else {
				comShip := pb.CombinedShipment{Id: "cmb - " + (orderMap[orderId.GetValue()].Destination), Status: "Processed!"}
				ord := orderMap[orderId.GetValue()]
				comShip.OrdersList = append(shipment.OrdersList, &ord)
				combinedShipmentMap[destination] = comShip
				log.Print(len(comShip.OrdersList), comShip.GetId())
			}

			if batchMarker == orderBatchSize {
				// Передаем клиенту поток заказов, объединенных в партии, group orderBatchSize
				for _, comb := range combinedShipmentMap {
					// Group of orders. Передаем клиенту партию объединенных заказов
					log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrdersList))
					if err := stream.Send(&comb); err != nil { // Writes group of orders. Запись объединенных заказов в поток
						return err
					}
				}
				batchMarker = 0
				combinedShipmentMap = make(map[string]pb.CombinedShipment)
			} else {
				batchMarker++
			}
		}
	}
}
