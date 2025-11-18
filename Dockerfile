FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o backend-app main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/backend-app ./

ENTRYPOINT [ "./backend-app" ]
