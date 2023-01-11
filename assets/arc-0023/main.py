
from pyteal import *
from algosdk.v2client import algod
import base64
from Folder.Application import calc
import os

def create_teal_bytecblock(bytes_slice):
    return f'bytecblock {" ".join(["0x"+b.hex() for b in bytes_slice])}'

def append_information_to_teal(teal_script: str, cid: str):
    cd_bytes = f"arc23 {len(cid)} {cid}".encode()
    bytecblk = create_teal_bytecblock([cd_bytes])
    return f'{teal_script}\n{bytecblk}\n'

def compile_program(client, source_code):
    compile_response = client.compile(source_code)
    return base64.b64decode(compile_response["result"])

if __name__ == "__main__":
    output = os.popen("ipfs add --cid-version=1 Folder/*  -q -w | tail -1")
    cid = output.read()[:-1] #bafybeihdzhfifv46p27ee6weqq6h7ydttzwni7gvdogr3mg6rwscdyzici
    compil_teal_app = compileTeal(calc(), mode=Mode.Application, version=5)
    contract = append_information_to_teal(compil_teal_app, cid)
    algod_client = algod.AlgodClient("a" * 64, "http://localhost:4001")
    print(contract)
    print(compile_program(algod_client, contract))