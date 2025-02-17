.PHONY: gen-user
gen-user:
	@protoc --go_out=./rpc_gen --go-grpc_out=./rpc_gen --grpc-gateway_out=./rpc_gen ./idl/cloudstorage/user.proto

.PHONY: gen-file
gen-file:
	@protoc --go_out=./rpc_gen --go-grpc_out=./rpc_gen --grpc-gateway_out=./rpc_gen ./idl/cloudstorage/file.proto

.PHONY: gen-sm
gen-sm:
	@protoc --go_out=./rpc_gen --go-grpc_out=./rpc_gen ./idl/cloudstorage/sm.proto
