from pyteal import * #pyteal==0.20.1

def calc():
    return Cond(
        [
            Txn.application_args.length() == Int(0),
            Cond(
                [Txn.application_id() == Int(0), Approve()],
                [Txn.on_completion() == OnComplete.OptIn, Approve()],
                [Txn.on_completion() == OnComplete.CloseOut, Approve()],
                [Txn.on_completion() == OnComplete.DeleteApplication, Reject()],
                [Txn.on_completion() == OnComplete.UpdateApplication, Reject()],
            )
        ],
        [
            Txn.on_completion() == OnComplete.NoOp,
            Cond(
                [
                    Txn.application_args[0] == Bytes('base16', 'fe6bdf69'),  # add(uint64,uint64)uint64
                    Seq([
                        Log(
                            Concat(
                                Bytes('base16', '151f7c75'),
                                Itob(Add(Btoi(Txn.application_args[1]), Btoi(Txn.application_args[2])))
                            )
                        ),
                        Approve()
                    ])
                ],
                [
                    Txn.application_args[0] == Bytes('base16', '766083a7'),  # multiply(uint64,uint64)uint64
                    Seq([
                        Log(
                            Concat(
                                Bytes('base16', '151f7c75'),
                                Itob(Mul(Btoi(Txn.application_args[1]), Btoi(Txn.application_args[2])))
                            )
                        ),
                        Approve()
                    ])
                ],

            )
        ]
    )