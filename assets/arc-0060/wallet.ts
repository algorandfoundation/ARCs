import Ajv, {JSONSchemaType} from 'ajv'
import { canonicalize } from 'json-canonicalize'
import nacl from 'tweetnacl'
import { ARC60SchemaType, ApprovalOption, Ed25519Pk, ScopeType, SignDataFunction, StdData, StdSignMetadata } from './types.js'
import { promptUser } from './utility.js'
import * as arc60Schema from "./simple-schema.json" with { type: "json" }

const ajv = new Ajv()
let forbiddenDomains = ["TX", "TG"]
let allowedDomains = ["", "arc60"]

// Signer mock
const keypair = nacl.sign.keyPair()
const signerPk: Ed25519Pk = keypair.publicKey

// Structured arbitrary data being signed
const simpleData = {
  ARC60Domain : "arc60",
  bytes : "ARC-60 is awesome"
}

const arc60Data: StdData = canonicalize(simpleData)


// Structured metadata 
const metadataMock: StdSignMetadata = {
  scope: ScopeType.MSGSIG,
  schema: canonicalize(arc60Schema),
  message: "This is a simple message signing"
}

// Example of signData function
const signData: SignDataFunction = async (data, metadata, signer) => {

  // Check null values
  if (data === null || metadata === null || signer === null) {
    throw new Error('Invalid input')
  }

  const parsedSchema: JSONSchemaType<ARC60SchemaType> = JSON.parse(metadata.schema)
  const parsedData = JSON.parse(data)

  console.log(parsedSchema)
  console.log(parsedData)

  // Check for forbidden domain separators
  if(forbiddenDomains.includes(parsedData.ARC60Domain)) {
    throw new Error('Invalid input')
  }

  // Check domain separator consistency
  if (metadata.scope === ScopeType.MSGSIG && !(allowedDomains.includes(parsedData.ARC60Domain))) {
    throw new Error('Invalid input')
  }

  // Validate the schema
  const validate = ajv.compile<ARC60SchemaType>(parsedSchema)
  const isValid = validate(parsedData)

  if (!isValid) {
    throw new Error('Invalid input')
  }

  // Validate bytes
  const tag = Buffer.from(parsedData.bytes.slice(0, 2)).toString()
  if (forbiddenDomains.includes(tag)) {
    throw new Error('Invalid input')
  }

  // Compute signData
  const signData =  Buffer.from(parsedData.ARC60Domain + parsedData.bytes)

  // Simulate user approval
  const userApproval = promptUser(signData, signer, metadata.message)

  if (userApproval == ApprovalOption.CONFIRM) {
    // sign with known private key
    const signatureBytes = nacl.sign(signData, keypair.secretKey)
    return Promise.resolve(signatureBytes)
  }

  else return Promise.resolve(null)
}

function verifySignature(sig: Uint8Array, pk: Ed25519Pk) {
  return nacl.sign.open(sig, pk)
}


const signedBytes = await signData(arc60Data, metadataMock, signerPk)
if (!(signedBytes === null)) {
  const signature = Buffer.from(signedBytes).toString('base64')
  console.log(`Signature: ${signature}`)

  const verifiedBytes = verifySignature(signedBytes, signerPk)
  if (verifiedBytes != null) {
    const msg = Buffer.from(verifiedBytes).toString('utf-8')
    console.log(`Verified signature for message: ${msg}`)
  } else {
    throw new Error('Signature cannot be verified.')
  }
}

