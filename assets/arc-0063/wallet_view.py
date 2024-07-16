from algosdk import mnemonic, transaction, encoding, util
from algosdk.v2client import algod, indexer
import base64
from nacl.signing import SigningKey
import nacl.encoding
import hashlib

import json



def generate_32bytes_from_addresses(addr1, addr2, sk):
    combined = addr1 + addr2
    combined_signed = util.sign_bytes(combined.encode(), sk)
    hash_digest = hashlib.sha256(combined_signed.encode()).digest()
    seed = hash_digest[:32]
    sk = SigningKey(seed,encoder=nacl.encoding.RawEncoder)
    vk = sk.verify_key
    a = encoding.encode_address(vk.encode())
    return  base64.b64encode(sk.encode() + vk.encode()).decode(), a

algod_address = "http://localhost:4001"  # Adjust if using a different port
algod_token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
client = algod.AlgodClient(algod_token, algod_address)
indexer_address = "http://localhost:8980"  # Adjust if using a different port
indexer_token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
indexer = indexer.IndexerClient(indexer_token, indexer_address)

add = [
    {
        "address": "45UO5ZGAAV3VSUFWPY72UITVNSWKLSYJBBALU2O56E32QQYHXHCI5D2PDA",
        "mnemonic": "dumb pencil plastic isolate butter ribbon glide tragic pulse empty grape double glass stadium disorder riot agent donkey city weird shadow bubble ladder absent kidney",
    }
]

bob_sk, bob_addr = mnemonic.to_private_key(add[0]["mnemonic"]), add[0]["address"]

acc_info = indexer.search_transactions_by_address(bob_addr)

b_64_prefix = base64.b64encode(b'arc_63:j:')


for t in acc_info['transactions']:
    if 'note' in t:
        try:
            if b_64_prefix == str.encode(t['note'][:len(b_64_prefix)]):
                vault_info = base64.b64decode(t['note'])[9:].decode('utf-8')
                json_vault_info = json.loads(vault_info)
                print(vault_info)
                txns = indexer.search_transactions_by_address(json_vault_info['vault'])["transactions"]
                for txn in txns:
                    if 'signature' in txn and 'logicsig' in txn['signature'] and 'logic' in txn['signature']['logicsig']  :
                        msig_info = txn['signature']['logicsig']['multisig-signature']['subsignature']
                        logic = txn['signature']['logicsig']['logic']

                        # Decode the Base64 string to get the bytecode
                        decoded_logic = base64.b64decode(logic)

                        teal_code = client.disassemble(decoded_logic)
                        # print("Disassembled TEAL Code:")
                        # print(teal_code['result'])
                        lsig = transaction.LogicSigAccount(decoded_logic)
                        
                        assert (lsig.address() ==  json_vault_info['lsigs'])
                        assert (encoding.encode_address(base64.b64decode(msig_info[0]['public-key'])) 
                                ==  json_vault_info['s'])
                        assert (encoding.encode_address(base64.b64decode(msig_info[1]['public-key'])) 
                                ==  lsig.address())
                        
                        _, plug_in_address = generate_32bytes_from_addresses(bob_addr, lsig.address(), bob_sk)

                        assert (encoding.encode_address(base64.b64decode(msig_info[2]['public-key'])) 
                                ==  plug_in_address)
                        assert(msig_info[2]['signature'] == json_vault_info['sigS'])
                        print("OK")
                        break
        except json.decoder.JSONDecodeError:  # noqa: E722
            pass

