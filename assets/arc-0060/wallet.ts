import Ajv, {JSONSchemaType} from 'ajv'
import { canonicalize } from 'json-canonicalize'
import nacl from 'tweetnacl'
import { ARC60SchemaType, ApprovalOption, Ed25519Pk, ScopeType, SignDataFunction, StdData, StdSignMetadata } from './types.js'
import { promptUser } from './utility.js'
import * as arc60Schema from "./simple-schema.json" with { type: "json" }

const ajv = new Ajv()
let forbiddenDomains = [`TX`, `TG`]

// Signer mock
const keypair = nacl.sign.keyPair()
const signerPk: Ed25519Pk = keypair.publicKey

// Structured arbitrary data being signed
const simpleData = {
  ARC60Domain : "arc60",
  bytes : [65, 82, 67, 45, 54, 48, 32, 105, 115, 32, 97, 119, 101, 115, 111, 109, 101] //"ARC-60 is awesome"
}

const arc60Data: StdData = canonicalize(simpleData)


// Structured metadata 
const metadataMock: StdSignMetadata = {
  scope: ScopeType.ARBITRARY,
  schema: canonicalize(arc60Schema),
  message: "This is a simple arbitrary bytes signing"
}

// Example of signData function
const signData: SignDataFunction = async (data, metadata, signer) => {
  
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
  if (metadata.scope === ScopeType.ARBITRARY && !(parsedData.ARC60Domain === "arc60")) {
    throw new Error('Invalid input')
  }

  // Validate the schema
  const validate = ajv.compile<ARC60SchemaType>(parsedSchema)
  const isValid = validate(parsedData)

  if (!isValid) {
    throw new Error('Invalid input')
  }

  // Simulate user approval
  const userApproval = promptUser(signer, metadata.message)

  if (userApproval == ApprovalOption.CONFIRM) {
    // sign with known private key
    const signingBytes = new Uint8Array(parsedData.bytes)
    const signatureBytes = nacl.sign(signingBytes, keypair.secretKey)
    return Promise.resolve(signatureBytes)
  }

  else return Promise.resolve(null)
}

const signedBytes = await signData(arc60Data, metadataMock, signerPk)
if (!(signedBytes === null)) {
  const signature = Buffer.from(signedBytes).toString('base64')
  console.log(signature)
}