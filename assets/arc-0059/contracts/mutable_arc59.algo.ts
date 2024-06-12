import { ARC59 } from './arc59.algo';

class MutableARC59 extends ARC59 {
  updateApplication(): void {
    assert(this.txn.sender === this.app.creator);
  }
}
