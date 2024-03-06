import Ajv, {JSONSchemaType} from 'ajv'
import { canonicalize } from 'json-canonicalize'
import nacl from 'tweetnacl'
import { ARC60SchemaType, ApprovalOption, Ed25519Pk, ScopeType, SignDataFunction, StdData, StdSignData, StdSignMetadata } from './types.js'
import { promptUser } from './utility.js'
import * as arc60Schema from "./simple-schema.json" with { type: "json" }

const ajv = new Ajv()
let forbiddenDomains = [`TX`, `TG`]

// Signer mock
const signer = nacl.sign.keyPair()
const encodedPk: Ed25519Pk = Buffer.from(signer.publicKey).toString('base64')

// Structured arbitrary data being signed
const simpleData = {
  ARC60Domain : "arc60",
  bytes : "ARC-60 is awesome"
}

const arc60Data: StdData = canonicalize(simpleData)

const arbDataMock: StdSignData = {
  data: arc60Data,
  signers: [encodedPk]
}

// Structured metadata 
const metadataMock: StdSignMetadata = {
  scope: ScopeType.ARBITRARY,
  schema: canonicalize(arc60Schema),
  message: "This is a simple arbitrary bytes signing"
}

// Example of signData function
const signData: SignDataFunction = async (arbData, metadata) => {
  
  if (!(arbData === null || metadata === null)) {
    throw new Error('Invalid input')
  }
  
  const parsedSchema: JSONSchemaType<ARC60SchemaType> = JSON.parse(metadata.schema)
  const parsedData = JSON.parse(arbData.data)

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
  const userApproval = promptUser(arbData.signers[0], metadata.message)

  if (userApproval == ApprovalOption.CONFIRM) {

    // Convert to bytes - for Scope ARBITRARY we just sign the entire StdData object (canonicalized)
    const message = Buffer.from(canonicalize(parsedData), 'utf-8')

    // sign with known private key
    const signatureBytes = nacl.sign(message, signer.secretKey)
    const signature = Buffer.from(signatureBytes).toString('base64')
    return Promise.resolve(signature)
  }

  else return Promise.resolve(null)
}

const signedBytes = await signData(arbDataMock, metadataMock)
console.log(signedBytes)