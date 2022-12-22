import algosdk
import base64
from pyteal import *
from beaker import *
from application import  *
from nacl.public import PrivateKey, SealedBox, PublicKey
from nacl.encoding import *

def store_key(app_client, private_key):  
    key = "arc15-nacl-curve25519:"
    pub_key = key + str(private_key.public_key.encode(encoder=HexEncoder))[2:-1]
    app_client.call(Master.set_public_key, public_key=pub_key)
    application_state = app_client.get_application_state()
    return application_state['public_key'][len(key):]

def encrypt_text(nacl_public_key, text):  
    sealed_box = SealedBox(PublicKey(bytes.fromhex((nacl_public_key))))
    encrypted_text = sealed_box.encrypt(text, encoder=Base64Encoder).decode("utf-8")
    return encrypted_text

def abi_encode(abi_type, value):
    record_codec = algosdk.abi.ABIType.from_string(str(abi_type.type_spec()))
    return record_codec.encode(value)

def abi_decode(abi_type, value):
    record_codec = algosdk.abi.ABIType.from_string(str(abi_type.type_spec()))
    return record_codec.decode(value)
     

def demo():
    algodAddress = "http://localhost:4001"
    algodToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

    algodclient = algosdk.algod.AlgodClient(algodToken, algodAddress)

    accts = sandbox.get_accounts()
    acct_sender = accts.pop()
    acct_receiver = accts.pop()
    acct_receiver = accts.pop()
    print(f"Sender {acct_sender.address}")
    print(f"Receiver {acct_receiver.address}")
    algod_client = sandbox.get_algod_client()
    app_client = client.ApplicationClient(
        algod_client, Master(), signer=acct_receiver.signer
    )
    app_client.create()
    app_sender = client.ApplicationClient(
        algod_client, Master(),  app_client.app_id, signer=acct_sender.signer
    )

    #create private key
    private_key = PrivateKey.generate()

    nacl_public_key = store_key(app_client, private_key)
    

    text = b"This is an ARC-15 message"
    encrypted_text = encrypt_text(nacl_public_key, text)
    print(f"Encrypted_text {encrypted_text}")
    #Fund application for using box
    app_client.fund(161000 + (len(encrypted_text) - 64)* 400) 

    app_client.call(
        Master.authorize,
        boxes=[[app_client.app_id, algosdk.encoding.decode_address(acct_sender.address)]],
        address_to_add=acct_sender.address,
        info=""
    )


    #Get round when message will be send
    current_round = algodclient.status().get("lastRound") + 1

    #Abi encode of the tuple (sender, round)
    sender_round = abi_encode(SR(), (acct_sender.address,current_round))

    #Write messages to receiver App
    app_sender.call(
        Master.write, 
        boxes=[[app_client.app_id, sender_round],[app_client.app_id, algosdk.encoding.decode_address(acct_sender.address)]],
        text=encrypted_text, 
        )

    #Read all messages 
    for box in app_client.client.application_boxes(app_client.app_id)["boxes"]:
        name = base64.b64decode(box["name"])
        if name == sender_round:
            contents = app_client.client.application_box_by_name(app_client.app_id, name)
            text_to_decrypt = abi_decode(abi.String(), base64.b64decode(contents["value"]))
            box_key = abi_decode(SR(), name)
            print(f"Current Box: {box_key} {text_to_decrypt}")
            encrypted_from_box = text_to_decrypt
            unseal_box = SealedBox(private_key)
            plaintext = unseal_box.decrypt(encrypted_from_box, encoder=Base64Encoder)
            print(f"Decrypted Text: {plaintext.decode('utf-8')}")

    #Delete the message
    app_sender.call(
        Master.remove,
        boxes=[[app_client.app_id, sender_round]],
        address=acct_sender.address,
        round=current_round
        )

if __name__ == "__main__":
    demo()