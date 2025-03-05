#############################
# Stage 1: Build the binary #
#############################
FROM golang:1.23.3 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

# Stage 2: Create the runtime   #

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app .
EXPOSE 8080

CMD ["./myapp"]
