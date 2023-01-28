FROM golang

WORKDIR /go/src/app
COPY . .

RUN go install -v crypto/tls

VOLUME /tls

CMD ["app", "0.0.0.0:80", "0.0.0.0:433"]
