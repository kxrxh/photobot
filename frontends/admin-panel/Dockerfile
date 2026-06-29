# Stage 1: Build the frontend application
FROM node:20-bookworm-slim AS builder
ARG VITE_HOST
ARG VITE_AUTH_API_URL
WORKDIR /app
# Install Bun globally
RUN npm install -g bun
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . ./
RUN bun run build

# Stage 2: Serve the application with unprivileged Nginx
FROM nginxinc/nginx-unprivileged:stable-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
# Copy custom Nginx configuration
COPY default.conf /etc/nginx/conf.d/default.conf
EXPOSE 4000
CMD ["nginx", "-g", "daemon off;"]