FROM golang:1.25-alpine as builder

ARG SERVICE_NAME

WORKDIR /$SERVICE_NAME

COPY . .

RUN go mod download && go mod tidy

RUN go build -o /tmp/$SERVICE_NAME ./src/cmd/

FROM alpine as deploy

ARG SERVICE_NAME
ENV SERVICE_NAME=$SERVICE_NAME

COPY --from=builder /tmp/$SERVICE_NAME /usr/bin/$SERVICE_NAME

ENTRYPOINT ["/bin/sh", "-c", "/usr/bin/$SERVICE_NAME"]
