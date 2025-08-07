FROM golang:latest

WORKDIR /app

COPY ./back/go.mod ./

RUN go mod download

COPY ./back .

RUN go build -o main main.go

CMD ["./main"]