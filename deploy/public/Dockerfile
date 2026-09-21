FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /slangd ./cmd/slangd

FROM alpine:3.23
RUN addgroup -g 65532 slang && adduser -D -u 65532 -G slang slang
COPY --from=build /slangd /usr/local/bin/slangd
USER 65532:65532
ENV SLANG_DIR=/workspace SLANG_LIB=/library SLANG_UI=/tmp/ui SLANG_LIB_REPO_PATH=/library
ENTRYPOINT ["slangd", "--public", "--skip-checks", "--only-daemon", "--without-ui"]
