FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o /kartusche

FROM alpine:3.15.4
COPY --from=builder /kartusche /
ENV WORK_DIR=/data
VOLUME [ "/data" ]
ENTRYPOINT [ "/kartusche" ]
CMD [ "server" ]
EXPOSE 3002
EXPOSE 3003
