FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG NUXT_PUBLIC_API_URL=https://shopapicaovanson.xyz
ARG NUXT_PUBLIC_WS_URL=wss://shopapicaovanson.xyz

ENV NUXT_PUBLIC_API_URL=$NUXT_PUBLIC_API_URL
ENV NUXT_PUBLIC_WS_URL=$NUXT_PUBLIC_WS_URL

RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/.output/public /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
