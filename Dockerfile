FROM golang:1.16 as builder
WORKDIR /go/src
ADD . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -v -o /go/bin/main /go/src/cmd/main.go

FROM scratch as app
COPY --from=builder /go/bin/main /

CMD ["/main"]