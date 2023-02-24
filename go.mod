module github.com/blablatov/bidistream-mtls-grpc

go 1.18

replace github.com/blablatov/bidistream-mtls-grpc/bs-mtls-proto => ./bs-mtls-proto

replace github.com/blablatov/bidistream-mtls-grpc/bs-mcerts => ./bs-mcerts

require (
	github.com/blablatov/stream-mtls-grpc v0.0.0-20230221155128-be0e4bb827a0
	github.com/golang/protobuf v1.5.2
	golang.org/x/oauth2 v0.5.0
	google.golang.org/grpc v1.53.0
	google.golang.org/protobuf v1.28.1
)

require (
	cloud.google.com/go/compute v1.15.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
)
