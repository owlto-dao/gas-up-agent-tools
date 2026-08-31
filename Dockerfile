FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/gas-up-mcp ./cmd/mcp

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/gas-up-mcp /usr/local/bin/gas-up-mcp

USER appuser
EXPOSE 4010
ENTRYPOINT ["gas-up-mcp"]
