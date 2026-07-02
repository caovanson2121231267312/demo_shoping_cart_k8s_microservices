FROM node:20-alpine AS base
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci --omit=dev 2>/dev/null || npm install --omit=dev
COPY . .
EXPOSE 8087
ENV NODE_ENV=production
CMD ["sh", "-c", "node src/db/migrate.js && node src/index.js"]
