FROM node:20-slim AS base
WORKDIR /client

FROM base AS build
COPY package*.json ./
RUN npm ci && npm cache clean --force
COPY . .
RUN npm run build

FROM base AS production
COPY --from=build /client/.output /client/.output
EXPOSE 3000
ENTRYPOINT ["node", ".output/server/index.mjs"]
