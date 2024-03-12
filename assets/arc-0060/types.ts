export enum ScopeType {
    ARBITRARY,
    LSIG,
  }

export type StdData = string
export type Ed25519Pk = Uint8Array

export interface StdSignMetadata {
    scope: ScopeType;
    schema: string;
    message?: string;
}

export type SignDataFunction = (
    data: string,
    metadata: StdSignMetadata,
    signer: Ed25519Pk,
) => Promise<(Uint8Array | null)>

export interface ARC60SchemaType {
    ARC60Domain: string;
    bytes: Uint8Array;
}

export enum ApprovalOption {
    CONFIRM = "Confirm",
    REJECT = "Reject",
}