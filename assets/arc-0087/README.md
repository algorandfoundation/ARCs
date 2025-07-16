This reference illustrates the mapping of the specification to a target language, in this case TypeScript

#### Assumptions

- All values are Numbers or Strings ("it should be fine"™ for demonstration or simple use cases)
- A future ARC will handle value treatment and key name packing

## Prerequesistes

- [Node.js](https://nodejs.org/en/download)
- [AlgoKit](https://dev.algorand.co/getting-started/algokit-quick-start/)

## Getting started

> [!WARNING]  
> Algokit may fail with `UND_ERR_SOCKET`, in this case `algokit localnet reset` will resolve the issue

Start localnet

```bash
algokit localnet start
```

Install dependencies

```bash
algokit bootstrap npm
```

Run the build

```bash
algokit project run build
```

Run the demonstration

```bash
algokit project deploy
```

Resetting the demo (make sure to run the build again)

```bash
algokit project run clean
```
