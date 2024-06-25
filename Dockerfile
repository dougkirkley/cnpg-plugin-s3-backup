FROM alpine:3.20

COPY cnpg-plugin-s3-backup /usr/bin/s3-backup

ARG POSTGRES_VERSION=16

RUN apk add --no-cache postgresql${POSTGRES_VERSION}-client shadow && \
    usermod -u 26 postgres && \
    mkdir -p /backup && \
    chown 26 /backup

USER 10001:10001

ENTRYPOINT ["s3-backup"]
CMD ["plugin"]
