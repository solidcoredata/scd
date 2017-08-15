package proto

//go:generate protoc --go_out=plugins=grpc:../api -I ../proto/ ../proto/auth.proto ../proto/request.proto ../proto/router.proto
