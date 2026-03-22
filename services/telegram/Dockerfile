FROM golang:1.26.0-bookworm
WORKDIR /app
COPY . .
ENV GONOSUMDB="*"
ENV GONOPROXY=""
ENV GOFLAGS="-mod=mod"
RUN go mod tidy
RUN go build -o telegram-service cmd/main.go
CMD ["./telegram-service"]
