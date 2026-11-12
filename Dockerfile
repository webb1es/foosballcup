# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM node:24.18.0-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-fund --no-audit
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.26.11 AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
ARG TARGETOS=linux
ARG TARGETARCH=arm64
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/foosballcup .
RUN mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/foosballcup /foosballcup
COPY --from=build --chown=65532:65532 /out/data /data
ENV PORT=8080
ENV DATABASE_PATH=/data/foosballcup.db
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/foosballcup"]
