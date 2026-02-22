from algosdk import mnemonic, transaction, encoding, constants, v2client, account
from typing import Dict, Any
import base64
import nacl
import json

algod_address = "http://localhost:4001"  # Adjust if using a different port
algod_token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
client = v2client.algod.AlgodClient(algod_token, algod_address)
indexer_address = "http://localhost:8980"  # Adjust if using a different port
indexer_token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
indexer = v2client.indexer.IndexerClient(indexer_token, indexer_address)
sp = client.suggested_params()

teal_program = """
#pragma version 10
txn TypeEnum
pushint 4
==
txn AssetAmount
pushint 0
==
&&
txn AssetCloseTo
global ZeroAddress
==
&&
txn RekeyTo
global ZeroAddress
==
&&
txn Fee
global MinTxnFee
==
&&
return
"""

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

owner_sk, owner_addr = mnemonic.to_private_key(add[0]["mnemonic"]), add[0]["address"]
asa_creator_sk, asa_creator_addr = mnemonic.to_private_key(add[1]["mnemonic"]), add[1]["address"]

    #******************** Plug_IN_Signer Account Generation ****************************#
if True:
    plug_in_sk, plug_in_addr  = account.generate_account()
else:
    plug_in_sk = "Cq9JfCTzMp9bSKVNxuF2YsNm0fS9RsshOVbN6I8Av5zbhEH2I1Qd8UN6UhgOfD1REDY9/pNjPy+D++ib2xTAAg=="
    plug_in_addr = "3OCED5RDKQO7CQ32KIMA47B5KEIDMPP6SNRT6L4D7PUJXWYUYABBZG56JE"
    #******************** Plug_IN_Signer Public Signature   ****************************#
public_key, secret_key = nacl.bindings.crypto_sign_seed_keypair(base64.b64decode(plug_in_sk)[: constants.key_len_bytes])
message = constants.logic_prefix + program
raw_signed = nacl.bindings.crypto_sign(message, secret_key)
crypto_sign_BYTES = nacl.bindings.crypto_sign_BYTES
signature = nacl.encoding.RawEncoder.encode(raw_signed[:crypto_sign_BYTES])
plug_in_public_sig = base64.b64encode(signature).decode()

owner_vault_msig = transaction.Multisig(1,1,[owner_addr, plug_in_addr])


rekey_info = indexer.search_transactions_by_address(plug_in_addr, rekey_to=True)["transactions"]

if (int(client.account_info(owner_vault_msig.address())["amount"]) == 0):
       #********************* MSIG VAULT  Generation **************************************#
       #********************* MSIG VAUL Funding      **************************************#
    note_field = json.dumps({
    "pk": plug_in_addr,
    "sk": plug_in_public_sig,
    "lsig": lsig.address()
    })
    ptxn_vault = transaction.PaymentTxn(
        owner_addr, sp, owner_vault_msig.address(), int(1e6), note=f"arc63:j{note_field}"
    )
    
        #******************** Fund Signer before rekey ADDRESS ****************************# 

    ptxn_signer = transaction.PaymentTxn(
        owner_addr, sp, plug_in_addr, int(1e5 + 1e3)
    )
        #******************** REKEY PLUG_IN TO 0 ADDRESS       ****************************#
         
    zero_msig = transaction.Multisig(1,1,[constants.ZERO_ADDRESS])
    rekey_txn = transaction.PaymentTxn(
        plug_in_addr, sp, plug_in_addr, 0, rekey_to=zero_msig.address()
    )
    transaction.assign_group_id([ptxn_vault, ptxn_signer, rekey_txn])
    
    signed_ptxn_vault =  ptxn_vault.sign(owner_sk)
    signed_ptxn_signer = ptxn_signer.sign(owner_sk)

    signed_rekey = rekey_txn.sign(plug_in_sk)

    signed_group = [signed_ptxn_vault, signed_ptxn_signer, signed_rekey]
    print(signed_group)
    txid = client.send_transactions(signed_group)
    result: Dict[str, Any] = transaction.wait_for_confirmation(
        client, txid, 4
    )
    print(f"txID: {txid} confirmed in round: {result.get('confirmed-round', 0)}")

print("owner vault addresses : ", owner_vault_msig.address()) #NUVYGSZCMMH65PGYGPRB63JNZMMIKT6ROK5HNM3BKVKLSNL77FTB7DKMJU
for i in owner_vault_msig.subsigs:
    print("owner vault address: ", encoding.encode_address(i.public_key), base64.b64encode(i.public_key))


asa_creator_info = client.account_info(asa_creator_addr)
if 'assets' in asa_creator_info and (len(asa_creator_info['assets']) > 0):
    a_id = client.account_info(asa_creator_addr)['assets'][0]['asset-id']
else:
    #******************** asa_creator CREATE ASA ****************************#
    print("asa_creator Create ASA")
    actxn = transaction.AssetConfigTxn( sender=asa_creator_addr, sp=sp, default_frozen=False, unit_name="rug2", asset_name="2 Really Useful Gift", manager=asa_creator_addr, reserve=asa_creator_addr, freeze=asa_creator_addr, clawback=asa_creator_addr, url="https://path/to/my/asset/details", total=10, decimals=0, )
    sactxn = actxn.sign(asa_creator_sk)
    tx_id = client.send_transaction(sactxn)
    print(f"Sent asset create transaction with txid: {tx_id}")
    # Wait for the transaction to be confirmed
    results = transaction.wait_for_confirmation(client, tx_id, 4)
    a_id = results['asset-index']
    print(f"Result confirmed in round: {results['confirmed-round']} ASA ID : {results['asset-index']}")
print(f'asa_creator created 10 ASA: {a_id}')


    #******************** MSIG VAULT OPT IN ****************************#
optin_txn = transaction.AssetTransferTxn(
    sender=owner_vault_msig.address(),
    sp=sp,
    receiver=owner_vault_msig.address(),
    amt=0,
    index=a_id,
)
lsig.lsig.msig = owner_vault_msig
lsig.append_to_multisig(plug_in_sk)

assert (lsig.lsig.msig.subsigs[1].signature ==  base64.b64decode(plug_in_public_sig)) # signature from plug_in_public
lstx = transaction.LogicSigTransaction(optin_txn, lsig)



optin_txid = client.send_transaction(lstx)


print(f"Sent Msig Vault Opt-in with txid: {optin_txid}")
# Wait for the transaction to be confirmed
results = transaction.wait_for_confirmation(client, optin_txid, 4)
print(f"Result confirmed in round: {results['confirmed-round']}")
print("    #********************Opt Out ASA***************************#")
owner_vault_msig = transaction.Multisig(
    1,
    1,
    [owner_addr, plug_in_addr]
)

optout_txn = transaction.AssetCloseOutTxn(
    sender=owner_vault_msig.address(),
    sp=sp,
    receiver=owner_vault_msig.address(),
    index=a_id
)

msig_txn = transaction.MultisigTransaction(optout_txn, owner_vault_msig)
msig_txn.sign(owner_sk)
optout_txn = client.send_transaction(msig_txn)

print(f"Sent Msig Vault Opt-out with txid: {optout_txn}")
# Wait for the transaction to be confirmed
results = transaction.wait_for_confirmation(client, optout_txn, 4)
print(f"Result confirmed in round: {results['confirmed-round']}")
print(f"{owner_vault_msig.address()} Opted - Out {a_id}")
