FROM node:20-alpine AS base
WORKDIR /server

FROM base AS build
COPY package*.json ./
COPY tsconfig.json ./
COPY packages/arc31/package*.json ./packages/arc31/
COPY server/package*.json ./server/
COPY server/tsconfig.json ./server/
RUN npm ci
COPY . .
RUN npm run build --workspace=arc31 && npm run build --workspace=server

FROM base AS production
COPY --from=build /server/node_modules ./node_modules
COPY --from=build /server/packages/arc31/dist ./node_modules/arc31/dist
COPY --from=build /server/packages/arc31/package.json ./node_modules/arc31/package.json
COPY --from=build /server/server/dist ./dist
COPY --from=build /server/server/package.json ./package.json
EXPOSE 3000
CMD ["npm", "start"]
