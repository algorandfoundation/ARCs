# arc-0031

## Users registration

Add registered users to the `server/mock/users.json` array, user `nonce` must be set to `''`.

## Run via Docker Compose

From `client` folder run

```
docker build -t arc-0031-client .
```

From `server` folder run

```
docker build -t arc-0031-server .
```

From `arc-0031` folder run

```
docker-compose up
```

Visit `localhost:3000`

## Run local development environment

From `client` folder run

```
yarn install --frozen-lockfile
yarn dev
```

From `server` folder run

```
yarn install --frozen-lockfile
yarn dev
```

Visit `localhost:3000`
