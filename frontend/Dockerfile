# ----------------------------
# 1) Build Stage
# ----------------------------
FROM node:20-alpine AS build

WORKDIR /app

# Install dependencies first to leverage caching
COPY package*.json ./
COPY package-lock.json ./

RUN npm install --legacy-peer-deps

# Copy project files
COPY . .

# Build the production files
RUN npm run build


# ----------------------------
# 2) Production Stage - Nginx
# ----------------------------
FROM nginx:1.25-alpine

# Remove default nginx static assets
RUN rm -rf /usr/share/nginx/html/*

# Copy built frontend from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Optional: Copy your custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Replace env.js at container start
CMD sh -c "echo \"window._env_ = { VITE_API_URL: '${VITE_API_URL}' };\" > /usr/share/nginx/html/env.js && nginx -g 'daemon off;'"
