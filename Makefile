.PHONY: protos

protos:
	protoc -I heartbeat_pb/ --go_out=. --go-grpc_out=. heartbeat_pb/heartbeat.proto