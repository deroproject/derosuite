# DERO: Secure, Private CryptoNote DAG Blockchain with Smart Contracts
## ABOUT DERO PROJECT
DERO is decentralized DAG(Directed Acyclic Graph) based blockchain with enhanced reliability, privacy, security, and usability. Consensus algorithm is PoW based on original cryptonight. DERO is industry leading and the first blockchain to have bulletproofs, TLS encrypted Network.

### DERO blockchain has the following salient features:
 - DAG Based: No orphan blocks, No soft-forks.
 - 12 Second Block time.
 - Extremely fast transactions with 2 minutes confirmation time.
 - SSL/TLS P2P Network.
 - CryptoNote: Fully Encrypted Blockchain
 - BulletProofs: Zero Knowledge range-proofs(NIZK).
 - Ring signatures.
 - Fully Auditable Supply.
 - DERO blockchain is written from scratch in Golang. 
 - Developed and maintained by original developers.

### DERO DAG
DERO DAG implementation builds outs a main chain from the DAG network of blocks which refers to main blocks (100% reward) and side blocks (67% rewards). Side blocks contribute to chain PoW security and thus traditional 51% attacks are not possible on DERO network. If DERO network finds another block at the same height, instead of choosing one, DERO include both blocks. Thus, rendering the 51% attack futile.


Traditional Blockchains process blocks as single unit of computation(if a double-spend tx occurs within the block, entire block is rejected). However DERO network accepts such blocks since DERO blockchain considers transaction as a single unit of computation.DERO blocks may contain duplicate or double-spend transactions which are filtered by client protocol and ignored by the network. DERO DAG processes transactions atomically one transaction at a time.


**DERO DAG in action**
![DERO DAG](http://seeds.dero.io/images/Dag1.jpeg)


## Downloads
| Operating System | Download                                 |
| ---------------- | ---------------------------------------- |
| Windows 32       | https://github.com/deroproject/derosuite/releases |
| Windows 64       | https://github.com/deroproject/derosuite/releases |
| Mac 10.8 & Later | https://github.com/deroproject/derosuite/releases |
| Linux 32         | https://github.com/deroproject/derosuite/releases |
| Linux 64         | https://github.com/deroproject/derosuite/releases |
| OpenBSD 64       | https://github.com/deroproject/derosuite/releases |
| FreeBSD 64       | https://github.com/deroproject/derosuite/releases |
| Linux ARM 64     | https://github.com/deroproject/derosuite/releases |
| More Builds      | https://github.com/deroproject/derosuite/releases |


### Build from sources:
In go workspace: **go get -u github.com/deroproject/derosuite/...**

Check bin folder for derod, explorer and wallet  binaries. Use golang-1.10.3 version minimum.

### DERO Quickstart
1. Choose your Operating System and [download Dero software](https://github.com/deroproject/derosuite/releases)
2. Extract the file and change to extracted folder in cmd prompt.
3. Start derod daemon and wait to fully sync till prompt goes green.
4. Open new cmd prompt and run dero-wallet-cli.


For detailed walk through to create/restore Dero wallet pls see: [Create/Restore DERO Wallet in one minute](https://forum.dero.io/t/create-backup-restore-dero-wallet-in-one-minute/110)


**DERO Daemon in action**
![DERO Daemon](http://seeds.dero.io/images/derod1.png)


**DERO Wallet in action**
![DERO Wallet](http://seeds.dero.io/images/wallet1.jpeg)

## Technical
For specific details of current DERO core (daemon) implementation and capabilities, see below:

1. **DAG:** No orphan blocks, No soft-forks.
2. **BulletProofs:** Zero Knowledge range-proofs(NIZK)
3. **Cryptonight Hash:** This is memory-bound algorithm. This provides assurance that all miners are equal. ( No miner has any advantage over common miners).
4. **P2P Protocol:** This layers controls exchange of blocks, transactions and blockchain itself.
5.  **Pederson Commitment:** (Part of ring confidential transactions): Pederson commitment algorithm is a cryptographic primitive that allows user to commit to a chosen value  while keeping it hidden to others. Pederson commitment  is used to hide all amounts without revealing the actual amount. It is a homomorphic commitment scheme.
6.  **Borromean Signature:**  (Part of ring confidential transactions):  Borromean Signatures are used to prove that the commitment has a specific value, without revealing the value itself.
7.  **Additive Homomorphic Encryption:** Additive Homomorphic Encryption is used to prove that sum of encrypted Input transaction amounts is EQUAL to sum of encrypted output amounts. This is based on Homomorphic Pederson commitment scheme.
8.  **Multilayered Linkable Spontaneous Anonymous Group (MLSAG) :** (Part of ring confidential transactions): MLSAG gives DERO untraceability and increases privacy and fungibility. MLSAG is a user controlled parameter ( Mixin) which the user can change to improve his privacy. Mixin of minimal amount is enforced and user cannot disable it.
9.  **Ring Confidential Transactions:** Gives untraceability , privacy and fungibility while making sure that the system is stable and secure.
10.  **Core-Consensus Protocol implemented:** Consensus protocol serves 2 major purpose
   1. Protects the system from adversaries and protects it from forking and tampering.
   2. Next block in the chain is the one and only correct version of truth ( balances).
11.  **Proof-of-Work(PoW) algorithm:**  PoW part of core consensus protocol which is used to cryptographically prove that X amount of work has been done to successfully find a block.
12.  **Difficulty algorithm**: Difficulty algorithm controls the system so as blocks are found roughly at the same speed, irrespective of the number and amount of mining power deployed.
13.  **Serialization/De-serialization of blocks**: Capability to encode/decode/process blocks .
14.  **Serialization/De-serialization of transactions**: Capability to encode/decode/process transactions.
15.  **Transaction validity and verification**: Any transactions flowing within the DERO network are validated,verified.
16.  **Socks proxy:** Socks proxy has been implemented and integrated within the daemon to decrease user identifiability and  improve user anonymity.
17.  **Interactive daemon** can print blocks, txs, even entire blockchain from within the daemon 
18.  **status, diff, print_bc, print_block, print_tx** and several other commands implemented
19.  GO DERO Daemon has both mainnet, testnet support.
20.  **Enhanced Reliability, Privacy, Security, Useability, Portabilty assured.** For discussion on each point how pls visit forum.

## Crypto
Secure and fast crypto is the basic necessity of this project and adequate amount of time has been devoted to develop/study/implement/audit it. Most of the crypto such as ring signatures have been studied by various researchers and are in production by number of projects. As far as the Bulletproofs are considered, since DERO is the first one to implement/deploy, they have been given a more detailed look. First, a bare bones bulletproofs was implemented, then implementations in development were studied (Benedict Bunz,XMR, Dalek Bulletproofs) and thus improving our own implementation.Some new improvements were discovered and implemented (There are number of other improvements which are not explained here). Major improvements are in the Double-Base Double-Scalar Multiplication while validating bulletproofs. A typical bulletproof takes ~15-17 ms to verify. Optimised bulletproofs takes ~1 to ~2 ms(simple bulletproof, no aggregate/batching). Since, in the case of bulletproofs the bases are fixed, we can use precompute table to convert 64*2 Base Scalar multiplication into doublings and additions (NOTE: We do not use Bos-Coster/Pippienger methods). This time can be again easily decreased to .5 ms with some more optimizations.With batching and aggregation, 5000 range-proofs (~2500 TX) can be easily verified on even a laptop. The implementation for bulletproofs is in github.com/deroproject/derosuite/crypto/ringct/bulletproof.go , optimized version is in github.com/deroproject/derosuite/crypto/ringct/bulletproof_ultrafast.go

There are other optimizations such as base-scalar multiplication could be done in less than a microsecond. Some of these optimizations are not yet deployed and may be deployed at a later stage.

###  About Dero Rocket Bulletproofs
 - Dero ultrafast bulletproofs optimization techniques in the form used did not exist anywhere in publicly available cryptography literature at the time of implementation. Please contact for any source/reference to include here if it exists.  Ultrafast optimizations verifies Dero bulletproofs 10 times faster than other/original bulletproof implementations. See: https://github.com/deroproject/derosuite/blob/master/crypto/ringct/bulletproof_ultrafast.go

 - DERO rocket bulletproof implementations are hardened, which protects DERO from certain class of attacks.  

 - DERO rocket bulletproof transactions structures are not compatible with other implementations.

Also there are several optimizations planned in near future in Dero rocket bulletproofs which will lead to several times performance boost. Presently they are under study for bugs, verifications, compatibilty etc.

For technical issues and discussion, please visit https://forum.dero.io
```
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
=H4Qj
-----END PGP PUBLIC KEY BLOCK-----
```




