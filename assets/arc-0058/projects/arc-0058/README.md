# arc-0058

This project has been generated using AlgoKit. See below for default getting started instructions.

# Setup

### Pre-requisites

- [Nodejs 22](https://nodejs.org/en/download) or later
- [AlgoKit CLI 2.5](https://github.com/algorandfoundation/algokit-cli?tab=readme-ov-file#install) or later
- [Docker](https://www.docker.com/) (only required for LocalNet)
- [Puya Compiler 4.4.4](https://pypi.org/project/puyapy/) or later

> For interactive tour over the codebase, download [vsls-contrib.codetour](https://marketplace.visualstudio.com/items?itemName=vsls-contrib.codetour) extension for VS Code, then open the [`.codetour.json`](./.tours/getting-started-with-your-algokit-project.tour) file in code tour extension.

### Initial Setup

#### 1. Clone the Repository
Start by cloning this repository to your local machine.

#### 2. Install Pre-requisites
Ensure the following pre-requisites are installed and properly configured:

- **Docker**: Required for running a local Algorand network.
- **AlgoKit CLI**: Essential for project setup and operations. Verify installation with `algokit --version`, expecting `2.6.0` or later.

#### 3. Bootstrap Your Local Environment
Run the following commands within the project folder:

- **Setup Project**: Execute `algokit project bootstrap all` to install dependencies and setup npm dependencies.
- **Configure environment**: Execute `algokit generate env-file -a target_network localnet` to create a `.env.localnet` file with default configuration for `localnet`.
- **Start LocalNet**: Use `algokit localnet start` to initiate a local Algorand network.

### Development Workflow

#### Terminal
Directly manage and interact with your project using AlgoKit commands:

1. **Build Contracts**: `algokit project run build` compiles all smart contracts. You can also specify a specific contract by passing the name of the contract folder as an extra argument.
For example: `algokit project run build -- hello_world` will only build the `hello_world` contract.
2. **Deploy**: Use `algokit project deploy localnet` to deploy contracts to the local network. You can also specify a specific contract by passing the name of the contract folder as an extra argument.
For example: `algokit project deploy localnet -- hello_world` will only deploy the `hello_world` contract.

#### VS Code 
For a seamless experience with breakpoint debugging and other features:

1. **Open Project**: In VS Code, open the repository root.
2. **Install Extensions**: Follow prompts to install recommended extensions.
3. **Debugging**:
   - Use `F5` to start debugging.

#### JetBrains IDEs
While primarily optimized for VS Code, JetBrains IDEs are supported:

1. **Open Project**: In your JetBrains IDE, open the repository root.
2. **Automatic Setup**: The IDE should configure the Node.js environment.
3. **Debugging**: Use `Shift+F10` or `Ctrl+R` to start debugging. Note: Windows users may encounter issues with pre-launch tasks due to a known bug. See [JetBrains forums](https://youtrack.jetbrains.com/issue/IDEA-277486/Shell-script-configuration-cannot-run-as-before-launch-task) for workarounds.

## AlgoKit Workspaces and Project Management
This project supports both standalone and monorepo setups through AlgoKit workspaces. Leverage [`algokit project run`](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/features/project/run.md) commands for efficient monorepo project orchestration and management across multiple projects within a workspace.

## AlgoKit Generators

This template provides a set of [algokit generators](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/features/generate.md) that allow you to further modify the project instantiated from the template to fit your needs, as well as giving you a base to build your own extensions to invoke via the `algokit generate` command.

### Generate Smart Contract 

By default the template creates a single `HelloWorld` contract under abstracted_account folder in the `smart_contracts` directory. To add a new contract:

1. From the root of the project (`../`) execute `algokit generate smart-contract`. This will create a new starter smart contract and deployment configuration file under `{your_contract_name}` subfolder in the `smart_contracts` directory.
2. Each contract potentially has different creation parameters and deployment steps. Hence, you need to define your deployment logic in `deploy-config.ts` file.
3. Technically, you need to reference your contract deployment logic in the `index.ts` file. However, by default, `index.ts` will auto import all TypeScript deployment files under `smart_contracts` directory. If you want to manually import specific contracts, modify the default code provided by the template in `index.ts` file.

> Please note, above is just a suggested convention tailored for the base configuration and structure of this template. The default code supplied by the template in the `index.ts` file is tailored for the suggested convention. You are free to modify the structure and naming conventions as you see fit.

### Generate '.env' files

By default the template instance does not contain any env files to deploy to different networks. Using [`algokit project deploy`](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/features/project/deploy.md) against `localnet` | `testnet` | `mainnet` will use default values for `algod` and `indexer` unless overwritten via `.env` or `.env.{target_network}`. 

To generate a new `.env` or `.env.{target_network}` file, run `algokit generate env-file`

### Debugging Smart Contracts

This project is optimized to work with AlgoKit AVM Debugger extension. To activate it:

Refer to the commented header in the `index.ts` file in the `smart_contracts` folder.Since you have opted in to include VSCode launch configurations in your project, you can also use the `Debug TEAL via AlgoKit AVM Debugger` launch configuration to interactively select an available trace file and launch the debug session for your smart contract.


For information on using and setting up the `AlgoKit AVM Debugger` VSCode extension refer [here](https://github.com/algorandfoundation/algokit-avm-vscode-debugger). To install the extension from the VSCode Marketplace, use the following link: [AlgoKit AVM Debugger extension](https://marketplace.visualstudio.com/items?itemName=algorandfoundation.algokit-avm-vscode-debugger).### Continuous Integration / Continuous Deployment (CI/CD)

This project uses [GitHub Actions](https://docs.github.com/en/actions/learn-github-actions/understanding-github-actions) to define CI/CD workflows, which are located in the [.github/workflows](`../../.github/workflows`) folder.

> Please note, if you instantiated the project with --workspace flag in `algokit init` it will automatically attempt to move the contents of the `.github` folder to the root of the workspace.

### AlgoKit Workspaces

To define custom `algokit project run` commands refer to [documentation](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/features/project/run.md). This allows orchestration of commands spanning across multiple projects within an algokit workspace based project (monorepo).

#### Setting up GitHub for CI/CD workflow and TestNet deployment

  1. Every time you have a change to your smart contract, and when you first initialize the project you need to [build the contract](#initial-setup) and then commit the `smart_contracts/artifacts` folder so the [output stability](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/articles/output_stability.md) tests pass
  2. Decide what values you want to use for the `allowUpdate` and `allowDelete` parameters specified in [`deploy-config.ts`](./smart_contracts/abstracted_account/deploy-config.ts).
     When deploying to LocalNet these values are both set to `true` for convenience. But for non-LocalNet networks
     they are more conservative and use `false`
     These default values will allow the smart contract to be deployed initially, but will not allow the app to be updated or deleted if is changed and the build will instead fail.
     To help you decide it may be helpful to read the [AlgoKit Utils app deployment documentation](https://github.com/algorandfoundation/algokit-utils-ts/blob/main/docs/capabilities/app-deploy.md) or the [AlgoKit smart contract deployment architecture](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/architecture-decisions/2023-01-12_smart-contract-deployment.md#upgradeable-and-deletable-contracts).
  3. Create a [Github Environment](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#creating-an-environment) named `Test`.
     Note: If you have a private repository and don't have GitHub Enterprise then Environments won't work and you'll need to convert the GitHub Action to use a different approach. Ignore this step if you picked `Starter` preset.
  4. Create or obtain a mnemonic for an Algorand account for use on TestNet to deploy apps, referred to as the `DEPLOYER` account.
  5. Store the mnemonic as a [secret](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment#environment-secrets) `DEPLOYER_MNEMONIC`
     in the Test environment created in step 3.
  6. The account used to deploy the smart contract will require enough funds to create the app, and also fund it. There are two approaches available here:
     * Either, ensure the account is funded outside of CI/CD.
       In Testnet, funds can be obtained by using the [Algorand TestNet dispenser](https://bank.testnet.algorand.network/) and we recommend provisioning 50 ALGOs.
     * Or, fund the account as part of the CI/CD process by using a `DISPENSER_MNEMONIC` GitHub Environment secret to point to a separate `DISPENSER` account that you maintain ALGOs in (similarly, you need to provision ALGOs into this account using the [TestNet dispenser](https://bank.testnet.algorand.network/)).

#### Continuous Integration

For pull requests and pushes to `main` branch against this repository the following checks are automatically performed by GitHub Actions:
 - NPM dependencies are audited using [better-npm-audit](https://github.com/jeemok/better-npm-audit#readme)
 - Code formatting is performed using [Prettier](https://prettier.io/)
 - Linting is checked using [ESLint](https://eslint.org/)
- The base framework for testing is [vitest](https://vitest.dev/), and the project includes two separate kinds of tests:
- - `Algorand TypeScript` smart contract unit tests, that are run using [`algorand-typescript-testing`](https://github.com/algorandfoundation/algorand-typescript-testing), which are executed in a Node.js intepreter emulating major AVM behaviour
- - End-to-end (e2e) `AppClient` tests that are run against `algokit localnet` and test the behaviour in a real network environment
 - Smart contract artifacts are built
 - Smart contract artifacts are checked for [output stability](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/articles/output_stability.md).
 - Smart contract is deployed to a AlgoKit LocalNet instance

> NOTE: By default smart contract artifacts are compiled with `--debug-level` set to 0, to change this, modify the compiler invocation under the `build` script in `package.json`

#### Continuous Deployment

For pushes to `main` branch, after the above checks pass, the following deployment actions are performed:
  - The smart contract(s) are deployed to TestNet using [AlgoNode](https://algonode.io).

> Please note deployment is also performed via `algokit deploy` command which can be invoked both via CI as seen on this project, or locally. For more information on how to use `algokit deploy` please see [AlgoKit documentation](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/features/deploy.md).

# Tools

This project makes use of Algorand TypeScript to build Algorand smart contracts. The following tools are in use:

- [Algorand](https://www.algorand.com/) - Layer 1 Blockchain; [Developer portal](https://dev.algorand.co/), [Why Algorand?](https://dev.algorand.co/getting-started/why-algorand/)
- [AlgoKit](https://github.com/algorandfoundation/algokit-cli) - One-stop shop tool for developers building on the Algorand network; [docs](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/algokit.md), [intro tutorial](https://github.com/algorandfoundation/algokit-cli/blob/main/docs/tutorials/intro.md)
- [Algorand TypeScript](https://github.com/algorandfoundation/puya-ts/) - A semantically and syntactically compatible, typed TypeScript language that works with standard TypeScript tooling and allows you to express smart contracts (apps) and smart signatures (logic signatures) for deployment on the Algorand Virtual Machine (AVM); [docs](https://github.com/algorandfoundation/puya-ts/), [examples](https://github.com/algorandfoundation/puya-ts/tree/main/examples)
- [AlgoKit Utils](https://github.com/algorandfoundation/algokit-utils-ts) - A set of core Algorand utilities that make it easier to build solutions on Algorand.
- [NPM](https://www.npmjs.com/): TypeScript packaging and dependency management.
- [TypeScript](https://www.typescriptlang.org/): Strongly typed programming language that builds on JavaScript
- [ts-node-dev](https://github.com/wclr/ts-node-dev): TypeScript development execution environment- [Prettier](https://prettier.io/): A code formatter.- [ESLint](https://eslint.org/): A JavaScript / TypeScript linter.
- [vitest](https://vitest.dev/): Automated testing.
- [better-npm-audit](https://github.com/jeemok/better-npm-audit#readme): Tool for scanning JavaScript / TypeScript environments for packages with known vulnerabilities.
- [pre-commit](https://pre-commit.com/): A framework for managing and maintaining multi-language pre-commit hooks, to enable pre-commit you need to run `pre-commit install` in the root of the repository. This will install the pre-commit hooks and run them against modified files when committing. If any of the hooks fail, the commit will be aborted. To run the hooks on all files, use `pre-commit run --all-files`.


It has also been configured to have a productive dev experience out of the box in [VS Code](https://code.visualstudio.com/), see the [.vscode](./.vscode) folder.



It has also been configured to have a productive dev experience out of the box in [Jetbrains IDEs](https://www.jetbrains.com/ides/), see the [.idea](./.idea) folder.

