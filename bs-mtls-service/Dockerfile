# Multi-stage bs-mtls-server build
# Многоэтапная сборка bs-mtls-service

FROM golang AS build

ENV location /go/src/github.com/blablatov/bidistream-mtls-grpc

WORKDIR ${location}/bs-mtls-service

ADD ./main.go ${location}/bs-mtls-service
ADD ./mtls-proto  ${location}/bs-mtls-proto

#ADD ./go.mod  ${location}/bs-mtls-service
#ADD ./go.sum  ${location}/bs-mtls-service
RUN go mod init github.com/blablatov/bidistream-mtls-grpc/bs-mtls-service

# Download all the dependencies. Загрузка зависимостей
RUN go get -d ./...
# Install the package. Установка пакетов
RUN go install ./...

# Build server app. Сборка серверного приложения.
RUN CGO_ENABLED=0 go build -o /bin/bs-mtls-service 

# Go binaries are self-contained executables. Используя директиву FROM scratch - 
# Go образы  не должны содержать ничего, кроме одного двоичного исполняемого файла.
FROM scratch
COPY --from=build /bin/bs-mtls-service /bin/bs-mtls-service

ENTRYPOINT ["/bin/bs-mtls-service"]
EXPOSE 50051