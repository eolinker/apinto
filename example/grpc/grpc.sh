#!/bin/bash


OUT_DIR=demo_service

set -e

mkdir -p $OUT_DIR

protoc --grpc-gateway_opt logtostderr=true \
       --grpc-gateway_opt paths=source_relative \
       --grpc-gateway_opt generate_unbound_methods=true \
       --grpc-gateway_out=$OUT_DIR \
       --go_out=$OUT_DIR --go_opt=paths=source_relative     \
       --go-grpc_out=$OUT_DIR --go-grpc_opt=paths=source_relative *.proto