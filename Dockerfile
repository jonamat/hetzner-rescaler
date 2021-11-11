FROM golang:1.17.0-bullseye AS builder
WORKDIR /build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .

# Create statically linked server binary
RUN go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ./bin/hetzner-rescaler ./main.go

FROM scratch AS runner
WORKDIR /

COPY --from=builder /build/bin/hetzner-rescaler /bin/hr
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["hr", "start", "-s"]