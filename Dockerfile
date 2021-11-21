FROM alpine:3.15
ARG DOCKER_BIN

LABEL MAINTAINER="onyxhat"
LABEL REPO="https://github.com/onyxhat/github-webhooks"
LABEL FORKED_FROM="https://github.com/hobbsh/github-webhooks"

WORKDIR /app/

COPY ./bin/${DOCKER_BIN} /app/webhook-svc
RUN chmod -R +x /app

EXPOSE 8080
CMD ["/app/webhook-svc"]
