FROM golang:1.21-bookworm

ENV SSH_AUTH_SOCK_ENV="/ssh-agent"
ENV SSH_AUTH_SOCK_VOLUME_SRC=/run/host-services/ssh-auth.sock
ENV SSH_AUTH_SOCK_VOLUME_DST=/ssh-agent

RUN set -ex && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    git \
    curl

RUN git config --system url."ssh://git@github.com/".insteadOf "https://github.com/"

WORKDIR /app

COPY ./build/entrypoint.sh /usr/local/bin/
COPY ./build/ssh.conf /etc/ssh/ssh_config

RUN chmod +x /usr/local/bin/entrypoint.sh

RUN mkdir /app/tmp

RUN curl -fLo install.sh https://raw.githubusercontent.com/cosmtrek/air/master/install.sh \
    && chmod +x install.sh && sh install.sh && cp ./bin/air /bin/air
