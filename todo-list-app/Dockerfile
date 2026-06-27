# syntax=docker/dockerfile:1

FROM node:22-alpine
WORKDIR /app
RUN yarn install --production
CMD ["node", "src/index.js"]
COPY . .
EXPOSE 3000
