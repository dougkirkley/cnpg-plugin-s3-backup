FROM golang:1.22 AS base

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download

# Compile the application
COPY . /app
RUN --mount=type=cache,target=/root/.cache/go-build ./scripts/build.sh

FROM alpine:3.20

ARG POSTGRES_VERSION=16

RUN apk add --no-cache postgresql${POSTGRES_VERSION}-client shadow && \
    usermod -u 26 postgres && \
    mkdir -p /backup && \
    chown 26 /backup

# copy backup plugin
COPY --from=base /app/bin/plugin-s3-backup /usr/bin/s3-backup

USER 10001:10001

ENTRYPOINT ["s3-backup"]
CMD ["plugin"]
