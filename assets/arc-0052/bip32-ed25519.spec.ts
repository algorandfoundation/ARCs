import { crypto_scalarmult_ed25519_base_noclamp, ready } from "libsodium-wrappers-sumo"
import { deriveChildNodePrivate, deriveChildNodePublic, fromSeed, trunc_256_minus_g_bits } from "./bip32-ed25519"
import * as bip39 from "bip39"
import { harden } from "./contextual.api.crypto"

// function that prints bits
const printBits = (arr: Uint8Array) => {
    const bits = arr.reduce((acc, byte) => {
        return acc + byte.toString(2).padStart(8, "0") + " ";
    }, "")

    console.log("bits", bits)
}


describe("BIP32-Ed25519", () => {
    describe("zero last number of bits", () => {
        it("", () => {
            // create 32 byte array with values
            const input = new Uint8Array(32); // Create a 32-byte (256-bit) array
            for (let i = 0; i < 32; i++) {
                input[i] = 0xFF; // Initialize all bytes to 0xFF for demonstration
        }

            // log bits binary with space every 8 bits
            const inputBits = input.reduce((acc, byte) => {
                return acc + byte.toString(2).padStart(8, "0") + " ";
            }, "");

            console.log("bits", inputBits)

            const extracted: Uint8Array = trunc_256_minus_g_bits(input, 9);

            // log bits binary with space every 8 bits
            const bits = extracted.reduce((acc, byte) => {
                return acc + byte.toString(2).padStart(8, "0") + " ";
            }, "");

            console.log("bits", bits)
        })
    })

    // describe group for N levels of derivation
    describe.only("For N levels of derivation check that secret keys match public keys", () => {
        let bip39Mnemonic: string = "salon zoo engage submit smile frost later decide wing sight chaos renew lizard rely canal coral scene hobby scare step bus leaf tobacco slice"
        let seed: Buffer
    
        beforeAll(() => {
            seed = bip39.mnemonicToSeedSync(bip39Mnemonic, "")
        })
        
        it("Derive 7 levels with g == Peikeirt == 9, private and public derivations, check matching", async () => {
            await ready // libsodium ready

            const rootKey = fromSeed(seed)
            // 44'/283'/0'/0
            const bip44Path = [harden(44), harden(283), harden(0), 0, 0]

            await ready // libsodium

            // Pick `g`, which is amount of bits zeroed from each derived node
            const g: number = 9
    
            let derivationNode: Uint8Array = rootKey
            for (let i = 0; i < bip44Path.length + 85; i++) {
                derivationNode = await deriveChildNodePrivate(derivationNode, bip44Path[i], g)
            }

            // top end of the field Ed25519
            const maxValue: bigint = 2n ** 255n - 19n

            // derivationNode = await deriveChildNodePrivate(derivationNode, 0, g)

            // extendedPublicNodeFormat
            const extendedPublicDerivationNode = new Uint8Array(Buffer.concat([crypto_scalarmult_ed25519_base_noclamp(derivationNode.subarray(0, 32)), derivationNode.subarray(64, 96)]))

            let derivationPublicNode = await deriveChildNodePublic(extendedPublicDerivationNode, 0, g)
            derivationNode = await deriveChildNodePrivate(derivationNode, 0, g)

            // assuming in a third party that only has public information
            // I'm provided with the wallet level m'/44'/283'/0'/0 root [public, chaincode]
            // no private information is shared
            // i can SOFTLY derive N public keys / addresses from this root
            const derivedKey: Uint8Array = derivationPublicNode.subarray(0, 32)

            // print bits
            printBits(derivedKey)

            const scalar: Uint8Array = derivationNode.subarray(0, 32)

            // scalar print bits
            printBits(scalar)

            // check if scalar is smaller than the top end of the field
            expect(BigInt(`0x${Buffer.from(scalar).toString("hex")}`)).toBeLessThan(maxValue)

            const myKey: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(scalar)

            // print bits
            printBits(myKey)
                
            // they should match 
            // derivedKey.subarray(0, 32) ==  public key (excluding chaincode)
            expect(derivedKey).toEqual(myKey)

        })
    })
})