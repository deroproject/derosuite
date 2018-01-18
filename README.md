# DERO: Secure, Private Blockchain with Smart Contracts

## DERO Project :  Cryptonote + Smart contracts + Lightning fast transactions.

DERO blockchain is a complete new blockchain supporting privacy on cryptonote protocol and smart contracts with lightning fast transactions.
DERO blockchain will be implemented in Golang.

We are pleased to announce status update release 1 of DERO Blockchain.

| Operating System | Download |
| ---------------- | -------- |
| Windows 32       |  http://seeds.dero.io/build/derod-windows-386.exe |
| Windows 64       | http://seeds.dero.io/build/derod-windows-amd64.exe |
| Mac 10.8 & Later | http://seeds.dero.io/build/derod-darwin-amd64 |
| Linux 32         | http://seeds.dero.io/build/derod-linux-386 |
| Linux 64         | http://seeds.dero.io/build/derod-linux-amd64 |
| OpenBSD 64       | http://seeds.dero.io/build/derod-openbsd-amd64 |
| FreeBSD 64       | http://seeds.dero.io/build/derod-freebsd-amd64 |
| Linux ARM 64     | http://seeds.dero.io/build/derod-linux-arm64 |
| More Builds | http://seeds.dero.io/build/ |


**NOTE: DO NOT MIGRATE to this daemon. This is strictly for evaluation.**

![Status-release-update-1](http://seeds.dero.io/images/status-release-1.png)

In the status update release 1, Following parts of cryptonote has been rewritten in golang.

At present, Golang DERO daemon can sync and verify blockchain with the existing DERO network.

For specific details of current DERO core implementation and capabilities, see below:

1.  **Cryptonight Hash:** This is an ASIC resistant, memory-bound algorithm. This provides assurance that all miners are equal. ( No miner has any advantage over common miners).
2.  **Wire protocol (90% completed):** This protocol is used to exchange data between 2 DERO daemon nodes. At this point, Go daemon  can connect to C daemon and vice-versa, sync blockchain and exchange, already possible. Complete interoperability has been achieved. This has 3 sub protocols:
   1. **Levin Protocol:** Bottom most layer, basically message framing.
   2. **P2P Protocol:** Handshake exchange, P2P commands and timed synchronization.
   3. **CryptoNote Protocol:** This layers controls exchange of blocks, transactions and blockchain itself.
3.  **Pederson Commitment:** (Part of ring confidential transactions): Pederson commitment algorithm is a cryptographic primitive that allows user to commit to a chosen value  while keeping it hidden to others. Pederson commitment  is used to hide all amounts without revealing the actual amount. It is a homomorphic commitment scheme.
4.  **Borromean Signature:**  (Part of ring confidential transactions):  Borromean Signatures are used to prove that the commitment has a specific value, without revealing the value itself.
5.  **Additive Homomorphic Encryption:** Additive Homomorphic Encryption is used to prove that sum of encrypted Input transaction amounts is EQUAL to sum of encrypted output amounts. This is based on Homomorphic Pederson commitment scheme.
6.  **Multilayered Linkable Spontaneous Anonymous Group (MLSAG) Work-in-Progress:** (Part of ring confidential transactions): MLSAG gives DERO untraceability and increases privacy and fungibility. MLSAG is a user controlled parameter ( Mixin) which the user can change to improve his privacy. Mixin of minimal amount is enforced and user cannot disable it.
7.  **Core-Consensus Protocol implemented:** Consensus protocol serves 2 major purpose
   1. Protects the system from adversaries and protects it from forking and tampering.
   2. Next block in the chain is the one and only correct  version of truth ( balances)
8.  **Proof-of-Work(PoW) algorithm:**  PoW part of core consensus protocol which is used to cryptographically prove that X amount of work has been done to successfully find a block. To deter use of specialized hardware,  an ASIC resistant, memory bound  cryptonight algorithm is used in DERO project.
9.  **Difficulty algorithm**: Difficulty algorithm controls the system so as blocks are found roughly at the same speed, irrespective of the number and amount of mining power deployed.
10.  **Serialization/De-serialization of blocks**: Capability to encode/decode/process blocks .
11.  **Serialization/De-serialization of transactions**: Capability to encode/decode/process transactions.
12.  **Socks proxy:** Socks proxy has been implemented and integrated within the daemon to decrease user identifiability and  improve user anonymity.
13.  **Interactive daemon** can print blocks, txs, even entire blockchain from within the daemon 
14.  **status, diff, print_bc, print_block, print_tx** and several other commands implemented
15.  GO DERO Daemon has both mainnet, testnet support.
16.  Tree-hash for transactions (based on keccak): This merkle root allows user to verify transactions as needed without adding transaction body to block header.
17.  BOLT Database: DERO Blockchain uses BoltDB for future scalability and low latency reads.
18.  **Enhanced Reliability, Privacy, Security, Useability, Portabilty assured.** For discussion on each point how pls visit forum.



The daemon is only for limited testing and evaluation purposes only.



**NOTE: DO NOT MIGRATE to this daemon. This is strictly for evaluation.**


**NOTE:** Following limitations apply in the current golang version.

- Go Daemon has hard-code seed node address.

- RPC interface not exported

## Build:
In go workspace: **go get -u github.com/deroproject/derosuite/...**
    
Check bin folder for derod binary. Use golang-1.9 version minimum.


For technical issues and discussion, please visit https://forum.dero.io
```
-----BEGIN PGP PUBLIC KEY BLOCK-----

mQSuBFpgP9IRDAC5HFDj9beW/6THlCHMPmjSCUeT0lKtT22uHbTA5CpZFTRvrjF8
l1QFpECuax2LiQUWCg2rl5LZtjE2BL53uNhPagGiUOnMC7w50i3YD/KWoanM9or4
8uNmkYRp7pgnjQKX+NK9TWJmLE94UMUgCUach+WXRG4ito/mc2U2A37Lonokpjb2
hnc3d2wSESg+N0Am91TNSiEo80/JVRcKlttyEHJo6FE1sW5Ll84hW8QeROwYa/kU
N8/jAAVTUc2KzMKknlVlGYRcfNframwCu2xUMlyX5Ghjrr3PmLgQX3qc3k/eTwAr
fHifdvZnsBTquLuOxFHk0xlvdSyoGeX3F0LKAXw1+Y6uyX9v7F4Ap7vEGsuCWfNW
hNIayxIM8iOeb6AOFQycL/GkI0Mv+SCd/8KqdAHT8FWjsJUnOWcYYKvFdN5jcORw
C6OVxf296Sj1Zrti6XVQv63/iaJ9at142AcVwbnvaR2h5IqyXdmzmszmoYVvf7jG
JVsmkwTrRvIgyMcBAOLrwQ7I4JGlL54nKr1mIvGRLZ2lH/2sfM2QHcTgcCQ5DACi
P0wOKlt6UgRQ27Aeh0LtOuFuZReXE8dIpD8f6l+zLS5Kii1SB1yffeSsQbTD6bvt
Ic6h88iUKypNHiFcFNncyad6f4zFYPB1ULXyFoZcpPo3jKjwNW/h//AymgfbqFUa
4dWgdVhdkSKB1BzSMamxKSv9O87Q/Zc2vTcA/0j9RjPsrRIfOCziob+kIcpuylA9
a71R9dJ7r2ivwvdOK2De/VHkEanM8qyPgmxdD03jLsx159fX7B9ItSdxg5i0K9sV
6mgfyGiHETminsW28f36O/WMH0SUnwjdG2eGJsZE2IOS/BqTXHRXQeFVR4b44Ubg
U9h8moORPxc1+/0IFN2Bq4AiLQZ9meCtTmCe3QHOWbKRZ3JydMpoohdU3l96ESXl
hNpD6C+froqQgemID51xe3iPRY947oXjeTD87AHDBcLD/vwE6Ys2Vi9mD5bXwoym
hrXCIh+v823HsJSQiN8QUDFfIMIgbATNemJTXs84EnWwBGLozvmuUvpVWXZSstcL
/ROivKTKRkTYqVZ+sX/yXzQM5Rp2LPF13JDeeATwrgTR9j8LSiycOOFcp3n+ndvy
tNg+GQAKYC5NZWL/OrrqRuFmjWkZu0234qZIFd0/oUQ5tqDGwy84L9f6PGPvshTR
yT6B4FpOqvPt10OQFfpD/h9ocFguNBw0AELjXUHk89bnBTU5cKGLkb1iOnGwtAgJ
mV6MJRjS/TKL6Ne2ddiv46fXlY05zJfg0ZHehe49BIZXQK8/9h5YJGmtcUZP19+6
xPTF5zXWs0k3yzoTGP2iCW/Ksf6b0t0fIIASGFAhQJUmGW1lKAcZTTt425G3NYOc
jmhJaFzcLpTnoqB8RKOTUzWXESXmA86cq4DtyQ2yzeLKBkroRGdpwvpZLH3MeDJ4
EIWSmcKPxm8oafMk6Ni9I4qQLFeSTHcF2qFoBMLKai1lqLd+NAzQmbXHDw6gOac8
+DBfIcaj0f5AK/0G39dOV+pg29pISt2PWDDhZ/XsjetrqcrnhsqNNRyplmmy0xR0
srQwQ2FwdGFpbiBEZXJvIChodHRwczovL2Rlcm8uaW8pIDxzdXBwb3J0QGRlcm8u
aW8+iJAEExEIADgWIQQPOeQljGU5R3AqgjQIsgNgoDqd6AUCWmA/0gIbAwULCQgH
AgYVCAkKCwIEFgIDAQIeAQIXgAAKCRAIsgNgoDqd6FYnAQChtgDnzVwe28s6WDTK
4bBa60dSZf1T08PCKl3+c3xx1QEA2R9K2CLQ6IsO9NXD5kA/pTQs5AxYc9bLo/eD
CZSe/4u5Aw0EWmA/0hAMALjwoBe35jZ7blE9n5mg6e57H0Bri43dkGsQEQ1fNaDq
7XByD0JAiZ20vrrfDsbXZQc+1SBGGOa38pGi6RKEf/q4krGe7EYx4hihHQuc+hco
PqOs6rN3+hfHerUolKpYlkGOSxO1ZjpvMOPBF1hz0Bj9NoPMWwVb5fdWis2BzKAu
GHFAX5Ls86KKZs19DRejWsdFtytEiqM7bAjUW75o3O24faxtByTa2SVmmkavCFS4
BpjDhIU2d5RqhJRkb9fqBU8MDFrmCQqSraQs/CqmOTYzM7E8wlk1SwylXN6yBFX3
RAwq1koFMw8yRMVzswEy917kTHS4IyM2yfYjbnENmWJuHiYJmgn8Lqw1QA3syIfP
E4qpzGBTBq3YXXOSymsNKZmKH0rK/G0l3p33rIagl5UXfr1LVd5XJRu6BzjKuk+q
uL3zb6d0ZSaT+aQ/Sju3shhWjGdCRVoT1shvBbQeyEU5ZLe5by6sp0FH9As3hRkN
0PDALEkhgQwl5hU8aIkwewADBQv/Xt31aVh+k/l+CwThAt9rMCDf2PQl0FKDH0pd
7Tcg1LgbqM20sF62PeLpRq+9iMe/pD/rNDEq94ANnCoqC5yyZvxganjG2Sxryzwc
jseZeq3t/He8vhiDxs3WwFbJSylzPG3u9xgyGkKDfGA74Iu+ASPOPOEOT4oLjI5E
s/tB7muD8l/lpkWij2BOopiZzieQntn8xW8eCFTocSAjZW52SoI1x/gw3NasILoB
nrTy0yOYlM01ucZOTB/0JKpzidkJg336amZdF4bLkfUPyCTE6kzG0PrLrQSeycr4
jkDfWfuFmRhKD2lDtoWDHqiPfe9IJkcTMnp5XfXAG3V2pAc+Mer1WIYajuHieO8m
oFNCzBc0obe9f+zEIBjoINco4FumxP78UZMzwe+hHrj8nFtju7WbKqGWumYH0L34
47tUoWXkCZs9Ni9DUIBVYWzEobgS7pl/H1HLR36klfAHLut0T9PZgipKRjSx1Ljz
M78wxVhupdDvHDEdKnq9E9lD6018iHgEGBEIACAWIQQPOeQljGU5R3AqgjQIsgNg
oDqd6AUCWmA/0gIbDAAKCRAIsgNgoDqd6LTZAQDESAvVHbtyKTwMmrx88p6Ljmtp
pKxKP0O5AFM7b7INbQEAtE3lAIBUA31x3fjC5L6UyGk/a2ssOWTsJx98YxMcPhs=
=H4Qj
-----END PGP PUBLIC KEY BLOCK-----
```


