module github.com/blablatov/bidistream-mtls-grpc

go 1.18

replace github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto => ./bs-mtls-proto

replace github.com/blablatov/bidistream-mtls-grpc/bs-mcerts => ./bs-mcerts

require (
	github.com/golang/mock v1.1.1
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	golang.org/x/oauth2 v0.5.0
	google.golang.org/grpc v1.52.0-dev
	google.golang.org/protobuf v1.28.1
)

require (
	cloud.google.com/go/compute/metadata v0.2.0 // indirect
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
)
