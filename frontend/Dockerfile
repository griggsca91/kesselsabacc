# ---- Build stage ----
FROM node:22-alpine AS build

WORKDIR /app

# Cache dependency installs
COPY package.json package-lock.json ./
RUN npm ci

# Copy source
COPY . .

# Build with empty env vars so the app uses relative URLs (proxied by nginx)
ENV VITE_API_URL=""
ENV VITE_WS_URL=""
RUN npm run build

# ---- Runtime stage ----
FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config and built assets
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
