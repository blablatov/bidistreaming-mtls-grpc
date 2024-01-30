### Тестирование функциональность клиентского кода без подключения к серверу   
### Testing code without conn to server              

### Go 1.16+  
```shell script
go install github.com/golang/mock/mockgen@v1.6.0  
```  

### Использование Gomock, генерация макетов интерфейсов клиентского gRPC-приложения. Use Gomock        
Для генерации макета интерфейса OrderManagementClient, выполнить.     
Run generation code of mock up for interface OrderManagementClient:   
       
```shell script
mockgen github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto OrderManagementClient > order_manager_mock.go
```  

### Run test    

```shell script
go test order_manager_mock_test.go
```  


