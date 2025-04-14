FROM docker.io/library/golang:latest AS builder

WORKDIR /build

ADD go.mod go.sum ./

COPY . .

RUN go build -o downdtc downdtc.go

FROM drpo-docker.rian.ru/base/ubuntu:24.04

WORKDIR /build

COPY --from=builder /build/downdtc /build/downdtc
COPY *.json /build/.

CMD ["./downdtc"]