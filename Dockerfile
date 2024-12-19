FROM golang:1.22.5

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY .env ./

COPY *.go ./

COPY src/ ./src/

RUN CGO_ENABLED=0 GOOS=linux go build -o /godss

EXPOSE 8080

CMD ["/godss"]