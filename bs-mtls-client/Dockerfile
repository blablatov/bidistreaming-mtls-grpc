# Multi-stage bs-mtls-client build
# Многоэтапная сборка bs-mtls-client

FROM golang AS build

ENV location /go/src/github.com/blablatov/bidistream-mtls-grpc

WORKDIR ${location}/bs-mtls-client

ADD ./main.go ${location}/bs-mtls-client

RUN go mod init github.com/blablatov/bidistream-mtls-grpc/bs-mtls-client

RUN CGO_ENABLED=0 go build -o bs-mtls-client

# Go binaries are self-contained executables. Используя директиву FROM scratch - 
# Go образы  не должны содержать ничего, кроме одного двоичного исполняемого файла.
FROM scratch
COPY --from=build ./bs-mtls-client ./bs-mtls-client

ENTRYPOINT ["./bs-mtls-client"]
EXPOSE 50051