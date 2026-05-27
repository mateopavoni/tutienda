# One Dockerfile builds any of the services: pass --build-arg SERVICE=<cmd dir>.
# Static binary on distroless keeps the image tiny and shell-less.
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY . .
RUN go mod tidy
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
