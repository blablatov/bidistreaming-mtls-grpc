package main

import (
	"io"
	"log"

	pb "github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto"
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
		orderId, err := stream.Recv() // Читаем ID заказов из входящего потока
		log.Printf("Reading Proc order : %s", orderId)
		if err == io.EOF { // Продолжаем читать, пока не обнаружим конец потока.
			// Client has sent all the messages
			// Send remaining shipments
			log.Printf("EOF : %s", orderId)
			for _, shipment := range combinedShipmentMap {
				// При обнаружении конца потока отправляем клиенту все сгруппированные данные, которые еще остались.
				if err := stream.Send(&shipment); err != nil {
					return err
				}
			}
			return nil //Сервер завершает поток, возвращая nil
		}
		if err != nil {
			log.Println(err)
			return err
		}
		// Логика для объединения заказов в партии на основе адреса доставки
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
			// Передаем клиенту поток заказов, объединенных в партии
			for _, comb := range combinedShipmentMap {
				// Передаем клиенту партию объединенных заказов
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrdersList))
				if err := stream.Send(&comb); err != nil { // Запись объединенных заказов в поток
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
