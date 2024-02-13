FROM node:20-alpine AS base
WORKDIR /server

FROM base AS build
COPY package*.json ./
RUN npm ci && npm cache clean --force
COPY . .
RUN npm run build

FROM base AS production
COPY --from=build /server/dist ./dist
COPY --from=build /server/node_modules ./node_modules
COPY --from=build /server/package.json ./package.json
EXPOSE 3000
CMD ["npm", "start"]
