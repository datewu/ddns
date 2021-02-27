FROM golang:1.15.5-alpine as builder
RUN apk add ca-certificates git
ARG gitCommit
ARG semVer
COPY ./ /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.GitCommit=${gitCommit} \
    -X main.SemVer=${semVer} \
    " -o ./app-binary && \
    mv ./app-binary /app/ && \
    chmod +x /app/app-binary

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /app/app-binary /app-binary
ENTRYPOINT [ "/app-binary", "-mode", "prod" ]
