# =============================================================================
# AIStudio Frontend - Multi-stage Docker Build
# =============================================================================
FROM node:20-alpine AS builder

WORKDIR /build

COPY Frontend/package.json Frontend/package-lock.json ./
RUN npm ci

COPY Frontend/ .

ARG VITE_APP_ENV=production
ARG VITE_API_BASE_URL=
ARG VITE_WS_URL=
ARG VITE_ENGINE_URL=

ENV VITE_APP_ENV=${VITE_APP_ENV}
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
ENV VITE_WS_URL=${VITE_WS_URL}
ENV VITE_ENGINE_URL=${VITE_ENGINE_URL}

RUN npm run build

# -----------------------------------------------------------------------------
FROM nginx:1.25-alpine

COPY --from=builder /build/dist /usr/share/nginx/html
COPY deploy/nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]