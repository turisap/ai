#FROM golang:1.25-alpine AS builder
#WORKDIR /app
#COPY go.mod go.sum ./
#RUN go mod download
#COPY . .
#RUN CGO_ENABLED=0 go build -o server ./cmd/server
#
#FROM alpine:3.19
#RUN apk add --no-cache ca-certificates
#COPY --from=builder /app/server /server
#ENTRYPOINT ["/server"]

# BEFORE
#mcp          slim        de4138613346   16 seconds ago   15.7MB
# AFTER
#mcp          latest      d4a347c9f278   24 hours ago     24MB


# Stage 1: build — this DOES have the Go toolchain, that's fine
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-task-server ./cmd/server

# Stage 2: runtime — this must NOT have the toolchain. FROM scratch is the key
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/mcp-task-server /mcp-task-server
USER nonroot:nonroot
ENTRYPOINT ["/mcp-task-server"]
