### Welcome to the Dero Project
[Explorer](https://explorer.dero.io) [Source](https://github.com/deroproject/derosuite) [Twitter](https://twitter.com/DeroProject) [Discord](https://discord.gg/H95TJDp) [Wiki](https://wiki.dero.io) [Github](https://github.com/deroproject/derosuite) [Stats](http://network.dero.io) [WebWallet](https://wallet.dero.io/) 


### Table of Contents
1. [ABOUT DERO PROJECT](#about-dero-project) 
1. [DERO Crypto](#dero-crypto) 
1. [DERO PORTS](#dero-ports) 
1. [Technical](#technical) 
1. [DERO blockchain salient features](#dero-blockchain-salient-features) 
1. [DERO Innovations](#dero-innovations) 
    1. [Dero DAG](#dero-dag)  
    1. [Client Protocol](#client-protocol)
    1. [Dero Rocket Bulletproofs](#dero-rocket-bulletproofs)
    1. [51% Attack Resistant](#51-attack-resistant)
1. [DERO Mining](#dero-mining) 
1. [DERO Installation](#dero-installation) 
    1. [Installation From Source](#installation-from-source)  
    1. [Installation From Binary](#installation-from-binary)  
1. [Next Step After DERO Installation](#next-step-after-dero-installation) 
    1. [Running DERO Daemon](#running-dero-daemon)  
    1. [Running DERO wallet](#running-dero-wallet)  
        1. [DERO Cmdline Wallet](#dero-cmdline-wallet)  
        1. [DERO WebWallet](#dero-web-wallet)  
        1. [DERO Gui Wallet ](#dero-gui-wallet)  
1. [DERO Explorer](#dero-explorer) 
1. [Proving DERO Transactions](#proving-dero-transactions) 

#### ABOUT DERO PROJECT
&nbsp; &nbsp; &nbsp; &nbsp; [DERO](https://github.com/deroproject/derosuite) is decentralized DAG(Directed Acyclic Graph) based blockchain with enhanced reliability, privacy, security, and usability. Consensus algorithm is PoW based on original cryptonight. DERO is industry leading and the first blockchain to have bulletproofs, TLS encrypted Network.  
&nbsp; &nbsp; &nbsp; &nbsp; DERO is the first crypto project to combine a Proof of Work blockchain with a DAG block structure and wholly anonymous transactions based on [CryptoNote Protocol](https://github.com/deroproject/documentation/blob/master/CryptoNote-WP.pdf). The fully distributed ledger processes transactions with a twelve-second average block time and is secure against majority hashrate attacks. DERO will be the first CryptoNote blockchain to have smart contracts on its native chain without any extra layers or secondary blockchains. At present DERO have Smart Contracts on [testnet](https://github.com/deroproject/documentation/blob/master/testnet/stargate.md).

#### DERO Crypto
&nbsp; &nbsp; &nbsp; &nbsp; Secure and fast crypto is the basic necessity of this project and adequate amount of time has been devoted to develop/study/implement/audit it. Most of the crypto such as ring signatures have been studied by various researchers and are in production by number of projects. As far as the Bulletproofs are considered, since DERO is the first one to implement/deploy, they have been given a more detailed look. First, a bare bones bulletproofs was implemented, then implementations in development were studied (Benedict Bunz,XMR, Dalek Bulletproofs) and thus improving our own implementation.  
&nbsp; &nbsp; &nbsp; &nbsp; Some new improvements were discovered and implemented (There are number of other improvements which are not explained here). Major improvements are in the Double-Base Double-Scalar Multiplication while validating bulletproofs. A typical bulletproof takes ~15-17 ms to verify. Optimised bulletproofs takes ~1 to ~2 ms(simple bulletproof, no aggregate/batching). Since, in the case of bulletproofs the bases are fixed, we can use precompute table to convert 64*2 Base Scalar multiplication into doublings and additions (NOTE: We do not use Bos-Coster/Pippienger methods). This time can be again easily decreased to .5 ms with some more optimizations. With batching and aggregation, 5000 range-proofs (~2500 TX) can be easily verified on even a laptop. The implementation for bulletproofs is in github.com/deroproject/derosuite/crypto/ringct/bulletproof.go , optimized version is in github.com/deroproject/derosuite/crypto/ringct/bulletproof_ultrafast.go

&nbsp; &nbsp; &nbsp; &nbsp; There are other optimizations such as base-scalar multiplication could be done in less than a microsecond. Some of these optimizations are not yet deployed and may be deployed at a later stage.  

#### DERO PORTS
**Mainnet:**  
P2P Default Port: 20202  
RPC Default Port: 20206  
Wallet RPC Default Port: 20209  

**Testnet:**  
P2P Default Port: 30303  
RPC Default Port: 30306  
Wallet RPC Default Port: 30309  

#### Technical
&nbsp; &nbsp; &nbsp; &nbsp; For specific details of current DERO core (daemon) implementation and capabilities, see below:  
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
20.  **Enhanced Reliability, Privacy, Security, Useability, Portabilty assured.**


#### DERO blockchain salient features
 - [DAG Based: No orphan blocks, No soft-forks.](#dero-dag)
 - [51% Attack resistant.](#51-attack-resistant) 
 - 12 Second Block time.
 - Extremely fast transactions with 2 minutes confirmation time.
 - SSL/TLS P2P Network.
 - CryptoNote: Fully Encrypted Blockchain
 - [Dero Fastest Rocket BulletProofs](#dero-rocket-bulletproofs): Zero Knowledge range-proofs(NIZK). 
 - Ring signatures.
 - Fully Auditable Supply.
 - DERO blockchain is written from scratch in Golang. [See all unique blockchains from scratch.](https://twitter.com/cryptic_monk/status/999227961059528704) 
 - Developed and maintained by original developers.

#### DERO Innovations
&nbsp; &nbsp; &nbsp; &nbsp; Following are DERO first and leading innovations.

#### DERO DAG
&nbsp; &nbsp; &nbsp; &nbsp; DERO DAG implementation builds outs a main chain from the DAG network of blocks which refers to main blocks (100% reward) and side blocks (8% rewards).  

![DERO DAG stats.dero.io](https://raw.githubusercontent.com/deroproject/documentation/master/images/Dag1.jpeg)  
*DERO DAG Screenshot* [Live](https://stats.dero.io/)  

![DERO DAG network.dero.io](https://raw.githubusercontent.com/deroproject/documentation/master/images/dagx4.png)  
*DERO DAG Screenshot* [Live](https://network.dero.io/)  

#### Client Protocol
&nbsp; &nbsp; &nbsp; &nbsp; Traditional Blockchains process blocks as single unit of computation(if a double-spend tx occurs within the block, entire block is rejected). However DERO network accepts such blocks since DERO blockchain considers transaction as a single unit of computation.DERO blocks may contain duplicate or double-spend transactions which are filtered by client protocol and ignored by the network. DERO DAG processes transactions atomically one transaction at a time.

####  DERO Rocket Bulletproofs
 - Dero ultrafast bulletproofs optimization techniques in the form used did not exist anywhere in publicly available cryptography literature at the time of implementation. Please contact for any source/reference to include here if it exists.  Ultrafast optimizations verifies Dero bulletproofs 10 times faster than other/original bulletproof implementations. See: https://github.com/deroproject/derosuite/blob/master/crypto/ringct/bulletproof_ultrafast.go

 - DERO rocket bulletproof implementations are hardened, which protects DERO from certain class of attacks.  

 - DERO rocket bulletproof transactions structures are not compatible with other implementations.

&nbsp; &nbsp; &nbsp; &nbsp; Also there are several optimizations planned in near future in Dero rocket bulletproofs which will lead to several times performance boost. Presently they are under study for bugs, verifications, compatibilty etc.

#### 51% Attack Resistant  
&nbsp; &nbsp; &nbsp; &nbsp; DERO DAG implementation builds outs a main chain from the DAG network of blocks which refers to main blocks (100% reward) and side blocks (8% rewards). Side blocks contribute to chain PoW security and thus traditional 51% attacks are not possible on DERO network. If DERO network finds another block at the same height, instead of choosing one, DERO include both blocks. Thus, rendering the 51% attack futile.

#### DERO Mining  
[Mining](https://github.com/deroproject/wiki/wiki/Mining)  

#### DERO Installation
&nbsp; &nbsp; &nbsp; &nbsp; DERO is written in golang and very easy to install both from source and binary. 
#### Installation From Source
1. Install Golang, Golang version 1.12.12 required.  
1. In go workspace: ```go get -u github.com/deroproject/derosuite/...```  
1. Check go workspace bin folder for binaries. 
1. For example on Linux machine following binaries will be created:
    1. derod-linux-amd64 -> DERO daemon.  
    1. dero-wallet-cli-linux-amd64 -> DERO cmdline wallet.  
    1. explorer-linux-amd64 -> DERO Explorer. Yes, DERO has prebuilt personal explorer also for advance privacy users.

#### Installation From Binary
&nbsp; &nbsp; &nbsp; &nbsp; Download [DERO binaries](https://github.com/deroproject/derosuite/releases) for ARM, INTEL, MAC platform and Windows, Mac, FreeBSD, OpenBSD, Linux etc. operating systems.  
Most users required following binaries:  
[Windows 7-10, Server 64bit/amd64 ](https://github.com/deroproject/derosuite/releases/download/v2.1.6-1/dero_windows_amd64_2.1.6-1.alpha.atlantis.07032019.zip)  
[Windows 32bit/x86/386](https://github.com/deroproject/derosuite/releases/download/v2.1.6-1/dero_windows_x86_2.1.6-1.alpha.atlantis.07032019.zip)  
[Linux 64bit/amd64](https://github.com/deroproject/derosuite/releases/download/v2.1.6-1/dero_linux_amd64_2.1.6-1.alpha.atlantis.07032019.tar.gz)  
[Linux 32bit/x86](https://github.com/deroproject/derosuite/releases/download/v2.1.6-1/dero_linux_386_2.1.6-1.alpha.atlantis.07032019.tar.gz)  
[FreeBSD 64bit/amd64](https://github.com/deroproject/derosuite/releases/download/v2.1.6-1/dero_freebsd_amd64_2.1.6-1.alpha.atlantis.07032019.tar.gz)  
[OpenBSD 64bit/amd64](https://github.com/deroproject/derosuite/releases/download/v2.1.6-1/dero_openbsd_amd64_2.1.6-1.alpha.atlantis.07032019.tar.gz)  
[Mac OS](https://github.com/deroproject/derosuite/releases/download/v2.1.6-1/dero_apple_mac_darwin_amd64_2.1.6-1.alpha.atlantis.07032019.tar.gz)  
Contact for support of other hardware and OS.  

#### Next Step After DERO Installation  
&nbsp; &nbsp; &nbsp; &nbsp; Running DERO daemon supports DERO network and shows your support to privacy.  

#### Running DERO Daemon
&nbsp; &nbsp; &nbsp; &nbsp; Run derod.exe or derod-linux-amd64 depending on your operating system. It will start syncing.
1. DERO daemon core cryptography is highly optimized and fast. 
1. Use dedicated machine and SSD for best results.  
1. VPS with 2-4 Cores, 4GB RAM, 60GB disk is recommended.  

![DERO Daemon](https://raw.githubusercontent.com/deroproject/documentation/master/images/derod1.png)  
*DERO Daemon Screenshot*

#### Running DERO Wallet 
Dero cmdline wallet is most reliable and has support of all functions. Cmdline wallet is most secure and reliable.

#### DERO Cmdline Wallet 
&nbsp; &nbsp; &nbsp; &nbsp; DERO cmdline wallet is menu based and very easy to operate. 
Use various options to create, recover, transfer balance etc.  
**NOTE:** DERO cmdline wallet by default connects DERO daemon running on local machine on port 20206.  
If DERO daemon is not running start DERO wallet with --remote option like following:  
**./dero-wallet-cli-linux-amd64 --remote** 
 
![DERO Wallet](https://raw.githubusercontent.com/deroproject/documentation/master/images/wallet-recover2.png)  
*DERO Cmdline Wallet Screenshot*  

#### DERO WEB Wallet 
&nbsp; &nbsp; &nbsp; &nbsp; [Web Wallet](https://wallet.dero.io) runs in your browser, your seeds, keys etc. never leave your browser.

#### DERO GUI Wallet 
&nbsp; &nbsp; &nbsp; &nbsp; [Download DERO GUI Wallet](https://github.com/deroproject/dero_gui_wallet/releases)


#### DERO Explorer
[DERO Explorer](https://explorer.dero.io/) is used to check and confirm transaction  on DERO Network.  
DERO users can run their own explorer on local machine and can [browse](http://127.0.0.1:8080) on local machine port 8080.  
![DERO Explorer](https://github.com/deroproject/documentation/raw/master/images/dero_explorer.png)
*DERO EXPLORER Screenshot*  

#### Proving DERO Transactions
DERO blockchain is completely private, so anyone cannot view, confirm, verify any other's wallet balance or any transactions. 
So to prove any transaction you require *Tx private key* and *receiver address*.  
Tx private key can be obtained using get_tx_key command in dero-wallet-cli.  
Enter the *Tx private key* and *receiver address* in [DERO EXPLORER](https://explorer.dero.io)  
![DERO Explorer Proving Transaction](https://github.com/deroproject/documentation/raw/master/images/explorer-prove-tx.png)
*DERO Explorer Proving Transaction*  








