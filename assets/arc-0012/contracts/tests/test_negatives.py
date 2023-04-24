from fixtures import *
from algosdk.error import AlgodHTTPError
import pytest


class TestNegatives(ARC12TestClass):
    def test_premature_claim_delete(
        self,
        second_verify_axfer,
    ):
        with pytest.raises(AlgodHTTPError) as e:
            self._claim(self.receiver)
        assert e.match("opcodes=pop\n!\nassert\n")

    def test_wrong_claimer(
        self,
        second_verify_axfer,
    ):
        with pytest.raises(AlgodHTTPError) as e:
            self._claim(self.random_acct)
        assert e.match("opcodes=app_global_get\n==\nassert")
