##  Intersystem bidirectional gRPC exchange of encrypted and compressed data     

### Bidirectional Streaming RPC. Двунаправленный потоковый RPC
В двунаправленном потоковом режиме запрос клиента и ответ сервера представлены в виде потоков сообщений.   
Вызов должен быть инициирован на клиентской стороне, клиент устанавливает соединение, отправляя заголовочные фреймы.   
Затем клиентское и серверное приложения обмениваются сообщениями с префиксом длины, не дожидаясь завершения взаимодействия с противоположной стороны. Клиент и сервер отправляют сообщения одновременно.
Используется модель ошибок, встроенная в протокол gRPC и более развитая модель ошибок, реализованная в пакете Google API google.rpc  
Демонстрация bidirectional-grpc обмена локального grpc-клиента с grpc-сервисом запущенном на ВМ Yandex Cloud, развитая модель ошибок, создает сообщение сервиса об недостоверном ID в запросе клиента `Request ID received is not found - Invalid information`.  
<div id="header" align="center">
  <img src="http://gitgif.website.yandexcloud.net/bidirectional-grpc.gif" width="1000"/>
</div>      

### Сборка, запуск и тестирование gRPC-сервиса. Building, running, testing gRPC-service  
Перейти в `bidistream-mtls-grpc/bs-mtls-service` и выполнить  
In order to build, Go to ``Go`` module directory location `bidistream-mtls-grpc/bs-mtls-service` and execute the following
 shell command:
```
go build .
```     
```
./bs-mtls-service
```   

Тестирование бизнес-логики удаленных методов без передачи по сети. Имитация запуска сервера gRPC-сервера поверх HTTP/2 на реальном порту, с использованием буфера.  
Testing remote functions without using network. Using buffer. Bench-test  
```
go test bs-service_test.go
```   
```
go test -bench .
```   


### Сборка, запуск и тестирование gRPC-клиента. Building, running, testing gRPC-client  
Перейти в `bidistream-mtls-grpc/bs-mtls-service` и выполнить.    
In order to build, Go to ``Go`` module directory location `bidistream-mtls-grpc/bs-mtls-client` and execute the following shell command:
```
go build .
```     
```
./bs-mtls-client
```  

Традиционный тест, который запускает клиент для проверки удаленного метода сервиса    
Перед его выполнением запустить grpc-сервер `./bs-mtls-service`. Bench-test     
Conventional test that starts a gRPC client test the service with RPC. Before his execute run grpc-server:   
```
go test bs-client_test.go
```     
```
go test -bench .
```    



### Генерация серверного и клиентского кода из IDL Protocol Buffers. Generate via IDL of Protocol Buffers Server side and Client side code  
Перейти в `bidistream-mtls-grpc/bs-mtls-proto` и выполнить.     
Go to ``Go`` module directory location `bidistream-mtls-grpc/bs-mtls-proto` and execute the following shell commands:    
``` 
protoc product_info.proto --go_out=./ --go-grpc_out=./
protoc product_info.proto --go-grpc_out=require_unimplemented_servers=false:.
``` 
