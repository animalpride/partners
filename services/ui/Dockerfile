FROM node:22-alpine AS build
WORKDIR /app

# Allow API base to be injected at build time
ARG VITE_API_BASE=/api/auth/v1
ENV VITE_API_BASE=${VITE_API_BASE}
ARG VITE_APP_VERSION=local
ENV VITE_APP_VERSION=${VITE_APP_VERSION}

COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
