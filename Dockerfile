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
