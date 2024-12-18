import { getPartKeyIntegrityHash } from "../dist/index.js";

const integrityHash = getPartKeyIntegrityHash({
    account: "OHQTAISSIGRGIGVN6TVJ6WYLBHFTUC437T4E2LRRXGWVNJ4GSZOXKPH7N4",
    selectionKeyB64: "e4kBLu7zXOorjLVzJHOiAn+IhOBsYBCqqHKaJCiCdJs=",
    voteKeyB64: "WWHePYtNZ2T3sHkqdd/38EvoFWrnIKPrTo6xN/4T1l4=",
    stateProofKeyB64: "1GdNPOck+t6yXvuXxrDEPKqgi4I2sTaNugV1kd5ksUW2G1U6x1FT0WR3aT3ZSSmbYoDt3cVrB3vIPJA8GkqSYg==",
    voteFirstValid: 3118965,
    voteLastValid: 4104516,
    keyDilution: 993,
})

const expectedIntegrityHash = "JUKBSNRYTU4PU";

console.log("Integrity hash:", integrityHash, "\nExpecting:", expectedIntegrityHash)

if(integrityHash !== expectedIntegrityHash) {
    throw new Error(`Error calculating integrity hash, expected: ${expectedIntegrityHash} received: ${integrityHash}`)
}

console.log("OK");