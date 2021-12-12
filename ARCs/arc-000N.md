Smart Signature Template Validation
-----------------------------------

# Motivation

A Smart Signature (SmartSig) has some advantages compared to a Smart Contract. For example it may use up to 20k ops per invocation (compared to 700 in a smart contract). It may also be used as an account that can opt into an application and obeys pre-defined rules (e.g. doesn't close out unexpectedly). 

If the SmartSig is constant (i.e. not a Template with potentially different variables) the Address of the Sender may be used and compared to some predefined address.

However, if a Template is used (e.g. in the case of a unique SmartSig per Account), there is no straight forward method to validate that the SmartSig adheres to the logic expected. 

# Solution

In order for a Smart Contract to perform this validation on a Template, two pieces of information should be known ahead of time:

1. The positions of the Template Variables within the compiled contract
2. The bytecode of the compiled SmartSig with all Template Variables set to the 0 value (0 for uint64, "" for bytes)

With this information, the Smart Contract performing the validation may accept a number of arguments corressponding to the order and type of the Template Variables in the SmartSig. It may then encode the variables properly (uint64=>varuint) and insert them in the bytecode at its start position.

> Note: For each variable inserted, the position in the bytcode will need to be increased by the length of the variable added in bytes.

With the Populated bytecode, the contract may call Sha512_265("Program"||populated_bytecode) to produce the expected Address. 

If the Sender of the transaction matches the address produced, the contract is considered valid we know it is an instance of the Template. 


# Prerequisites

During assembly of a SmartSig Template, the assembler should note the bytecode position and type of any Template Variables (strings with the prefix `TMPL_`). These Template Variables should be set to the zero value for their type. 

The output of the assembly step should be 2 artifacts:

1. A JSON file containing the list of template variables 

ex:
```js
{
    "template_variables":[
        {
            "name":"TMPL_ASSET_ID",
            "type": 1,
            "position":12
        },
        {
            "name":"TMPL_CREATOR_ADDR",
            "type": 2,
            "position":25
        },
    ],
    ...
}
```

2. A binary file consisting of the bytecode of the SmartSig Template


These two artifact should contain enough information to produce an ABI style method that accepts as arguments all the template variables specified in the contract and produce the address of the expected SmartSig.

`populate(tmpl_var_1_type, tmpl_var_2_type, ...)address`

In the body of the method, the steps described above should be followed to populate the contract and produce the address. 

The callee may be an off chain client, an external contract, or itself.


