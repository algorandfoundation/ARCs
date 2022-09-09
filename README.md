# ARCs: Algorand Requests for Comments

## Mission
This repository serves to provide a location for the proposal and discussion of ARCs.
You can find the webpage related here: https://arc.algorand.foundation/

## Status
Please consider anything which is not published on https://github.com/algorandfoundation/ARCs as a working paper.
And please consider anything published at https://github.com/algorandfoundation/ARCs with a status of "draft" as an incomplete draft.

## Process
Before submitting a new ARC, please have a look at [ARC-0](ARCs/arc-0000.md).

## Validation

Pull requests in this repository must pass automated validation checks:

* EIP front matter and formatting are [checked](https://github.com/algorandfoundation/ARCs/blob/master/.github/workflows/auto-merge-bot.yml) using [ARC Validator](https://github.com/algorandfoundation/arcw).

To install `arcw` and validate the ARCs repository:

> You will need [Rust/cargo](https://doc.rust-lang.org/cargo/getting-started/installation.html)

```console
git clone git@github.com:algorandfoundation/arcw.git
cargo install --path=arcw arcw
arcw /path/to/ARCs
```


Here is a dedicated part on the [forum](https://forum.algorand.org/c/arc/19) to talk  about ARCs.

To discuss ARCs ideas, see the open [Pull Requests](https://github.com/algorandfoundation/ARCs/pulls) of this repository. 

To discuss ARC drafts, use the corresponding issue in the [issue tracker](https://github.com/algorandfoundation/ARCs/issues).
