### Create & run Docker image. Создание Docker образа.    

Создание Docker контейнера для gRPC-сервера (build container of server):      

```shell script
docker build -t bs-mtls-service .
```

Развернуть задание с серверным gRPC-приложением:         

```shell script
kubectl apply -f grpc-bs-mtls-service.yaml
```
