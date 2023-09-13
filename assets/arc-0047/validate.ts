import { readFileSync } from "fs"
import algosdk from 'algosdk'

function validateTEAL(teal: string) {
    const validMatches = teal.match(/(int|byte|addr) TMPL_[A-Z]+(\n|;)/g)?.length || 0
    const totalMatches = teal.match(/TMPL_/g)?.length || 0
    if (validMatches !== totalMatches) throw Error(`TEAL contains invalid TMPL_ variable(s)`)
}

function validateValue(value: string | number, type: string): void {
    if (type === 'address') {
        if (typeof value !== 'string') throw Error('address must be a string')
        algosdk.decodeAddress(value)
        return
    }
    

    if (['uint64', 'asset', 'application'].includes(type)) {
        if (typeof value !== 'number') throw Error(`${type} must be a number`)
        algosdk.encodeUint64(value)
        return
    }

    if (type === 'string') {
        if (typeof value !== 'string') throw Error('string must be a string')
        if (value.match(/(?!<=\\)"/)?.length) throw Error('string contains unescaped double quote')
        return
    }

    if (typeof value !== 'string' || value.match(/^0x/) === undefined) throw Error(`${type} must be a hex string prefixed with '0x'`)
    const abiType = algosdk.ABIType.from(type)
    const hex = (value as string).replace(/^0x/, '')

    abiType.decode(Buffer.from(hex, 'hex'))
}

validateTEAL(readFileSync('./lsig.teal').toString())
validateValue('0x01', 'uint8')
validateValue(1, 'uint64')
validateValue(algosdk.generateAccount().addr, 'address')
validateValue('hello', 'string')

try {
    validateValue('hello"', 'string')
} catch (e) {
    console.log(`Caught ${e}`)
}

try {
    validateValue(1, 'uint8')
} catch (e) {
    console.log(`Caught ${e}`)
}

try {
    validateValue('01', 'uint8')
} catch (e) {
    console.log(`Caught ${e}`)
}

try {
    validateValue('01', 'address')
} catch (e) {
    console.log(`Caught ${e}`)
}

const badTeal = 'int TMPL_FOO;byte "TMPL_BAR"'

try {
    validateTEAL(badTeal)
} catch (e) {
    console.log(`Caught ${e}`)
}