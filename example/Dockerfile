# --- fetch the sshblog binary from its module ---------------------------
FROM golang:1.21-alpine AS build
# Pull the latest released sshblog. Pin a tag for reproducible builds, e.g.
#   RUN go install github.com/cwqt/sshblog@v0.2.0
RUN go install github.com/cwqt/sshblog@latest

# --- runtime: nginx + sshblog under supervisord -------------------------
FROM alpine:3.19
RUN apk add --no-cache nginx supervisor

COPY --from=build /go/bin/sshblog /usr/local/bin/sshblog

# sshblog reads sshblog.yaml, posts/, and .ssh/ from its working directory.
WORKDIR /app
COPY sshblog.yaml ./sshblog.yaml
COPY posts/ ./posts/

# Static HTTP landing page + process configs.
COPY deploy/index.html /var/www/html/index.html
COPY deploy/nginx.conf /etc/nginx/nginx.conf
COPY deploy/supervisord.conf /etc/supervisord.conf

# Generate the RSS feed into nginx's document root at build time, so it is
# served statically at https://<your-domain>/rss.xml. The destination is the
# feed_path set in sshblog.yaml.
RUN sshblog -gen-feed

# 8080 = HTTP (nginx), 2222 = SSH (sshblog)
EXPOSE 8080 2222

CMD ["supervisord", "-c", "/etc/supervisord.conf"]
