FROM alpine:3.20.3

COPY scriptcheck /usr/local/bin/scriptcheck

ENTRYPOINT ["scriptcheck check ."]
