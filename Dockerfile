# build stage
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


# final image
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 9999

CMD ["./main"] 