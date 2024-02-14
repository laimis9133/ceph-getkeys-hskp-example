FROM golang:1.22-alpine AS build
RUN apk update && apk add ca-certificates
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o /get-keys

FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /get-keys /get-keys
COPY ./template-aws.txt /app/template-aws.txt
COPY ./template-minio.txt /app/template-minio.txt

ENV ceph_bucket=""

ENTRYPOINT [ "/get-keys" ]
