FROM golang:1.26.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o raft-node ./cmd

FROM debian:bookworm-slim

WORKDIR /app
RUN mkdir scallup_data

COPY --from=builder /app/raft-node .

CMD ["./raft-node"]
