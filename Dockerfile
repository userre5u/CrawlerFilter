FROM golang:1.20

workdir /app

RUN mkdir Client

COPY go.sum go.mod /app
COPY Client/ Client/

RUN go mod download
RUN go build -o Extractor /app/Client/client/main.go

CMD ["./Extractor"]