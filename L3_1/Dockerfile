FROM golang:1.24.6

WORKDIR /app

COPY L3_1/go.mod L3_1/go.sum ./
RUN go mod download
COPY L3_1/ .

ARG SERVICE=L3_1
RUN go build -o server ./cmd/server

EXPOSE 8080
CMD ["./server"]