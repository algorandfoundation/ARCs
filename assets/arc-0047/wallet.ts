import algosdk from 'algosdk'
import { sha256 } from 'js-sha256'
import { canonicalize } from 'json-canonicalize'

const algodClient = new algosdk.Algodv2('a'.repeat(64), 'http://localhost', 4001)

const allowList: string[] = ['032c6b017bdad49f54d41170fef9c13acdb8e5ff9a76fcaf0cfbecb6b3fdb5d0']

const mockRequest = [ "{\"description\":\"Allows a payment to be made every 25000 blocks of a specific amount to a specific address\",\"name\":\"25000 block payment\",\"program\":\"I3ByYWdtYSB2ZXJzaW9uIDkKCi8vIFZlcmlmeSB0aGlzIGlzIGEgcGF5bWVudAp0eG4gVHlwZUVudW0KaW50IHBheQo9PQoKLy8gVmVyaWZ5IHRoaXMgaXMgbm90IHJla2V5aW5nIHRoZSBzZW5kZXIgYWRkcmVzcwp0eG4gUmVrZXlUbwpnbG9iYWwgWmVyb0FkZHJlc3MKPT0KYXNzZXJ0CgovLyBWZXJpZnkgdGhlIHNlbmRlcidzIGFjY291bnQgaXMgbm90IGJlaW5nIGNsb3NlZAp0eG4gQ2xvc2VSZW1haW5kZXJUbwpnbG9iYWwgWmVyb0FkZHJlc3MKPT0KYXNzZXJ0CgovLyBWZXJpZnkgdGhlIHJlY2VpdmVyIGlzIGVxdWFsIHRvIHRoZSB0ZW1wbGF0ZWQgcmVjZWl2ZXIgYWRkcmVzcwp0eG4gUmVjZWl2ZXIKYWRkciBUTVBMX1JFQ0VJVkVSCj09CmFzc2VydAoKLy8gVmVyaWZ5IHRoZSBhbW91bnQgaXMgZXF1YWwgdG8gdGhlIHRlbXBsYXRlZCBhbW91bnQKdHhuIEFtb3VudAppbnQgVE1QTF9BTU9VTlQKPT0KYXNzZXJ0CgovLyBWZXJpZnkgdGhlIGN1cnJlbnQgcm91bmQgaXMgd2l0aGluIDUwMCByb3VuZHMgb2YgYSBwcm9kdWN0IG9mIDI1XzAwMApnbG9iYWwgUm91bmQKaW50IDI1XzAwMAolCnN0b3JlIDAKCmxvYWQgMAppbnQgNTAwCjw9Cgpsb2FkIDAKaW50IDI0XzUwMAo+PQoKfHwKYXNzZXJ0CgovLyBWZXJpZnkgbGVhc2UgCnR4biBMZWFzZQpieXRlICJzY2hlZHVsZWQgMjVfMDAwIHBheW1lbnQiCnNoYTI1Ngo9PQo=\",\"variables\":[{\"description\":\"Amount of the payment transaction in microAlgos\",\"name\":\"Payment Amount\",\"type\":\"uint64\",\"variable\":\"TMPL_AMOUNT\"},{\"description\":\"Address to which the payment transaction is sent\",\"name\":\"Payment Receiver\",\"type\":\"address\",\"variable\":\"TMPL_RECEIVER\"}]}",
"{\"TMPL_AMOUNT\":1000000,\"TMPL_RECEIVER\":\"Y76M3MSY6DKBRHBL7C3NNDXGS5IIMQVQVUAB6MP4XEMMGVF2QWNPL226CA\"}",
"6INR7PDVBEVPFXMYOWG2J7KLGMQUWKB7CFX3KW2ERW4E42NW5R7WVB4R3A" ]

async function processTemplatedLsig(requestParams: string[], signer: (lsig: algosdk.LogicSig) => Promise<Uint8Array>): Promise<Uint8Array> {
    const arc47 = JSON.parse(requestParams[0])
    const values = JSON.parse(requestParams[1])
    const hash = requestParams[2]

    // allowList is not a required feature of ARC47, but it allows wallets to verify the lsig template before signing
    if (!allowList.includes(sha256(canonicalize(arc47)))) throw Error('Templated Lsig not in allow list')

    // base64 decode the program
    let finalTeal = atob(arc47.program)

    // substitute the variables
    for (const variable in values) {
        finalTeal = finalTeal.replaceAll(variable, values[variable].toString())
    }

    // use algod to compile the TEAL after substituting the variables
    const compileResponse = await algodClient.compile(finalTeal).do()

    // verify the compiled hash is the same as the given hash in the request
    if (compileResponse.hash !== hash) throw Error(`Compiled hash (${compileResponse.hash}) does not match expected hash (${hash})`)

    // create a LogicSig object from the compiled program
    const lsig = new algosdk.LogicSig(Buffer.from(compileResponse.result, 'base64'))

    // signer function is expected to return the signature of the lsig
    return signer(lsig)
}

const signature = await processTemplatedLsig(mockRequest, async (lsig) => lsig.signProgram((algosdk.generateAccount()).sk))

console.log(signature)
