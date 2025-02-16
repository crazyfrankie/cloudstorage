.PHONY: gen-user
gen-user:
	@protoc --go_out=./app/user/rpc_gen --go-grpc_out=./app/user/rpc_gen --grpc-gateway_out=./app/user/rpc_gen ./idl/cloudstorage/user.proto

.PHONY: gen-file
gen-file:
	@protoc --go_out=./app/file/rpc_gen --go-grpc_out=./app/file/rpc_gen --grpc-gateway_out=./app/file/rpc_gen ./idl/cloudstorage/file.proto

.PHONY: gen-sm
gen-sm:
	@protoc --go_out=./app/sm/rpc_gen --go-grpc_out=./app/sm/rpc_gen ./idl/cloudstorage/sm.proto
