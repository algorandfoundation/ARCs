import logging
import algokit_utils
from algosdk import (
    mnemonic,
    transaction,
    constants,
    account,
    encoding
)
from algosdk.v2client.algod import AlgodClient
from algosdk.v2client.indexer import IndexerClient
from typing import Dict, Any
import base64
import nacl

logger = logging.getLogger(__name__)


add = [
    {
        "address": "45UO5ZGAAV3VSUFWPY72UITVNSWKLSYJBBALU2O56E32QQYHXHCI5D2PDA",
        "mnemonic": "dumb pencil plastic isolate butter ribbon glide tragic pulse empty grape double glass stadium disorder riot agent donkey city weird shadow bubble ladder absent kidney",
    },
    {
        "address": "6QZBRTHUT4P4D26HBW7NSJJ26P3WV4NWXLBF7AB5TVDVJFXLFLN6RMZQKI",
        "mnemonic": "sense gate people glare window bright betray tiny group subject blast gasp cargo safe play news inhale evolve luggage coil biology wide custom absorb trust",
    },
]

owner_sk, owner_addr = mnemonic.to_private_key(add[0]["mnemonic"]), add[0]["address"]
asa_creator_sk, asa_creator_addr = (
    mnemonic.to_private_key(add[1]["mnemonic"]),
    add[1]["address"],
)


# define deployment behaviour based on supplied app spec
def deploy(
    algod_client: AlgodClient,
    indexer_client: IndexerClient,
    app_spec: algokit_utils.ApplicationSpecification,
    deployer: algokit_utils.Account,
) -> None:
    from smart_contracts.artifacts.smart_app.smart_app_client import SmartAppClient, Composer, AtomicTransactionComposer

    app_client = SmartAppClient(
        algod_client,
        creator=deployer,
        indexer_client=indexer_client,
    )
    dispenser = algokit_utils.get_dispenser_account(algod_client)

    sp = algod_client.suggested_params()
    if (int(algod_client.account_info(owner_addr)["amount"]) <= 1e7):
        ptxn_signer = transaction.PaymentTxn(
            dispenser.address, sp, owner_addr, int(1e7)
        ).sign(dispenser.private_key)
        algod_client.send_transaction(ptxn_signer)

    asa_creator_info = algod_client.account_info(asa_creator_addr)
    if "assets" in asa_creator_info and (len(asa_creator_info["assets"]) > 0):
        a_id = algod_client.account_info(asa_creator_addr)["assets"][0]["asset-id"]
    else:
        # ******************** asa_creator CREATE ASA ****************************#
        ptxn_signer = transaction.PaymentTxn(
            dispenser.address, sp, asa_creator_addr, int(2e5 + 1e3)
        ).sign(dispenser.private_key)
        algod_client.send_transaction(ptxn_signer)
        logger.info("asa_creator Create ASA")
        actxn = transaction.AssetConfigTxn(
            sender=asa_creator_addr,
            sp=sp,
            default_frozen=False,
            unit_name="rug2",
            asset_name="2 Really Useful Gift",
            manager=asa_creator_addr,
            reserve=asa_creator_addr,
            freeze=asa_creator_addr,
            clawback=asa_creator_addr,
            url="https://path/to/my/asset/details",
            total=1,
            decimals=0,
        )
        sactxn = actxn.sign(asa_creator_sk)
        tx_id = algod_client.send_transaction(sactxn)
        logger.info(f"Sent asset create transaction with txid: {tx_id}")
        # Wait for the transaction to be confirmed
        results = transaction.wait_for_confirmation(algod_client, tx_id, 4)
        a_id = results["asset-index"]
        logger.info(
            f"Result confirmed in round: {
                results['confirmed-round']} ASA ID : {results['asset-index']}"
        )
    logger.info(f"asa_creator created 10 ASA: {a_id}")

    app_client.deploy(
        on_schema_break=algokit_utils.OnSchemaBreak.AppendApp,
        on_update=algokit_utils.OnUpdate.AppendApp,
    )

    owner_vault_msig = transaction.Multisig(1, 1, [owner_addr, app_client.app_address])

    teal_program = f"""
    #pragma version 10
    txn TypeEnum
    int appl
    ==
    txn ApplicationID
    int {app_client.app_id}
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

    def compile_program(teal_program):
        compiled_program = algod_client.compile(teal_program)
        return base64.b64decode(compiled_program["result"])

    program = compile_program(teal_program)
    lsig = transaction.LogicSigAccount(program)

    # ******************** Plug_IN_Signer Account Generation *************************#
    if False:
        plug_in_sk, plug_in_addr = account.generate_account()
    else:
        plug_in_sk = "Cq9JfCTzMp9bSKVNxuF2YsNm0fS9RsshOVbN6I8Av5zbhEH2I1Qd8UN6UhgOfD1REDY9/pNjPy+D++ib2xTAAg=="
        plug_in_addr = "3OCED5RDKQO7CQ32KIMA47B5KEIDMPP6SNRT6L4D7PUJXWYUYABBZG56JE"
        # ******************** Plug_IN_Signer Public Signature   *********************#
    _, secret_key = nacl.bindings.crypto_sign_seed_keypair(
        base64.b64decode(plug_in_sk)[: constants.key_len_bytes]
    )
    message = constants.logic_prefix + program
    raw_signed = nacl.bindings.crypto_sign(message, secret_key)
    crypto_sign_BYTES = nacl.bindings.crypto_sign_BYTES
    signature = nacl.encoding.RawEncoder.encode(raw_signed[:crypto_sign_BYTES])
    plug_in_public_sig = base64.b64encode(signature).decode()

    owner_vault_msig = transaction.Multisig(1, 1, [owner_addr, plug_in_addr])
    logger.info(f"Address owner_addr       : {owner_addr}")
    logger.info(f"Address owner_vault_msig : {owner_vault_msig.address()}")
    logger.info(f"Address plug_in_addr     : {plug_in_addr}")
    logger.info(f"App add app_address      : {app_client.app_address}")
    logger.info(f"App ID  app_id           : {app_client.app_id}")
    logger.info(f"Signer  plug_in_public_sig: {plug_in_public_sig}")

    if int(algod_client.account_info(owner_vault_msig.address())["amount"]) == 0:

        # ********************* MSIG VAULT  Generation *******************************#
        # ********************* MSIG App Funding      *******************************#
        ptxn_app = transaction.PaymentTxn(
            owner_addr,
            sp,
            app_client.app_address,
            int(1e6)
        )
        # ********************* MSIG VAUL Funding      *******************************#
        ptxn_vault = transaction.PaymentTxn(
            owner_addr,
            sp,
            owner_vault_msig.address(),
            int(1e6)
        )
        # ******************** Fund Signer before rekey ADDRESS **********************#

        ptxn_signer = transaction.PaymentTxn(
            owner_addr, sp, plug_in_addr, int(1e5 + 1e3)
        )
        # ******************** REKEY PLUG_IN TO 0 ADDRESS       **********************#

        zero_msig = transaction.Multisig(1, 1, [constants.ZERO_ADDRESS])
        rekey_txn = transaction.PaymentTxn(
            plug_in_addr, sp, plug_in_addr, 0, rekey_to=zero_msig.address()
        )
        transaction.assign_group_id([ptxn_app, ptxn_vault, ptxn_signer, rekey_txn])

        ptxn_app
        signed_ptxn_app = ptxn_app.sign(owner_sk)
        signed_ptxn_vault = ptxn_vault.sign(owner_sk)
        signed_ptxn_signer = ptxn_signer.sign(owner_sk)

        signed_rekey = rekey_txn.sign(plug_in_sk)

        signed_group = [signed_ptxn_app, signed_ptxn_vault, signed_ptxn_signer, signed_rekey]
        logger.info(signed_group)
        txid = algod_client.send_transactions(signed_group)
        result: Dict[str, Any] = transaction.wait_for_confirmation(
            algod_client, txid, 4
        )
        logger.info(f"txID: {txid} confirmed in round: {
            result.get('confirmed-round', 0)}")

    asa_creator_info = algod_client.account_info(asa_creator_addr)
    if "assets" in asa_creator_info and (len(asa_creator_info["assets"]) > 0):
        a_id = algod_client.account_info(asa_creator_addr)["assets"][0]["asset-id"]
    else:
        # ******************** asa_creator CREATE ASA ****************************#
        logger.info("asa_creator Create ASA")
        actxn = transaction.AssetConfigTxn(
            sender=asa_creator_addr,
            sp=sp,
            default_frozen=False,
            unit_name="rug2",
            asset_name="2 Really Useful Gift",
            manager=asa_creator_addr,
            reserve=asa_creator_addr,
            freeze=asa_creator_addr,
            clawback=asa_creator_addr,
            url="https://path/to/my/asset/details",
            total=10,
            decimals=0,
        )
        sactxn = actxn.sign(asa_creator_sk)
        tx_id = algod_client.send_transaction(sactxn)
        logger.info(f"Sent asset create transaction with txid: {tx_id}")
        # Wait for the transaction to be confirmed
        results = transaction.wait_for_confirmation(algod_client, tx_id, 4)
        a_id = results["asset-index"]
        logger.info(
            f"Result confirmed in round: {
                results['confirmed-round']} ASA ID : {results['asset-index']}"
        )
    logger.info(f"asa_creator created 10 ASA: {a_id}")

    if (int(algod_client.account_info(app_client.app_address)["amount"]) <= 151300):
        ptxn_signer = transaction.PaymentTxn(
            dispenser.address, sp, app_client.app_address, int(151300)
        ).sign(dispenser.private_key)
        algod_client.send_transaction(ptxn_signer)
    logger.info(" Set public Sig")
    response = app_client.set_public_sig(
        account=owner_addr,
        sig=plug_in_public_sig,
        transaction_parameters=algokit_utils.TransactionParameters(
            boxes=[(app_client.app_id, encoding.decode_address(owner_addr))],
            accounts=[owner_addr]
        )
    )
    logger.info(f"result: {response.return_value}")

    logger.info(" Get public Sig")
    response = app_client.get_public_sig(
        account=owner_addr,
        transaction_parameters=algokit_utils.TransactionParameters(
            boxes=[(app_client.app_id, encoding.decode_address(owner_addr))]
        ),
    )
    logger.info(f"result: {response.return_value}")
    atc = AtomicTransactionComposer()
    client = algokit_utils.ApplicationClient(
        algod_client,
        app_id=app_client.app_id,
        app_spec=app_client.app_spec)

    composer = Composer(client, atc)
    composer.opt_in(
        id=a_id,
        account=app_client.app_address,
        transaction_parameters=algokit_utils.TransactionParameters(
            foreign_assets=[a_id], signer=app_client.signer,
            sender=owner_vault_msig.address()
        ),
    )
    opt_in_txn = composer.atc.txn_list[0].txn

    lsig.lsig.msig = owner_vault_msig
    lsig.lsig.msig.subsigs[1].signature = base64.b64decode(plug_in_public_sig)
    lstx = transaction.LogicSigTransaction(opt_in_txn, lsig)
    optin_txid = algod_client.send_transaction(lstx)

    logger.info(f"Sent Msig Vault Opt-in with txid: {optin_txid}")
    # Wait for the transaction to be confirmed
    results = transaction.wait_for_confirmation(algod_client, optin_txid, 4)
    logger.info(f"Result confirmed in round: {results['confirmed-round']}")
