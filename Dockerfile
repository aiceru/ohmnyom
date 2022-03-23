FROM golang:latest as builder

WORKDIR /app
COPY . ./

RUN go build -v -o server

FROM gcr.io/distroless/base:latest

COPY --from=builder /app/server /app/server
COPY --from=builder /app/assets /app/assets
CMD ["/app/server"]