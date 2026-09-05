FROM golang:1.27-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder

WORKDIR /app

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pix main.go

FROM scratch

WORKDIR /root/

COPY --from=builder /app/pix .

COPY assets/ ./assets/

EXPOSE 8080

CMD ["./pix"]