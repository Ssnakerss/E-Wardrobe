FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /wardrobe ./cmd/wardrobe

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /server /usr/local/bin/server
COPY --from=builder /wardrobe /usr/local/bin/wardrobe

RUN mkdir -p /photos && chown -R app:app /photos

USER app

WORKDIR /app

EXPOSE 8080

CMD ["server"]
