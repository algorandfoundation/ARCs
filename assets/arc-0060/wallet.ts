import Ajv, {JSONSchemaType} from 'ajv'
import nacl from 'tweetnacl'
import { ARC60SchemaType, ApprovalOption, Ed25519Pk, ScopeType, SignDataFunction, StdDataStr, StdSignData, StdSignMetadata } from './types.js'
import { promptUser, signingMessage } from './utility.js'
import * as arc60Schema from "./auth-schema.json" with { type: "json" }

const ajv = new Ajv()
let forbiddenDomains = ["TX", "TG"]
let allowedDomains = ["", "arc60"]
let mockMsg = "arc60176,34,195,92,88,19,199,5,244,77,96,7,209,123,229,94,218,245,31,159,12,57,75,89,250,201,173,66,96,84,28,78"

// Signer mock
const keypair = nacl.sign.keyPair()
const signerPk: Ed25519Pk = keypair.publicKey

// Structured arbitrary data being signed
const simpleDataJson = {
  ARC60Domain : "arc60",
  bytes : "sCLDXFgTxwX0TWAH0XvlXtr1H58MOUtZ+smtQmBUHE4=" // 32 bytes challenge encoded base64
}

const simpleData: StdDataStr = JSON.stringify(simpleDataJson)

const signingDataMock: StdSignData = {
  data: simpleData,
  signer: signerPk,
}

// Structured metadata 
const metadataMock: StdSignMetadata = {
  scope: ScopeType.AUTH,
  schema: JSON.stringify(arc60Schema)
}

// Example of signData function
const signData: SignDataFunction = async (signingData, metadata) => {

  const data = signingData.data
  const signer = signingData.signer
  
  // Check null values
  if (data === null || metadata === null || signer === null) {
    throw new Error('Invalid input')
  }

  const parsedSchema: JSONSchemaType<ARC60SchemaType> = JSON.parse(metadata.schema)
  const parsedData = JSON.parse(data)

  console.log(parsedSchema)
  console.log(parsedData)

  // Decode bytes according to ScopeType - default encoding is base64
  if (!metadata.encoding && metadata.scope === ScopeType.AUTH) {
    const base64Bytes = parsedData.bytes
    const bytes = Array.from(Buffer.from(base64Bytes, 'base64'))
    parsedData.bytes = bytes
  }

  // Validate the schema
  const validate = ajv.compile<ARC60SchemaType>(parsedSchema)
  const isValid = validate(parsedData)

  if (!isValid) {
    throw new Error('Invalid input')
  }

  // Check for forbidden domain separators
  if(forbiddenDomains.includes(parsedData.ARC60Domain)) {
    throw new Error('Invalid input')
  }

  // Check domain separator consistency
  if (metadata.scope === ScopeType.AUTH && !(allowedDomains.includes(parsedData.ARC60Domain))) {
    throw new Error('Invalid input')
  }

  // bytes cannot be a transaction
  const tag = Buffer.from(parsedData.bytes.slice(0, 2)).toString()
  if (forbiddenDomains.includes(tag)) {
    throw new Error('Invalid input')
  }

  // Compute msg
  const msg =  Buffer.from(parsedData.ARC60Domain + parsedData.bytes)

  // Generate warn message to display
  const warn_message = signingMessage(metadata.scope, signer, msg)

  // Simulate user approval
  const userApproval = promptUser(msg, signer, warn_message)

  if (userApproval == ApprovalOption.CONFIRM) {
    // sign with known private key
    const signatureBytes = nacl.sign.detached(msg, keypair.secretKey)
    return Promise.resolve(signatureBytes)
  }

  else return Promise.resolve(null)
}

function verifySignature(msg: Uint8Array, sig: Uint8Array, pk: Ed25519Pk) {
  // verify the signature with public key
  return nacl.sign.detached.verify(msg, sig, pk)
}


const signedBytes = await signData(signingDataMock, metadataMock)

if (!(signedBytes === null)) {
  const signature = Buffer.from(signedBytes).toString('base64')
  console.log(`Signature: ${signature}`)

  const verifiedBytes = verifySignature(Buffer.from(mockMsg), signedBytes, signerPk)
  if (verifiedBytes) {
    console.log(`Signature verified.`)
  } else {
    throw new Error('Signature cannot be verified.')
  }
}