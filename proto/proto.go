package proto

//go:generate protoc --go_out=plugins=grpc:../api -I ../proto/ ../proto/auth.proto ../proto/request.proto ../proto/router.proto ../proto/spa.proto
//go:generate go build -i github.com/solidcoredata/scd/api
//go:generate go build github.com/solidcoredata/scd/cmd/...
