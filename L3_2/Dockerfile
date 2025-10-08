FROM golang:1.24.6

WORKDIR /app

COPY L3_2/go.mod L3_2/go.sum ./
RUN go mod download
COPY L3_2/ .

ARG SERVICE=L3_2
RUN go build -o server ./cmd/server

EXPOSE 8080
CMD ["./server"]