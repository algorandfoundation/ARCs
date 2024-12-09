---
arc: 34
title: xGov Pilot - Proposal Process
description: Criteria for the creation of proposals.
author: Stéphane Barroso (@SudoWeezy), Adriana Belotti, Michel Treccani
discussions-to: https://github.com/algorandfoundation/ARCs/issues/151
status: Deprecated
type: Meta
created: 2022-11-22
---

## Abstract

The Goal of this ARC is to clearly define the steps involved in submitting proposals for the xGov Program, to increase transparency and efficiency, ensuring all proposals are given proper consideration. The goal of this grants scheme is to fund proposals that will help us in our goal of increasing the adoption of the Algorand network, as the most advanced layer 1 blockchain to date. The program aims to fund proposals to develop open source software, including tooling, as well as educational resources to help inform and grow the Algorand community.

## Specification

The key words "**MUST**", "**MUST NOT**", "**REQUIRED**", "**SHALL**", "**SHALL NOT**", "**SHOULD**", "**SHOULD NOT**", "**RECOMMENDED**", "**MAY**", and "**OPTIONAL**" in this document are to be interpreted as described in <a href="https://www.ietf.org/rfc/rfc2119.txt">RFC-2119</a>.

### What is a proposal

The xGov program aims to provide funding for individuals or teams to:

- Develop of open source applications and tools (eg. an open source AMM or contributing content to an Algorand software library).
- Develop Algorand education resources, preferably in languages where the resources are not yet available(eg. a video series teaching developers about Algorand in Portuguese or Indonesian).

The remainder of the xGov program pilot will not fund proposals for:

- Supplying liquidity.
- Reserving funds to pay for ad-hoc open-source development (devs can apply directly for an xGov grant).
- Buying ASAs, including NFTs.

Proposals **SHALL NOT** be divided in small chunks.
> Issues requiring resolution may have been discussed on various online platforms such as forums, discord, and social media networks.
> Proposals requesting a large amount of funds **MUST BE** split into a milestone-based plan. See [Submit a proposal](./arc-0034.md#submit-a-proposal)

### Duty of a proposer

Having the ability to propose measures for a vote is a significant privilege, which requires:

- A thorough understanding of the needs of the community.
- Alignment of personal interests with the advancement of the Algorand ecosystem.
- Promoting good behavior amongst proposers and discouraging "gaming the system".
- Reporting flaws and discussing possible solutions with the AF team and community using either the Algorand Forum or the xGov Discord channels.

### Life of a proposal

The proposal process will follow the steps below:

- Anyone can submit a proposal at any time.
- Proposals will be evaluated and refined by the community and xGovs before they are available for voting.
- Up to one month is allocated for voting on proposals.
- The community will vote on proposals that have passed the refinement and temperature check stage.

> If too many proposals are received in a short period of time. xGovs can elect to close proposals, in order to be able to handle the volume appropriately.

### Submit a proposal

In order to submit a proposal, a proposer needs to create a pull request on the following repository: <a href="https://github.com/algorandfoundation/xGov">xGov Proposals</a>.

Proposals **MUST**:

- Be posted on the <a href="https://forum.algorand.org/">Algorand Forum</a> (using tags: Governance and xGov Proposals) and discussed with the community during the review phased. Proposals without a discussion thread WILL NOT be included in the voting session.
- Follow the [template form provided](../assets/arc-0034/TemplateForm.md), filling all the template sections
- Follow the rules of the xGov Proposals Repository.
- The minimum requested Amount is 10000 Algo
- Have the status `Final` before the end of the temperature check.
- Be either Proactive (the content of the proposal is yet to be created) or Retroactive (the content of the proposal is already created)
- Milestone-based grants must submit a proposal for one milestone at a time.
- Milestones need to follow the governance periods cycle. With the current 3-months cycle, a milestone could be 3-months, 6 months, 9 months etc.
- The proposal must display all milestones with clear deliverables and the amount requested must match the 1st milestone. If a second milestone proposal is submitted, it must display the first completed milestone, linking all deliverables. If a third milestone proposal is submitted, it must display the first and second completed milestone, linking all deliverables. This repeats until all milestones are completed.
- Funding will only be disbursed upon the completion of deliverables.
- A proposal must specify how its delivery can be verified,  so that it can be checked prior to payment.
- Proposals must include clear, non-technical descriptions of deliverables. We encourage the use of multimedia (blog/video) to help explain your proposal’s benefits to the community.
- Contain the maintenance period, availability, and sustainability plans. This includes information on potential costs and the duration for which services will be offered at no or reduced cost.

Proposals **MUST NOT**:

- Request funds for marketing campaigns or organizing future meetups.

> Each entity, individual, or project can submit at most two proposals (one proactive proposal and one retroactive proposal). Attempts to circumvent this rule may lead to disqualification or denial of funds.

### Disclaimer jurisdictions and exclusions

To be eligible to apply for a grant, projects must abide by the <a href="https://www.algorand.foundation/disclaimers">Disclaimers</a> (in particular the “Excluded Jurisdictions” section) and be willing to enter into <a href="https://drive.google.com/file/d/1dsKwQGhnS3h_PrSkoidhnvqlX7soLpZ-/view">a binding contract with the Algorand Foundation</a>.
Additionally, applications promoting gambling, adult content, drug use, and violence of any kind are not permitted.

> We are currently accepting grant applications from US-based individual/business. If the grant is approved, Algos will be converted to USDCa upon payment. This exception will be reviewed periodically.

### Voting Power

When an account participates in its first session, the voting power assigned to it will be equivalent to the total governance rewards it would have received. For all following sessions, the account's voting power will adjust based on the rewards lost by members in their pool who did not meet their obligations.

The voting power for an upcoming session is computed as:
`new_account_voting_power = (initial_pool_voting_power * initial_account_voting_power) / pool_voting_power_used`

Where:

- `new_account_voting_power`: Voting power allocated to an account for the next session.
- `initial_account_voting_power`: The voting power originally assigned to an account, based on the governance rewards.
- `initial_pool_voting_power`: The total voting power of the pool during its initial phase. This is the sum of governance rewards for all pool participants.
- `pool_voting_power_used`: The voting power from the pool that was actually used in the last session.

### Proposal Approval Threshold

In order for a proposal to be approved, it is necessary for the number of votes in favor of the proposal to be proportionate to the amount of funds requested. This ensures that the allocation of funds is in line with the community's consensus and in accordance with democratic principles.

The formula to calculate the voting power needed to pass a proposal is as follows:
`voting_power_needed = (amount_requested) / (amount_available) * (current_session_voting_power_used)`

Where:

- `voting_power_needed`: Voting power required for a proposal to be accepted.
- `amount_requested`: The requested amount a proposal is seeking.
- `amount_available`: The entire grant funds available for the current session.
- `current_session_voting_power_used`: The voting power used in the current session.

> eg. 2 000 000 Algo are available to be given away as grants, 300 000 000 Algo are committed to the xGov Process, 200 000 000 Algo are used during the vote:
>
> - Proposal A request 100 000 Algos (5 % of the Amount available)
> - Proposal A needs 5 % of the used votes (10 000 000 Votes) to go through

### Voting on proposal

At the start of the voting period xGovs [ARC-33](arc-0033.md) will vote on proposals using the voting tool hosted at <a href="https://xgov.algorand.foundation/">https://xgov.algorand.foundation/</a>.

Vote will refer to the PR number and a cid hash of the proposal itself.

The CID MUST:

- Represent the file.

- Be a version V1 CID
  - E.g., use the option --cid-version=1 of ipfs add

- Use SHA-256 hash algorithm
  - E.g., use the option --hash=sha2-256 of ipfs add

### Grants calculation

The allocation of grants will consider the funding request amounts and the available amount of ALGO to be distributed.

### Grants contract & payment

- Once grants are approved, the Algorand Foundation team will handle the applicable contract and payment.
- **Before submitting your grant proposal**, review the contract template and ensure you're comfortable with its terms: <a href="https://drive.google.com/file/d/1dsKwQGhnS3h_PrSkoidhnvqlX7soLpZ-/view">Contract Template</a>.

> For milestone-based grants, please also refer to the [Submit a proposal section](./arc-0034.md#submit-a-proposal)

## Rationale

The current status of the proposal process includes the following elements:

- Proposals will be submitted off-chain and linked to the on-chain voting through a hash.
- Projects that require multiple funding rounds will need to submit separate proposals.
- The allocation of funds will be subject to review and adjustment during each governance period.
- Voting on proposals will take place on-chain.

We encourage the community to continue to provide input on this topic through the submission of questions and ideas in this ARC document.

## Security Considerations

None

## Copyright

Copyright and related rights waived via <a href="https://creativecommons.org/publicdomain/zero/1.0/">CCO</a>.
