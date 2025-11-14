FROM golang:alpine as builder

ENV CGO_ENABLED 1 \
    GOOS linux \
    GOARCH=amd64 \
    GO111MODULE=on

RUN apk add --no-cache build-base pkgconfig librdkafka-dev

# Set the Current Working Directory inside the container
WORKDIR /app


# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -tags dynamic -buildvcs=false -o events-service cmd/main.go

###### Start a new stage from scratch #######
FROM alpine

RUN apk add --no-cache librdkafka
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/events-service .

# Expose the port the app runs on
EXPOSE 8082

# Command to run the application
CMD ["./events-service"]
