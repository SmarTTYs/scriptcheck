FROM koalaman/shellcheck-alpine:v0.10.0

COPY scriptcheck /usr/local/bin/scriptcheck

ENTRYPOINT ["scriptcheck check ."]
