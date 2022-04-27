FROM golang:alpine

RUN apk add build-base
RUN mkdir /app
WORKDIR /app

COPY . /app

RUN go mod tidy
RUN go mod download

RUN go build -o /agenda cmd/main.go

ENTRYPOINT ["/agenda"]
