export enum ScopeType {
    AUTH,
    LSIG,
  }

export type StdDataStr = string
export type Ed25519Pk = Uint8Array
export type Signature = Uint8Array

export interface HdWalletMetadata {
    purpose: number;
    coinType: number;
    account: number;
    change: number;
    addrIdx: number;
}

export interface StdSignData {
    data: StdDataStr;
    signer: Ed25519Pk;
    hdPath?: HdWalletMetadata;
}

export interface StdSignMetadata {
    scope: ScopeType;
    schema: string;
    encoding?: string;
}

export type SignDataFunction = (
    signingData: StdSignData,
    metadata: StdSignMetadata,
) => Promise<(Signature | null)>

export interface ARC60SchemaType {
    ARC60Domain: string;
    bytes: Uint8Array;
}

export enum ApprovalOption {
    CONFIRM = "Confirm",
    REJECT = "Reject",
}