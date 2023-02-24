## Bi-streaming Service and Client of gRPC via mTLS   

### Bidirectional Streaming RPC. Двунаправленный потоковый RPC
В двунаправленном потоковом режиме запрос клиента и ответ сервера представлены в виде потоков сообщений.   
Вызов должен быть инициирован на клиентской стороне, клиент устанавливает соединение, отправляя заголовочные фреймы.   
Затем клиентское и серверное приложения обмениваются сообщениями с префиксом длины, не дожидаясь завершения взаимодействия с противоположной стороны. Клиент и сервер отправляют сообщения одновременно.    

## Building and Running Service

In order to build, Go to ``Go`` module directory location `bidistream-mtls-grpc/bs-mtls-service` and execute the following
 shell command:
```
go build -v 
./bs-mtls-service
```   

## Building and Running Client   

In order to build, Go to ``Go`` module directory location `bidistream-mtls-grpc/bs-mtls-client` and execute the following shell command:
```
go build -v 
./bs-mtls-client
```  

## Additional Information

### Generate Server and Client side code   
Go to ``Go`` module directory location `bidistream-mtls-grpc/bs-mtls-proto` and execute the following shell commands:    
``` 
protoc product_info.proto --go_out=./ --go-grpc_out=./
protoc product_info.proto --go-grpc_out=require_unimplemented_servers=false:.
``` 
