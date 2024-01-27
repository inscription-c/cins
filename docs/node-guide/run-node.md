# Run INS-C Node

[[toc]]

Before runing your own **INS-C** node, it is required to install the binary following the [installation guide](./installation.md).

## Start the server

Once you have installed the Node binary, you can start the server using the following command:

```bash
insc srv \
    --indexdir /path/to/node/data
    --btcnode http://localhost:8332
```

Then, the node will start to sync **INS-C** transactions from the Bitcoin node. For more customization options, please
refer to the `insc srv --help` command.
