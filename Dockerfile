FROM golang:1.8

WORKDIR /go/src/hood
COPY Godeps ./Godeps
COPY main.go .
RUN go get github.com/tools/godep
RUN $GOPATH/bin/godep restore

RUN go build -o hood .

CMD ["./hood"]