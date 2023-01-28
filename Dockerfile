FROM golang

WORKDIR /go/src/app
COPY . .

#RUN go get -d -v ./... <--- Deprecated, according to error message. 'go get' is no longer supported outsde a module. 
RUN go install -v ./...

VOLUME /tls

CMD ["app", "0.0.0.0:80", "0.0.0.0:433"]
