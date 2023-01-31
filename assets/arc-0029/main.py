import base64
import binascii
from algosdk.v2client import algod
from pyteal import compileTeal, Mode

from application_information.application import calc

# This programs assumes Kubo IPFS version 0.17.0
# is installed, in the PATH, and initialized (ipfs init)

# General parameters (assumes sandbox)

algod_token = "a" * 64
algod_address = "http://localhost:4001"

verbose = False

def compile_program(client: algod.AlgodClient, teal_script: str) -> bytes:
    compile_response = client.compile(teal_script)
    return base64.b64decode(compile_response["result"])

def get_interface_from_compiled_program(compiled_program: bytes) -> bytes:
    """
    Retrieve the interfaces from a compiled program.
    Return None if not found
    """

    # Production code SHOULD check the TEAL version of the compiled_program first
    # and be updated to support different encoding for bytecblock for potential future TEAL versions

    prefix = bytes.fromhex("2604056172633239")

    # 0x26 is the bytecblock opcode, 0x04 is the number of []byte strings,
    # 0x05 is the length = 5 bytes, 0x6172633239 is hexadecimal of `arc29`

    count_query = compiled_program.count(prefix)
    if count_query > 1:
        # The program contains several potential arc29 pattern
        # A production code would actually try each of them
        # but this code is more basic and will just fail
        return None

    if count_query == 0:
        # arc29 pattern not present
        return None

    idx = compiled_program.index(prefix) + len(prefix)
    if len(compiled_program) < idx + 7:
        # This was a false positive: there are not enough bytes after the prefix
        return None

    binary_interfaces = compiled_program[idx:idx + 7]
    hex_binary = binary_interfaces.hex()

    i = 0
    list_interfaces = []
    while i < len(hex_binary):
        # We get the length of the encoded decimal integer
        idx = int(hex_binary[i:i+2], 16)*2
        # We get each encoded decimal integer
        value = hex_binary[i+2: i+2+idx]
        list_interfaces.append(int(value, 16))
        i = i + idx + 2 
    return list_interfaces


def int_to_bytes(x: int) -> bytes:
    return x.to_bytes((x.bit_length() + 7) // 8, 'big')

def int_to_hex(x: int) -> bytes:
    return int_to_bytes(x).hex()


def main():
    # Connect to an algod client
    algod_client = algod.AlgodClient(algod_token=algod_token, algod_address=algod_address)

    # Convert the PyTEAL approval program into a TEAL script
    teal_script = compileTeal(calc(), mode=Mode.Application, version=5)

    # Compile the program without information
    compiled_program_wo_information = compile_program(algod_client, teal_script)

    # Compute the information bytes
    information_bytes = [b"arc29".hex(), int_to_hex(1) ,int_to_hex(20) ,int_to_hex(300)]
    # We need to have ['6172633239', '01', '14', '012c'] 
    # We take for the example ARC1 ARC20 & ARC300

    # Append it to the contract
    bytecblock = f'bytecblock {" ".join("0x" + b for b in information_bytes)}'
    assert bytecblock == "bytecblock 0x6172633239 0x01 0x14 0x012c"
    # For number > 256, be sure to have them on 4 digit (eg 300 -> 012c)
    # Append a bytecblock containing the byte string `information_bytes`
    # at the end of provided teal_script
    teal_script_with_information = f'{teal_script}\n{bytecblock}\n'
    if verbose:
        print("New TEAL script:")
        print(teal_script_with_information)
        print("----------------")

    # Compile the program
    compiled_program = compile_program(algod_client, teal_script_with_information)

    # Check that the new program is the old program concatenated with
    # 0x26040561726332390101011402012c
    # where 2604056172633239 is the prefix described above in `get_interface_from_compiled_program`
    # where 0101 is "1" and 01 is the number of bytes
    # where 0114 is "20" and 01 is the number of bytes
    # where 02012c is "300" and 02 is the number of bytes
    assert compiled_program == \
           compiled_program_wo_information +\
           bytes.fromhex("26040561726332390101011402012c")
    assert len(compiled_program) == len(compiled_program_wo_information) + 15
    # Read back interfaces
    list_interfaces = get_interface_from_compiled_program(compiled_program)
    print(list_interfaces)
    assert list_interfaces == [1,20,300]
    print(f"Success: the interfaces are : {list_interfaces}")

if __name__ == "__main__":
    main()
