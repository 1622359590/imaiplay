FROM golang:1.22-alpine AS backend-builder
RUN apk add --no-cache ca-certificates git
WORKDIR /src
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/imaiplay ./cmd/server

FROM node:20-alpine AS frontend-builder
WORKDIR /src
COPY web/package*.json web/
COPY web/shared/package.json web/shared/
COPY web/admin/package.json web/admin/
COPY web/pc/package.json web/pc/
COPY web/h5/package.json web/h5/
RUN cd web && npm ci
COPY web/ web/
RUN cd web && npm run build:all

FROM alpine:3.20
RUN apk add --no-cache ca-certificates nginx wget
WORKDIR /app
COPY --from=backend-builder /out/imaiplay /usr/local/bin/imaiplay
COPY --from=frontend-builder /src/web/admin/dist /usr/share/nginx/html/admin
COPY --from=frontend-builder /src/web/pc/dist /usr/share/nginx/html/pc
COPY --from=frontend-builder /src/web/h5/dist /usr/share/nginx/html/h5
COPY docker/nginx/conf/nginx.conf /etc/nginx/http.d/default.conf
RUN mkdir -p /var/lib/imaiplay/uploads
EXPOSE 80 8080
CMD ["sh", "-c", "nginx -g 'daemon off;' & exec /usr/local/bin/imaiplay"]
