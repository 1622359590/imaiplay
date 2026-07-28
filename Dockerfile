FROM golang:1.22-alpine AS backend-builder
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/imaiplay ./cmd/server

FROM node:20-alpine AS frontend-builder
WORKDIR /src
COPY web/admin/package*.json web/admin/
RUN cd web/admin && npm ci
COPY web/ web/
RUN cd web/admin && npm run build
RUN cd web/pc && npm ci && npm run build
RUN cd web/h5 && npm ci && npm run build

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
