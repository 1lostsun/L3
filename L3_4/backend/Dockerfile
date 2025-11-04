FROM golang:1.24.6

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN go build -o server ./cmd/server

EXPOSE 8080

CMD ["./server"]
