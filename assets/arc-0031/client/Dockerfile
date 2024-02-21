FROM node:20-alpine AS base
WORKDIR /arc31

FROM base AS build
COPY package*.json ./
COPY tsconfig.json ./
COPY packages/arc31/package*.json ./packages/arc31/
COPY client/package*.json ./client/
COPY client/tsconfig.json ./client/
RUN npm ci
COPY . .
RUN npm run build --workspace=arc31 && npm run build --workspace=client

FROM base AS production
COPY --from=build /arc31/client/.output ./.output
EXPOSE 3000
ENTRYPOINT ["node", ".output/server/index.mjs"]
