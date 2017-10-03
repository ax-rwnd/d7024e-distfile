FROM golang:latest

COPY ./d7024e-distfile /go/

RUN cd /go/src/dfs/ && go get ./...
RUN go build /go/src/dfs/*.go

RUN cd /go/src/server && go get ./...
RUN go build /go/src/server/*.go
