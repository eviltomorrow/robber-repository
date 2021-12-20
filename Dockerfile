# Build the manager binary
FROM golang:1.17.3-buster AS builder

LABEL maintainer="eviltomorrow@163.com"

ENV WORKSPACE=/app GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY="https://goproxy.io,direct"

WORKDIR $WORKSPACE

ADD . .

# Build
RUN go build -ldflags "-X main.GitSha=${GITSHA} -X main.GitTag=${GITTAG} -X main.GitBranch=${GITBRANCH} -X main.BuildTime=${BUILDTIME} -s -w" -gcflags "all=-trimpath=${GOPATH}" -o bin/robber-repository cmd/robber-repository.go

# Run
FROM alpine:3.15
# Copy binary file
COPY --from=builder /app/bin/robber-repository /bin/
COPY --from=builder /app/config/config.toml /etc/robber-repository/config.toml

VOLUME ["/var/log/robber-repository"]

EXPOSE 27321
ENTRYPOINT ["/bin/robber-repository", "-c", "/etc/robber-repository/config.toml"]