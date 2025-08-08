FROM golang:alpine3.22

WORKDIR /chatapp

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o app cmd/main.go

CMD [ "./app" ]
