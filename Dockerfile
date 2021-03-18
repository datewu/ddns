FROM golang:1.16.2-alpine as builder
RUN apk add ca-certificates git
ARG gitCommit
ARG semVer
COPY ./ /app
WORKDIR /app
RUN CGO_ENABLED=0 \
    GOOS=linux \
        GOARCH=arm \
        GOARM=7 \
    go build -ldflags "-s -w -X main.GitCommit=${gitCommit} \
    -X main.SemVer=${semVer} \
    " -o ./app-binary && \
    mv ./app-binary /app/ && \
    chmod +x /app/app-binary

FROM alpine@sha256:914aa9eb9fec7a7573d755e7fb1d4e95e4285985824cd9367464f454b7609fe1
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /app/app-binary /ddns
