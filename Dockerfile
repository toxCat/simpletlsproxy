FROM golang

WORKDIR /go/src/github/AnimusPEXUS/simpletlsproxy
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

VOLUME /tls

CMD ["./simpletlsproxy", "0.0.0.0:80", "0.0.0.0:433"]
