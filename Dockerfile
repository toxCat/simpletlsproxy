FROM golang

WORKDIR /go/src/app
COPY . .

RUN go install -v ./...

VOLUME /tls

CMD ["app", "0.0.0.0:80", "0.0.0.0:433"]
