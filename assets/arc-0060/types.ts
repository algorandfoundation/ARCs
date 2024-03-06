export enum ScopeType {
    ARBITRARY,
    LSIG,
  }

export type StdData = string
export type Ed25519Pk = string

export interface StdSignData {
    data: StdData;
    signers: Ed25519Pk[];
}

export interface StdSignMetadata {
    scope: ScopeType;
    schema: string;
    message?: string;
}

export type SignDataFunction = (
    arbData: StdSignData,
    metadata: StdSignMetadata,
) => Promise<(string | null)>

export interface ARC60SchemaType {
    ARC60Domain: string;
    bytes: string;
}

export enum ApprovalOption {
    CONFIRM = "Confirm",
    REJECT = "Reject",
}