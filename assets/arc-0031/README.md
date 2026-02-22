# ARC-0031 Reference Implementation

**⚠️ Disclamer: This repo provides a simple example of ARC-31. The code is not audited and therefore, it should not be used for production environments.**

The ARC-31 reference implementation is a self contained, standalone repo providing both client and server interfaces. It simulates a User, previously registered to the Validator, authenticating with his Algorand public key *PKa* to a service exposed by the Verifier.

The example shows the entire workflow of authentication using a standalone Algorand account signing an Authentication message. Rekey and MultiSig use cases are not covered yet.

## User registration

Add registered users to the `server/mock/users.db.json` array. The `nonce` field must be left empty `''`.

## Run via Docker Compose

From the root folder run

```
cp .env.example .env
docker compose up --build
```

Visit `localhost`

## Run local development environment

### Requirements
- `node.js`
- `npm`

From both `client`and `server` folders run

```
cp .env.example .env
npm ci
npm run dev
```

Visit `localhost:3001`
