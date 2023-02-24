## Сертификаты для localhost. Certificates for localhost.  

### Создание собственных сертификатов.    
Безопасное клиентское gRPC-приложение с токеном OAuth.    
Способ сгенерировать закрытый ключ и самоподписанный сертификат для `localhost`  
выполнить следующую команду из пакета `openssl` (Method make certificates for `localhost`):  
  
```shell script
openssl req -x509 -out server.crt -keyout server.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
```

Создание Docker контейнера для gRPC-сервера (build container of server):      

```shell script
./docker build -t tls-service .
```

Развернуть задание с серверным gRPC-приложением:         

```shell script
kubectl apply -f grpc-tls-service.yaml
```
