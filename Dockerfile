# syntax=docker/dockerfile:1
# ---- build backend ----
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server .

# ---- build frontend ----
FROM node:20-alpine AS frontend-builder
WORKDIR /app

COPY frontend/package*.json ./
RUN npm install

COPY frontend ./
RUN npm run build

# ---- final runtime ----
FROM alpine:3.22
WORKDIR /app

RUN apk add --no-cache nginx supervisor

# ---- fix nginx permissions (Alpine nginx defaults to "user nginx;") ----
# 1) Remove the "user nginx;" directive so nginx runs as root inside the container
RUN sed -i 's/^user nginx;/# user nginx;/' /etc/nginx/nginx.conf

# 2) Ensure nginx temp/log directories exist and are writable
RUN mkdir -p /var/lib/nginx/tmp/client_body \
    /var/lib/nginx/tmp/proxy \
    /var/lib/nginx/tmp/fastcgi \
    /var/lib/nginx/tmp/uwsgi \
    /var/lib/nginx/tmp/scgi \
    /var/lib/nginx/logs \
    && chmod -R 755 /var/lib/nginx

# backend
COPY --from=backend-builder /out/server /app/server
COPY backend/.env.example /app/.env.example

# frontend
COPY --from=frontend-builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/http.d/default.conf

# supervisor config
COPY supervisor.conf /etc/supervisor.conf

EXPOSE 7860
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor.conf"]
