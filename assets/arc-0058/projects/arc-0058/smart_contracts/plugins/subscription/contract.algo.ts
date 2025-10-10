import { GlobalState, uint64, Global, assert, Account, Bytes, Application, itxn, TemplateVar } from '@algorandfoundation/algorand-typescript';
import { Address, Contract } from '@algorandfoundation/algorand-typescript/arc4';
import { getSpendingAccount, rekeyAddress } from '../../utils/plugins';

/** How frequent this payment can be made */
const FREQUENCY: uint64 = TemplateVar<uint64>('FREQUENCY'); // 1
/** Amount of the payment */
const AMOUNT: uint64 = TemplateVar<uint64>('AMOUNT'); // 100_000

export class SubscriptionPlugin extends Contract {

  lastPayment = GlobalState<uint64>({ initialValue: 0 });

  makePayment(
    walletID: uint64,
    rekeyBack: boolean,
    // eslint-disable-next-line no-unused-vars
    _acctRef: Address
  ): void {
    const wallet = Application(walletID);
    const sender = getSpendingAccount(wallet);

    assert(Global.round - this.lastPayment.value > FREQUENCY);
    this.lastPayment.value = Global.round;
    
    itxn
      .payment({
        sender,
        amount: AMOUNT,
        receiver: Account(Bytes.fromBase32("46XYR7OTRZXISI2TRSBDWPUVQT4ECBWNI7TFWPPS6EKAPJ7W5OBXSNG66M").slice(0, 32)),
        rekeyTo: rekeyAddress(rekeyBack, wallet)
      })
      .submit();
  }
}
