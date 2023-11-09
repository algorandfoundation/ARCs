import algosdk from 'algosdk'
import { readFileSync } from 'fs'
import {canonicalize} from 'json-canonicalize'
import { formatJsonRpcRequest } from "@json-rpc-tools/utils";
import { sha256 } from 'js-sha256';

const algodClient = new algosdk.Algodv2('a'.repeat(64), 'http://localhost', 4001)

const teal = readFileSync('./lsig.teal').toString()

const arc47 = {
    name: "25000 block payment",
    description: "Allows a payment to be made every 25000 blocks of a specific amount to a specific address",
    program: btoa(teal),
    variables: [
        {
            name: "Payment Amount",
            variable: "TMPL_AMOUNT",
            type: "uint64",
            description: "Amount of the payment transaction in microAlgos"
        },
        {
            name: "Payment Receiver",
            variable: "TMPL_RECEIVER",
            type: "address",
            description: "Address to which the payment transaction is sent"
        }
    ],
}

const values: Record<string, string | number> = {
    TMPL_AMOUNT: 1000000,
    TMPL_RECEIVER: 'Y76M3MSY6DKBRHBL7C3NNDXGS5IIMQVQVUAB6MP4XEMMGVF2QWNPL226CA'
}

let finalTeal = teal

for (const variable in values) {
    finalTeal = finalTeal.replaceAll(variable, values[variable].toString())
}

const result = await algodClient.compile(finalTeal).do()

const requestParams = [canonicalize(arc47), JSON.stringify(values), result.hash]
    
const walletConnectRequest = formatJsonRpcRequest('algo_signTemplatedLsig', requestParams);

console.log(`Request Params: ${console.log(requestParams)}`)
console.log(`ARC47 SHA256: ${sha256(canonicalize(arc47))}`)