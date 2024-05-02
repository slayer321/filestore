FROM golang:1.22.2-bullseye
COPY . .
RUN go build -o server main.go
EXPOSE 8090
CMD ["./server"]
