FROM golang:1.26-alpine AS build
ENV GOTOOLCHAIN=auto
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/slotworker ./cmd/slotworker

# Образ для юнит-тестов
FROM build AS unit-test
WORKDIR /src
ENV CGO_ENABLED=0
CMD ["go", "test", "-count=1", "-v", "./tests/unit/..."]

# Образ для интеграционных тестов
FROM build AS integration-test
WORKDIR /src
ENV CGO_ENABLED=0
CMD ["go", "test", "-count=1", "-v", "./tests/integration/..."]

FROM alpine:3.21 AS slotworker
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/slotworker /slotworker
USER nobody
ENTRYPOINT ["/slotworker"]

FROM alpine:3.21 AS app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/app /app
EXPOSE 8080
USER nobody
ENTRYPOINT ["/app"]
