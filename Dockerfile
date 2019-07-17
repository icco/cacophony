FROM golang:1.12
ENV GO111MODULE=on
EXPOSE 8080
WORKDIR /go/src/github.com/icco/cacophony
COPY . .

RUN go build -o /go/bin/server .

CMD ["/go/bin/server"]
