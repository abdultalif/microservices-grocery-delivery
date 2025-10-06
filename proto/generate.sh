#!/bin/bash

echo "Generating proto files..."

# User Service
protoc --go_out=../user-service/proto --go_opt=paths=source_relative \
       --go-grpc_out=../user-service/proto --go-grpc_opt=paths=source_relative \
       user/user.proto

# Product Service  
protoc --go_out=../product-service/proto --go_opt=paths=source_relative \
       --go-grpc_out=../product-service/proto --go-grpc_opt=paths=source_relative \
       product/product.proto

# Order Service
protoc --go_out=../order-service/proto --go_opt=paths=source_relative \
       --go-grpc_out=../order-service/proto --go-grpc_opt=paths=source_relative \
       user/user.proto product/product.proto order/order.proto

echo "Proto files generated successfully!"