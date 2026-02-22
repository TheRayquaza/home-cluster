FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pix main.go

FROM scratch

WORKDIR /root/

COPY --from=builder /app/pix .

COPY assets/ ./assets/

EXPOSE 8080

CMD ["./pix"]