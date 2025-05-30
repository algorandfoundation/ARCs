from typing import Final

# State Schema
GLOBAL_BYTES: Final[int] = 8
GLOBAL_UINTS: Final[int] = 5
LOCAL_BYTES: Final[int] = 0
LOCAL_UINTS: Final[int] = 2

# Controlled ASA
TOTAL: Final[int] = 2**64 - 1
DECIMALS: Final[int] = 0
DEFAULT_FROZEN: Final[bool] = True
UNIT_NAME: Final[str] = "ARC-20"
NAME: Final[str] = "ARC-20 Smart ASA"
APP_BINDING: Final[bytes] = b"algorand://app/"
