import base64
import subprocess

from algosdk.v2client import algod
from pyteal import compileTeal, Mode

from application_information.application import calc

# This programs assumes Kubo IPFS version 0.17.0
# is installed, in the PATH, and initialized (ipfs init)

# General parameters (assumes sandbox)

algod_token = "a" * 64
algod_address = "http://localhost:4001"

verbose = True


def create_teal_bytecblock(bytes_slice: list[bytes]) -> str:
    """
    Return a TEAL bytecblock with the byte strings given as inputs
    """
    return f'bytecblock {" ".join(["0x" + b.hex() for b in bytes_slice])}'


def append_bytes_information_to_teal(teal_script: str, information_bytes: bytes) -> str:
    """
    Append a bytecblock containing the byte string `information_bytes`
    at the end of provided teal_script
    :param teal_script: string representing a TEAL program
    :param information_bytes: bytes to append as bytecblock at the end of the program
    :return: the new TEAL script
    """
    bytecblock = create_teal_bytecblock([information_bytes])
    return f'{teal_script}\n{bytecblock}\n'


def compile_program(client: algod.AlgodClient, teal_script: str) -> bytes:
    compile_response = client.compile(teal_script)
    return base64.b64decode(compile_response["result"])


def compute_information_bytes(folder: str) -> bytes:
    """
    Compute the (encoded) information byte string corresponding to all the files inside the folder `folder`
    """

    # Use Kubo IPFS command line
    # We don't use --wrap-directory as we are already in a folder
    output = subprocess.run(
        ["ipfs", "add", "--cid-version=1", "--hash=sha2-256", "--recursive", "--quiet",
         "--only-hash", "--ignore=__pycache__", folder],
        capture_output=True
    )

    # The CID is the last non-empty line
    text_cid = output.stdout.decode().strip().split("\n")[-1]
    assert text_cid == "bafybeiavazvdva6uyxqudfsh57jbithx7r7juzvxhrylnhg22aeqau6wte"

    # Check that the text CID is a base32 CID
    if text_cid[0] != "b":
        raise Exception("IPFS returned a non-base32 CID")

    # The CID is a base32 string starting with b, we need to remove this b to get the binary CID
    b32_string = text_cid[1:]
    assert b32_string == "afybeiavazvdva6uyxqudfsh57jbithx7r7juzvxhrylnhg22aeqau6wte"

    # We need to re-pad to make Python happy, pad_length = 6
    pad_length = (8 - (len(b32_string) % 8)) % 8
    binary_cid = base64.b32decode(b32_string.upper() + '=' * pad_length)
    assert binary_cid.hex() == "0170122015066a3a83d4c5e1419647efd2144cf7fc7e9a66b73c70b69cdad0090053d699"

    # Finally compute the bytes information
    information_bytes = b"arc23" + binary_cid
    assert information_bytes.hex() == \
           "61726332330170122015066a3a83d4c5e1419647efd2144cf7fc7e9a66b73c70b69cdad0090053d699"

    return information_bytes


def get_cid_from_compiled_program(compiled_program: bytes) -> bytes:
    """
    Retrieve the binary CID from a compiled program.
    Return None if not found
    """

    # Production code SHOULD check the TEAL version of the compiled_program first
    # and be updated to support different encoding for bytecblock for potential future TEAL versions

    # The binary CID consists in the 36 bytes after the following prefix
    prefix = bytes.fromhex("2601296172633233")
    # 0x26 is the bytecblock opcode, 0x01 is the number of []byte strings,
    # 0x29 is the length = 41 bytes, 0x6172633233 is hexadecimal of `arc23`

    count_query = compiled_program.count(prefix)

    if count_query > 1:
        # The program contains several potential CID
        # A production code would actually try each of them
        # but this code is more basic and will just fail
        return None

    if count_query == 0:
        # CID not present
        return None

    idx = compiled_program.index(prefix) + len(prefix)

    if len(compiled_program) < idx + 36:
        # This was a false positive: there are not enough bytes after the prefix
        return None

    binary_cid = compiled_program[idx:idx + 36]
    return binary_cid


def main():
    # Connect to an algod client
    algod_client = algod.AlgodClient(algod_token=algod_token, algod_address=algod_address)

    # Convert the PyTEAL approval program into a TEAL script
    teal_script = compileTeal(calc(), mode=Mode.Application, version=5)

    # Compile the program without information
    compiled_program_wo_information = compile_program(algod_client, teal_script)

    # Compute the information bytes
    information_bytes = compute_information_bytes("application_information")

    # Append it to the contract
    teal_script_with_information = append_bytes_information_to_teal(teal_script, information_bytes)
    if verbose:
        print("New TEAL script:")
        print(teal_script_with_information)
        print("----------------")

    # Compile the program
    compiled_program = compile_program(algod_client, teal_script_with_information)

    # Check that the new program is the old program concatenated with
    # 0x26012961726332330170122015066a3a83d4c5e1419647efd2144cf7fc7e9a66b73c70b69cdad0090053d699
    # where 0x2601296172633233 is the prefix described above in `get_cid_from_compiled_program`
    # and 0x0170122015066a3a83d4c5e1419647efd2144cf7fc7e9a66b73c70b69cdad0090053d699 is the binary CID
    assert compiled_program == \
           compiled_program_wo_information +\
           bytes.fromhex("26012961726332330170122015066a3a83d4c5e1419647efd2144cf7fc7e9a66b73c70b69cdad0090053d699")
    assert len(compiled_program) == len(compiled_program_wo_information) + 44

    # Now verify that we can read back
    binary_cid = get_cid_from_compiled_program(compiled_program)

    # Compute the associated text_cid
    # (remove padding, prefix by "b", put everything in lower case)
    text_cid = "b" + base64.b32encode(binary_cid).decode("ascii").lower().replace("=", "")
    assert text_cid == "bafybeiavazvdva6uyxqudfsh57jbithx7r7juzvxhrylnhg22aeqau6wte"

    print(f"Success: the CID is: {text_cid}")


if __name__ == "__main__":
    main()
