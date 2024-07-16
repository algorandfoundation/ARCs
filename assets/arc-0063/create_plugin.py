from algosdk import mnemonic, transaction, encoding, util, constants
from algosdk.v2client import algod, indexer
from algosdk.transaction import  AssetTransferTxn, AssetCloseOutTxn
from algosdk.transaction import LogicSigTransaction
from pyteal import Txn, And, TxnType, Int,Global, compileTeal, Mode
import base64
from nacl.signing import SigningKey
import nacl.encoding
import hashlib

import json

algod_address = "http://localhost:4001"  # Adjust if using a different port
algod_token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
client = algod.AlgodClient(algod_token, algod_address)
indexer_address = "http://localhost:8980"  # Adjust if using a different port
indexer_token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
indexer = indexer.IndexerClient(indexer_token, indexer_address)
sp = client.suggested_params()

def generate_32bytes_from_addresses(addr1, addr2, sk):
    combined = addr1 + addr2
    combined_signed = util.sign_bytes(combined.encode(), sk)
    hash_digest = hashlib.sha256(combined_signed.encode()).digest()
    seed = hash_digest[:32]
    sk = SigningKey(seed,encoder=nacl.encoding.RawEncoder)
    vk = sk.verify_key
    a = encoding.encode_address(vk.encode())
    return  base64.b64encode(sk.encode() + vk.encode()).decode(), a

def opt_in_logic_sig():
    return And(
        Txn.type_enum() == TxnType.AssetTransfer,
        Txn.asset_amount() == Int(0),
        Txn.rekey_to() == Global.zero_address(),
        Txn.fee() == Global.min_txn_fee()
    )
teal_program = compileTeal(opt_in_logic_sig(), Mode.Signature, version=10)
compiled_program = client.compile(teal_program)
program = base64.b64decode(compiled_program["result"])
lsig = transaction.LogicSigAccount(program)

add = [
    {
        "address": "45UO5ZGAAV3VSUFWPY72UITVNSWKLSYJBBALU2O56E32QQYHXHCI5D2PDA",
        "mnemonic": "dumb pencil plastic isolate butter ribbon glide tragic pulse empty grape double glass stadium disorder riot agent donkey city weird shadow bubble ladder absent kidney",
    },
    {
        "address": "6QZBRTHUT4P4D26HBW7NSJJ26P3WV4NWXLBF7AB5TVDVJFXLFLN6RMZQKI",
        "mnemonic": "sense gate people glare window bright betray tiny group subject blast gasp cargo safe play news inhale evolve luggage coil biology wide custom absorb trust",
    }
]

bob_sk, bob_addr = mnemonic.to_private_key(add[0]["mnemonic"]), add[0]["address"]
alice_sk, alice_addr = mnemonic.to_private_key(add[1]["mnemonic"]), add[1]["address"]

    #******************** Plug_IN_Signer Account Generation ****************************#
plug_in_sk, plug_in_addr  = generate_32bytes_from_addresses(bob_addr, lsig.address(), bob_sk)

    #******************** Plug_IN_Signer Public Signature   ****************************#
public_key, secret_key = nacl.bindings.crypto_sign_seed_keypair(base64.b64decode(plug_in_sk)[: constants.key_len_bytes])
message = constants.logic_prefix + program
raw_signed = nacl.bindings.crypto_sign(message, secret_key)
crypto_sign_BYTES = nacl.bindings.crypto_sign_BYTES
signature = nacl.encoding.RawEncoder.encode(raw_signed[:crypto_sign_BYTES])
plug_in_public_sig = base64.b64encode(signature).decode()
    #******************** REKEY PLUG_IN TO _ ****************************# #Alice for testing purposes but should be 0 address
ptxn = transaction.PaymentTxn(
    bob_addr, sp, plug_in_addr, int(1e5 + 1e3)
).sign(bob_sk)
txid = client.send_transaction(ptxn)
results = transaction.wait_for_confirmation(client, txid, 4)
print(f"Result confirmed in round: {results['confirmed-round']}")

rekey_txn = transaction.PaymentTxn(
    plug_in_addr, sp, plug_in_addr, 0, rekey_to=alice_addr
)
signed_rekey = rekey_txn.sign(plug_in_sk)
txid = client.send_transaction(signed_rekey)
result = transaction.wait_for_confirmation(client, txid, 4)
print(f"Rekey transaction confirmed in round {result['confirmed-round']}")



    #********************* MSIG VAULT  Generation **************************************#
bob_vault_msig = transaction.Multisig(1,1,[bob_addr, lsig.address(), plug_in_addr])


print("Bob vault addresses : ", bob_vault_msig.address()) #NUVYGSZCMMH65PGYGPRB63JNZMMIKT6ROK5HNM3BKVKLSNL77FTB7DKMJU
for i in bob_vault_msig.subsigs:
    print("Bob vault address: ", encoding.encode_address(i.public_key), base64.b64encode(i.public_key))


# print(int(client.account_info(bob_vault_msig.address())["amount"]))

if int(client.account_info(bob_vault_msig.address())["amount"]) < 2e5:
    #********************FUND BOB MSIG VAULT****************************#
    note_field = json.dumps({
    "s": bob_addr,
    "lsigs": lsig.address(),
    "sigA": plug_in_addr,
    "sigS": plug_in_public_sig,
    "vault": bob_vault_msig.address(),
    })
    ptxn = transaction.PaymentTxn(
        bob_addr, sp, bob_vault_msig.address(), int(1e6), note=f"arc_63:j:{note_field}"
    ).sign(bob_sk)
    txid = client.send_transaction(ptxn)
    results = transaction.wait_for_confirmation(client, txid, 4)
    print(f"Result confirmed in round: {results['confirmed-round']}")

alice_info = client.account_info(alice_addr)
if 'assets' in alice_info and (len(alice_info['assets']) > 0):
    a_id = client.account_info(alice_addr)['assets'][0]['asset-id']
else:
    #******************** ALICE CREATE ASA ****************************#
    print("Alice Create ASA")
    actxn = transaction.AssetConfigTxn( sender=alice_addr, sp=sp, default_frozen=False, unit_name="rug2", asset_name="2 Really Useful Gift", manager=alice_addr, reserve=alice_addr, freeze=alice_addr, clawback=alice_addr, url="https://path/to/my/asset/details", total=10, decimals=0, )
    sactxn = actxn.sign(alice_sk)
    tx_id = client.send_transaction(sactxn)
    print(f"Sent asset create transaction with txid: {tx_id}")
    # Wait for the transaction to be confirmed
    results = transaction.wait_for_confirmation(client, tx_id, 4)
    a_id = results['asset-index']
    print(f"Result confirmed in round: {results['confirmed-round']} ASA ID : {results['asset-index']}")
print(f'Alice created 10 ASA: {a_id}')


    #******************** MSIG VAULT OPT IN ****************************#
optin_txn = AssetTransferTxn(
    sender=bob_vault_msig.address(),
    sp=sp,
    receiver=bob_vault_msig.address(),
    amt=0,
    index=a_id,
)
lsig.lsig.msig = bob_vault_msig
lsig.lsig.msig.subsigs[2].signature = base64.b64decode(plug_in_public_sig) # signature from plug_in_public
lstx = LogicSigTransaction(optin_txn, lsig)



optin_txid = client.send_transaction(lstx)


print(f"Sent Msig Vault Opt-in with txid: {optin_txid}")
# Wait for the transaction to be confirmed
results = transaction.wait_for_confirmation(client, optin_txid, 4)
print(f"Result confirmed in round: {results['confirmed-round']}")
print("    #********************Opt Out ASA***************************#")
bob_vault_msig = transaction.Multisig(
    1,
    1,
    [bob_addr, transaction.LogicSigAccount(program).address(), plug_in_addr]
)
print(transaction.LogicSigAccount(program).address())

optout_txn = AssetCloseOutTxn(
    sender=bob_vault_msig.address(),
    sp=sp,
    receiver=bob_vault_msig.address(),
    index=a_id
)

msig_txn = transaction.MultisigTransaction(optout_txn, bob_vault_msig)
msig_txn.sign(bob_sk)
optout_txn = client.send_transaction(msig_txn)

print(f"Sent Msig Vault Opt-out with txid: {optout_txn}")
# Wait for the transaction to be confirmed
results = transaction.wait_for_confirmation(client, optout_txn, 4)
print(f"Result confirmed in round: {results['confirmed-round']}")
print(f"{bob_vault_msig.address()} Opted - Out {a_id}")

    #********************REKEY BACK************************#
rekey_back_txn = transaction.PaymentTxn(
    plug_in_addr, sp, plug_in_addr, 0, rekey_to=plug_in_addr, close_remainder_to=bob_addr
)
signed_rekey_back = rekey_back_txn.sign(alice_sk)
txid = client.send_transaction(signed_rekey_back)
result = transaction.wait_for_confirmation(client, txid, 4)
print(f"Rekey back transaction confirmed in round {result['confirmed-round']}")