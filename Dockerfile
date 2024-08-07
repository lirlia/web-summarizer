# ビルドステージ
FROM golang:1.22.5-bookworm AS build

WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/server

FROM chromedp/headless-shell:stable

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /bin/server /server
RUN chmod +x /server

ENTRYPOINT ["/server"]
