#Build Stage
FROM golang:1.20.1-buster AS builder
WORKDIR /app


ENV CGO_ENABLED=1
RUN apt-get update 
RUN apt-get install --assume-yes --no-install-recommends apt-utils build-essential
# Install OpenSSL
RUN apt-get update && apt-get install -y \
    libssl-dev \
    libcrypto++-dev


RUN find / -name libcrypto.so*

COPY go.mod go.sum ./
RUN go mod download

COPY . /app
#RUN go get -v -d ./...
RUN go build -o main patters/hello_world.go
