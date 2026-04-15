# Release Guidelines

This repository currently publishes binary releases for `arckit`.

## Preconditions

Before creating a release:

- the release commit is already merged to `main`;
- the local checkout is up to date with `origin/main`;
- the `arckit` validation and test checks pass locally;
- the release version is chosen in the form `vX.Y.Z`.

## Step-by-Step Procedure

1. Update your local `main` branch.

   ```sh
   git checkout main
   git pull --ff-only origin main
   ```

2. Run the local validation and build checks.

   ```sh
   pre-commit run --all-files
   pre-commit run codespell --all-files --hook-stage manual
   pre-commit run lychee --all-files --hook-stage manual
   cd arckit
   find . -name '*.go' -print0 | xargs -0 gofmt -w -s
   go vet ./...
   go test ./...
   go build ./cmd/arckit
   go run ./cmd/arckit validate repo ..
   ```

3. Create the release tag from `main`.

   ```sh
   git tag -a arckit/vX.Y.Z -m "arckit vX.Y.Z"
   ```

4. Push the tag to GitHub.

   ```sh
   git push origin arckit/vX.Y.Z
   ```

5. Wait for the `arckit release` GitHub Actions workflow to finish.

   The workflow only accepts tags matching `arckit/v*` and verifies that the tagged
   commit is on `main`.

6. Verify the GitHub release contents.

   Confirm that the release contains:

   - archives for Linux amd64 and arm64;
   - archives for macOS amd64 and arm64;
   - archives for Windows amd64 and arm64;
   - a `SHA256SUMS` file;
   - generated release notes.

7. Edit the GitHub release notes if any manual clarification is needed.

## Notes

- Releases are built directly by `.github/workflows/arckit-release.yml`.
- The workflow publishes archived binaries and checksums; it does not handle signing,
  provenance, Homebrew, or package-manager distribution.
