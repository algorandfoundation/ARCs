
from pyteal import *
from algosdk.v2client import algod
import base64
from Folder.application import calc
import os, subprocess

def create_teal_bytecblock(bytes_slice):
    return f'bytecblock {" ".join(["0x"+b.hex() for b in bytes_slice])}'

def append_bytes_information_to_teal(teal_script: str, cd_bytes: str):
    bytecblk = create_teal_bytecblock([cd_bytes])
    return f'{teal_script}\n{bytecblk}\n'

def compile_program(client, source_code):
    compile_response = client.compile(source_code)
    return compile_response["result"]

if __name__ == "__main__":
    algod_client = algod.AlgodClient("a" * 64, "http://localhost:4001")
    compil_teal_app = compileTeal(calc(), mode=Mode.Application, version=5)

    output = subprocess.run(["ipfs", "add", "--cid-version=1", "Folder", "-r", "-q", "-w","-n"], capture_output=True)
    cid = output.stdout.decode().split("\n")[-3] # bafybeibm6rgofb2jn6iibyo67vchj6r3s3v3xkho7xkydlhy7inalysxxa
    b32_string = cid[1:] # afybeibm6rgofb2jn6iibyo67vchj6r3s3v3xkho7xkydlhy7inalysxxa
    
    pad_length = (8 - (len(b32_string) % 8)) % 8 # need to re-pad to make Python happy, pad_length = 6
    bytes_data = b"arc23" + base64.b32decode(b32_string.upper() + '=' * pad_length)
    print(base64.b64encode(bytes_data)) # YXJjMjMBcBIgLPRM4odJb5CA4d79RHT6O5bru6ju/dWBrPj6GgXiV7g=
    contract = append_bytes_information_to_teal(compil_teal_app, bytes_data)

    compiled_program_raw = compile_program(algod_client, contract)
    compiled_program = base64.b64decode(compiled_program_raw)

    query = bytes.fromhex("2601296172633233")
    # 26 is the bytecblock opcode, 01 is the number of []byte strings, 29 is the length = 41 bytes, 6172633233 is hexadecimal of `arc23

    idx = compiled_program.index(query) + len(query)
    cid_bytes = base64.b32encode(compiled_program[idx:idx+36])
    print(f"b{cid_bytes.decode().lower()[:-pad_length]}") # bafybeibm6rgofb2jn6iibyo67vchj6r3s3v3xkho7xkydlhy7inalysxxa