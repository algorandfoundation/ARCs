FROM node:16-bullseye

WORKDIR /server
COPY package.json /client/package.json
COPY . .
RUN yarn install \
  --prefer-offline \
  --frozen-lockfile \
  --non-interactive
  
ENV PORT 3001
ENV NODE_ENV=production
EXPOSE 3001

CMD [ "yarn", "build" ]




