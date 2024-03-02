FROM golang:1.20

RUN git clone https://github.com/blablatov/bidistream-mtls-grpc.git
WORKDIR bidistream-mtls-grpc/bs-mtls-service

RUN go mod download

COPY /*.go ./
COPY *.conf ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /bs-service
EXPOSE 50051

#WORKDIR bidistream-mtls-grpc/bs-mtls-client
#RUN CGO_ENABLED=0 GOOS=linux go build -o /bs-client

CMD ["/bs-service"]
