FROM node:16-bullseye

WORKDIR /client
COPY package.json /client/package.json
COPY . .
RUN yarn install \
  --prefer-offline \
  --frozen-lockfile \
  --non-interactive
RUN NODE_ENV=production yarn build

ENV NUXT_APP_API_URL=http://localhost:3001/api
EXPOSE 3000

ENTRYPOINT ["node", ".output/server/index.mjs"]