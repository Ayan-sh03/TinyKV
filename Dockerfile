# build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git gcc libc-dev
WORKDIR /go/src/app
COPY . .
RUN go mod edit -module app
RUN go get -d -v ./...
RUN go install -v ./...

# final stage
FROM alpine:latest
LABEL Name=appName Version=0.0.1
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
ENTRYPOINT ./app
