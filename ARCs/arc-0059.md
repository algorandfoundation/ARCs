---
arc: 59
title: ASA Inbox Router
description: An application that can route ASAs to users or hold them to later be claimed
author: Kadir Can Çetin (@kadircancetin), Yigit Guler (@yigitguler), Joe Polny (@joe-p), Kevin Wellenzohn (@k13n), Brian Whippo (@silentrhetoric)
discussions-to: https://github.com/algorandfoundation/ARCs/issues/285
status: Final
type: Standards Track
category: ARC
sub-category: Wallet
created: 2024-03-08
requires: 4
---

## Abstract

The goal of this standard is to establish a standard in the Algorand ecosystem by which ASAs can be sent to an intended receiver even if their account is not opted in to the ASA.

A wallet custodied by an application will be used to custody assets on behalf of a given user, with only that user being able to withdraw assets. A master application will be used to map inbox addresses to user address. This master application can route ASAs to users performing whatever actions are necessary.

If integrated into ecosystem technologies including wallets, explorers, and dApps, this standard can provide enhanced capabilities around ASAs which are otherwise strictly bound at the protocol level to require opting in to be received.

## Motivation

Algorand requires accounts to opt in to receive any ASA, a fact which simultaneously:

1. Grants account holders fine-grained control over their holdings by allowing them to select which assets to allow and preventing receipt of unwanted tokens.
2. Frustrates users and developers when accounting for this requirement especially since other blockchains do not have this requirement.

This ARC lays out a new way to navigate the ASA opt in requirement.

### Contemplated Use Cases

The following use cases help explain how this capability can enhance the possibilities within the Algorand ecosystem.

#### Airdrops

An ASA creator who wants to send their asset to a set of accounts faces the challenge of needing their intended receivers to opt in to the ASA ahead of time, which requires non-trivial communication efforts and precludes the possibility of completing the airdrop as a surprise. This claimable ASA standard creates the ability to send an airdrop out to individual addresses so that the receivers can opt in and claim the asset at their convenience--or not, if they so choose.

#### Reducing New User On-boarding Friction

An application operator who wants to on-board users to their game or business may want to reduce the friction of getting people started by decoupling their application on-boarding process from the process of funding a non-custodial Algorand wallet, if users are wholly new to the Algorand ecosystem. As long as the receiver's address is known, an ASA can be sent to them ahead of them having ALGOs in their wallet to cover the minimum balance requirement and opt in to the asset.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

### Router Contract Interface

```json
{
  "name": "ARC59",
  "desc": "",
  "methods": [
    {
      "name": "createApplication",
      "desc": "Deploy ARC59 contract",
      "args": [],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "arc59_optRouterIn",
      "desc": "Opt the ARC59 router into the ASA. This is required before this app can be used to send the ASA to anyone.",
      "args": [
        {
          "name": "asa",
          "type": "uint64",
          "desc": "The ASA to opt into"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "arc59_getOrCreateInbox",
      "desc": "Gets the existing inbox for the receiver or creates a new one if it does not exist",
      "args": [
        {
          "name": "receiver",
          "type": "address",
          "desc": "The address to get or create the inbox for"
        }
      ],
      "returns": {
        "type": "address",
        "desc": "The inbox address"
      }
    },
    {
      "name": "arc59_getSendAssetInfo",
      "args": [
        {
          "name": "receiver",
          "type": "address",
          "desc": "The address to send the asset to"
        },
        {
          "name": "asset",
          "type": "uint64",
          "desc": "The asset to send"
        }
      ],
      "returns": {
        "type": "(uint64,uint64,bool,bool,uint64)",
        "desc": "Returns the following information for sending an asset:The number of itxns required, the MBR required, whether the router is opted in, and whether the receiver is opted in"
      }
    },
    {
      "name": "arc59_sendAsset",
      "desc": "Send an asset to the receiver",
      "args": [
        {
          "name": "axfer",
          "type": "axfer",
          "desc": "The asset transfer to this app"
        },
        {
          "name": "receiver",
          "type": "address",
          "desc": "The address to send the asset to"
        },
        {
          "name": "additionalReceiverFunds",
          "type": "uint64",
          "desc": "The amount of ALGO to send to the receiver/inbox in addition to the MBR"
        }
      ],
      "returns": {
        "type": "address",
        "desc": "The address that the asset was sent to (either the receiver or their inbox)"
      }
    },
    {
      "name": "arc59_claim",
      "desc": "Claim an ASA from the inbox",
      "args": [
        {
          "name": "asa",
          "type": "uint64",
          "desc": "The ASA to claim"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "arc59_reject",
      "desc": "Reject the ASA by closing it out to the ASA creator. Always sends two inner transactions.All non-MBR ALGO balance in the inbox will be sent to the caller.",
      "args": [
        {
          "name": "asa",
          "type": "uint64",
          "desc": "The ASA to reject"
        }
      ],
      "returns": {
        "type": "void"
      }
    },
    {
      "name": "arc59_getInbox",
      "desc": "Get the inbox address for the given receiver",
      "args": [
        {
          "name": "receiver",
          "type": "address",
          "desc": "The receiver to get the inbox for"
        }
      ],
      "returns": {
        "type": "address",
        "desc": "Zero address if the receiver does not yet have an inbox, otherwise the inbox address"
      }
    },
    {
      "name": "arc59_claimAlgo",
      "desc": "Claim any extra algo from the inbox",
      "args": [],
      "returns": {
        "type": "void"
      }
    }
  ]
}
```

### Sending an Asset

When sending an asset, the sender **SHOULD** call `ARC59_getSendAssetInfo` to determine relevant information about the receiver and the router. This information is included as a tuple described below

| Index | Object Property            | Description                                                                      | Type   |
| ----- | -------------------------- | -------------------------------------------------------------------------------- | ------ |
| 0     | itxns                      | The number of itxns required                                                     | uint64 |
| 1     | mbr                        | The amount of ALGO the sender **MUST** send the the router contract to cover MBR | uint64 |
| 2     | routerOptedIn              | Whether the router is already opted in to the asset                              | bool   |
| 3     | receiverOptedIn            | Whether the receiver is already directly opted in to the asset                   | bool   |
| 4     | receiverAlgoNeededForClaim | The amount of ALGO the receiver would currently need to claim the asset          | uint64 |

This information can then be used to send the asset. An example of using this information to send an asset is shown in [the reference implementation section](#typescript-send-asset-function).

### Claiming an Asset

When claiming an asset, the claimer **MUST** call `arc59_claim` to claim the asset from their inbox. This will transfer the asset to the claimer and any extra ALGO in the inbox will be sent to the claimer.

Prior to sending the `arc59_claim` app call, a call to `arc59_claimAlgo` **SHOULD** be made to claim any extra ALGO in the inbox if the inbox balance is above its minimum balance.

An example of claiming an asset is shown in [the reference implementation section](#typescript-claim-function).

## Rationale

This design was created to offer a standard mechanism by which wallets, explorers, and dapps could enable users to send, receive, and find claimable ASAs without requiring any changes to the core protocol.

This ARC is intended to replace [ARC-12](./arc-0012.md). This ARC is simpler than [ARC-12](./arc-0012.md), with the main feature lost being senders not getting back MBR. Given the significant reduction in complexity it is considered to be worth the tradeoff. No way to get back MBR is also another way to disincentivize spam.

### Rejection

The initial proposal for this ARC included a method for burning that leveraged [ARC-54](./arc-0054.md). After further consideration though it was decided to remove the burn functionality with a reject method. The reject method does not burn the ASA. It simply closes out to the creator. This decision was made to reduce the additional complexity and potential user friction that [ARC-54](./arc-0054.md) opt-ins introduced.

### Router MBR

It should be noted that the MBR for the router contract itself is non-recoverable. This was an intentional decision that results in more predictable costs for assets that may freuqently be sent through the router, such as stablecoins.

## Test Cases

Test cases for the JavaScript client and the [ARC-59](./arc-0059.md) smart contract implementation can be found <a href="https://github.com/algorandfoundation/ARCs/tree/main/assets/arc-0059/__test__/">here</a>

## Reference Implementation

A project with a the full reference implementation, including the smart contract and JavaScript library (used for testing), can be found <a href="https://github.com/algorandfoundation/ARCs/tree/main/assets/arc-0059/">here</a>.

### Router Contract

This contract is written using TEALScript v0.90.3

```ts
/* eslint-disable max-classes-per-file */

// eslint-disable-next-line import/no-unresolved, import/extensions
import { Contract } from "@algorandfoundation/tealscript";

type SendAssetInfo = {
  /**
   * The total number of inner transactions required to send the asset through the router.
   * This should be used to add extra fees to the app call
   */
  itxns: uint64;
  /** The total MBR the router needs to send the asset through the router. */
  mbr: uint64;
  /** Whether the router is already opted in to the asset or not */
  routerOptedIn: boolean;
  /** Whether the receiver is already directly opted in to the asset or not */
  receiverOptedIn: boolean;
  /** The amount of ALGO the receiver would currently need to claim the asset */
  receiverAlgoNeededForClaim: uint64;
};

class ControlledAddress extends Contract {
  @allow.create("DeleteApplication")
  new(): Address {
    sendPayment({
      rekeyTo: this.txn.sender,
    });

    return this.app.address;
  }
}

export class ARC59 extends Contract {
  inboxes = BoxMap<Address, Address>();

  /**
   * Deploy ARC59 contract
   *
   */
  createApplication(): void {}

  /**
   * Opt the ARC59 router into the ASA. This is required before this app can be used to send the ASA to anyone.
   *
   * @param asa The ASA to opt into
   */
  arc59_optRouterIn(asa: AssetID): void {
    sendAssetTransfer({
      assetReceiver: this.app.address,
      assetAmount: 0,
      xferAsset: asa,
    });
  }

  /**
   * Gets the existing inbox for the receiver or creates a new one if it does not exist
   *
   * @param receiver The address to get or create the inbox for
   * @returns The inbox address
   */
  arc59_getOrCreateInbox(receiver: Address): Address {
    if (this.inboxes(receiver).exists) return this.inboxes(receiver).value;

    const inbox = sendMethodCall<typeof ControlledAddress.prototype.new>({
      onCompletion: OnCompletion.DeleteApplication,
      approvalProgram: ControlledAddress.approvalProgram(),
      clearStateProgram: ControlledAddress.clearProgram(),
    });

    this.inboxes(receiver).value = inbox;

    return inbox;
  }

  /**
   *
   * @param receiver The address to send the asset to
   * @param asset The asset to send
   *
   * @returns Returns the following information for sending an asset:
   * The number of itxns required, the MBR required, whether the router is opted in, whether the receiver is opted in,
   * and how much ALGO the receiver would need to claim the asset
   */
  arc59_getSendAssetInfo(receiver: Address, asset: AssetID): SendAssetInfo {
    const routerOptedIn = this.app.address.isOptedInToAsset(asset);
    const receiverOptedIn = receiver.isOptedInToAsset(asset);
    const info: SendAssetInfo = {
      itxns: 1,
      mbr: 0,
      routerOptedIn: routerOptedIn,
      receiverOptedIn: receiverOptedIn,
      receiverAlgoNeededForClaim: 0,
    };

    if (receiverOptedIn) return info;

    const algoNeededToClaim =
      receiver.minBalance + globals.assetOptInMinBalance + globals.minTxnFee;

    // Determine how much ALGO the receiver needs to claim the asset
    if (receiver.balance < algoNeededToClaim) {
      info.receiverAlgoNeededForClaim += algoNeededToClaim - receiver.balance;
    }

    // Add mbr and transaction for opting the router in
    if (!routerOptedIn) {
      info.mbr += globals.assetOptInMinBalance;
      info.itxns += 1;
    }

    if (!this.inboxes(receiver).exists) {
      // Two itxns to create inbox (create + rekey)
      // One itxns to send MBR
      // One itxn to opt in
      info.itxns += 4;

      // Calculate the MBR for the inbox box
      const preMBR = globals.currentApplicationAddress.minBalance;
      this.inboxes(receiver).value = globals.zeroAddress;
      const boxMbrDelta = globals.currentApplicationAddress.minBalance - preMBR;
      this.inboxes(receiver).delete();

      // MBR = MBR for the box + min balance for the inbox + ASA MBR
      info.mbr +=
        boxMbrDelta + globals.minBalance + globals.assetOptInMinBalance;

      return info;
    }

    const inbox = this.inboxes(receiver).value;

    if (!inbox.isOptedInToAsset(asset)) {
      // One itxn to opt in
      info.itxns += 1;

      if (!(inbox.balance >= inbox.minBalance + globals.assetOptInMinBalance)) {
        // One itxn to send MBR
        info.itxns += 1;

        // MBR = ASA MBR
        info.mbr += globals.assetOptInMinBalance;
      }
    }

    return info;
  }

  /**
   * Send an asset to the receiver
   *
   * @param receiver The address to send the asset to
   * @param axfer The asset transfer to this app
   * @param additionalReceiverFunds The amount of ALGO to send to the receiver/inbox in addition to the MBR
   *
   * @returns The address that the asset was sent to (either the receiver or their inbox)
   */
  arc59_sendAsset(
    axfer: AssetTransferTxn,
    receiver: Address,
    additionalReceiverFunds: uint64
  ): Address {
    verifyAssetTransferTxn(axfer, {
      assetReceiver: this.app.address,
    });

    // If the receiver is opted in, send directly to their account
    if (receiver.isOptedInToAsset(axfer.xferAsset)) {
      sendAssetTransfer({
        assetReceiver: receiver,
        assetAmount: axfer.assetAmount,
        xferAsset: axfer.xferAsset,
      });

      if (additionalReceiverFunds !== 0) {
        sendPayment({
          receiver: receiver,
          amount: additionalReceiverFunds,
        });
      }

      return receiver;
    }

    const inboxExisted = this.inboxes(receiver).exists;
    const inbox = this.arc59_getOrCreateInbox(receiver);

    if (additionalReceiverFunds !== 0) {
      sendPayment({
        receiver: inbox,
        amount: additionalReceiverFunds,
      });
    }

    if (!inbox.isOptedInToAsset(axfer.xferAsset)) {
      let inboxMbrDelta = globals.assetOptInMinBalance;
      if (!inboxExisted) inboxMbrDelta += globals.minBalance;

      // Ensure the inbox has enough balance to opt in
      if (inbox.balance < inbox.minBalance + inboxMbrDelta) {
        sendPayment({
          receiver: inbox,
          amount: inboxMbrDelta,
        });
      }

      // Opt the inbox in
      sendAssetTransfer({
        sender: inbox,
        assetReceiver: inbox,
        assetAmount: 0,
        xferAsset: axfer.xferAsset,
      });
    }

    // Transfer the asset to the inbox
    sendAssetTransfer({
      assetReceiver: inbox,
      assetAmount: axfer.assetAmount,
      xferAsset: axfer.xferAsset,
    });

    return inbox;
  }

  /**
   * Claim an ASA from the inbox
   *
   * @param asa The ASA to claim
   */
  arc59_claim(asa: AssetID): void {
    const inbox = this.inboxes(this.txn.sender).value;

    sendAssetTransfer({
      sender: inbox,
      assetReceiver: this.txn.sender,
      assetAmount: inbox.assetBalance(asa),
      xferAsset: asa,
      assetCloseTo: this.txn.sender,
    });

    sendPayment({
      sender: inbox,
      receiver: this.txn.sender,
      amount: inbox.balance - inbox.minBalance,
    });
  }

  /**
   * Reject the ASA by closing it out to the ASA creator. Always sends two inner transactions.
   * All non-MBR ALGO balance in the inbox will be sent to the caller.
   *
   * @param asa The ASA to reject
   */
  arc59_reject(asa: AssetID) {
    const inbox = this.inboxes(this.txn.sender).value;

    sendAssetTransfer({
      sender: inbox,
      assetReceiver: asa.creator,
      assetAmount: inbox.assetBalance(asa),
      xferAsset: asa,
      assetCloseTo: asa.creator,
    });

    sendPayment({
      sender: inbox,
      receiver: this.txn.sender,
      amount: inbox.balance - inbox.minBalance,
    });
  }

  /**
   * Get the inbox address for the given receiver
   *
   * @param receiver The receiver to get the inbox for
   *
   * @returns Zero address if the receiver does not yet have an inbox, otherwise the inbox address
   */
  arc59_getInbox(receiver: Address): Address {
    return this.inboxes(receiver).exists
      ? this.inboxes(receiver).value
      : globals.zeroAddress;
  }

  /** Claim any extra algo from the inbox */
  arc59_claimAlgo() {
    const inbox = this.inboxes(this.txn.sender).value;

    assert(inbox.balance - inbox.minBalance !== 0);

    sendPayment({
      sender: inbox,
      receiver: this.txn.sender,
      amount: inbox.balance - inbox.minBalance,
    });
  }
}
```

### TypeScript Send Asset Function

```ts
/**
 * Send an asset to a receiver using the ARC59 router
 *
 * @param appClient The ARC59 client generated by algokit
 * @param assetId The ID of the asset to send
 * @param sender The address of the sender
 * @param receiver The address of the receiver
 * @param algorand The AlgorandClient instance to use to send transactions
 * @param sendAlgoForNewAccount Whether to send 201_000 uALGO to the receiver so they can claim the asset with a 0-ALGO balance
 */
async function arc59SendAsset(
  appClient: Arc59Client,
  assetId: bigint,
  sender: string,
  receiver: string,
  algorand: algokit.AlgorandClient
) {
  // Get the address of the ARC59 router
  const arc59RouterAddress = (await appClient.appClient.getAppReference())
    .appAddress;

  // Call arc59GetSendAssetInfo to get the following:
  // itxns - The number of transactions needed to send the asset
  // mbr - The minimum balance that must be sent to the router
  // routerOptedIn - Whether the router has opted in to the asset
  // receiverOptedIn - Whether the receiver has opted in to the asset
  const [
    itxns,
    mbr,
    routerOptedIn,
    receiverOptedIn,
    receiverAlgoNeededForClaim,
  ] = (await appClient.arc59GetSendAssetInfo({ asset: assetId, receiver }))
    .return!;

  // If the receiver has opted in, just send the asset directly
  if (receiverOptedIn) {
    await algorand.send.assetTransfer({
      sender,
      receiver,
      assetId,
      amount: 1n,
    });

    return;
  }

  // Create a composer to form an atomic transaction group
  const composer = appClient.compose();

  const signer = algorand.account.getSigner(sender);

  // If the MBR is non-zero, send the MBR to the router
  if (mbr || receiverAlgoNeededForClaim) {
    const mbrPayment = await algorand.transactions.payment({
      sender,
      receiver: arc59RouterAddress,
      amount: algokit.microAlgos(Number(mbr + receiverAlgoNeededForClaim)),
    });

    composer.addTransaction({ txn: mbrPayment, signer });
  }

  // If the router is not opted in, add a call to arc59OptRouterIn to do so
  if (!routerOptedIn) composer.arc59OptRouterIn({ asa: assetId });

  /** The box of the receiver's pubkey will always be needed */
  const boxes = [algosdk.decodeAddress(receiver).publicKey];

  /** The address of the receiver's inbox */
  const inboxAddress = (
    await appClient.compose().arc59GetInbox({ receiver }, { boxes }).simulate()
  ).returns[0];

  // The transfer of the asset to the router
  const axfer = await algorand.transactions.assetTransfer({
    sender,
    receiver: arc59RouterAddress,
    assetId,
    amount: 1n,
  });

  // An extra itxn is if we are also sending ALGO for the receiver claim
  const totalItxns = itxns + (receiverAlgoNeededForClaim === 0n ? 0n : 1n);

  composer.arc59SendAsset(
    { axfer, receiver, additionalReceiverFunds: receiverAlgoNeededForClaim },
    {
      sendParams: { fee: algokit.microAlgos(1000 + 1000 * Number(totalItxns)) },
      boxes, // The receiver's pubkey
      // Always good to include both accounts here, even if we think only the receiver is needed. This is to help protect against race conditions within a block.
      accounts: [receiver, inboxAddress],
      // Even though the asset is available in the group, we need to explicitly define it here because we will be checking the asset balance of the receiver
      assets: [Number(assetId)],
    }
  );

  // Disable resource population to ensure that our manually defined resources are correct
  algokit.Config.configure({ populateAppCallResources: false });

  // Send the transaction group
  await composer.execute();

  // Re-enable resource population
  algokit.Config.configure({ populateAppCallResources: true });
}
```

### TypeScript Claim Function

```ts
/**
 * Claim an asset from the ARC59 inbox
 *
 * @param appClient The ARC59 client generated by algokit
 * @param assetId The ID of the asset to claim
 * @param claimer The address of the account claiming the asset
 * @param algorand The AlgorandClient instance to use to send transactions
 */
async function arc59Claim(
  appClient: Arc59Client,
  assetId: bigint,
  claimer: string,
  algorand: algokit.AlgorandClient
) {
  const composer = appClient.compose();

  // Check if the claimer has opted in to the asset
  let claimerOptedIn = false;
  try {
    await algorand.account.getAssetInformation(claimer, assetId);
    claimerOptedIn = true;
  } catch (e) {
    // Do nothing
  }

  const inbox = (
    await appClient
      .compose()
      .arc59GetInbox({ receiver: claimer })
      .simulate({ allowUnnamedResources: true })
  ).returns[0];

  let totalTxns = 3;

  // If the inbox has extra ALGO, claim it
  const inboxInfo = await algorand.account.getInformation(inbox);
  if (inboxInfo.minBalance < inboxInfo.amount) {
    totalTxns += 2;
    composer.arc59ClaimAlgo(
      {},
      {
        sender: algorand.account.getAccount(claimer),
        sendParams: { fee: algokit.algos(0) },
      }
    );
  }

  // If the claimer hasn't already opted in, add a transaction to do so
  if (!claimerOptedIn) {
    composer.addTransaction({
      txn: await algorand.transactions.assetOptIn({ assetId, sender: claimer }),
      signer: algorand.account.getSigner(claimer),
    });
  }

  composer.arc59Claim(
    { asa: assetId },
    {
      sender: algorand.account.getAccount(claimer),
      sendParams: { fee: algokit.microAlgos(1000 * totalTxns) },
    }
  );

  await composer.execute();
}
```

## Security Considerations

The router application controls all user inboxes. If this contract is compromised, user assets might also be compromised.

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
