FROM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS builder

WORKDIR /app

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pix main.go

FROM scratch

WORKDIR /root/

COPY --from=builder /app/pix .

COPY assets/ ./assets/

EXPOSE 8080

CMD ["./pix"]