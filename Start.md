1] ### DERO Installation, https://github.com/deroproject/wiki/wiki#dero-installation  

        DERO is written in golang and very easy to install both from source and binary.
Installation From Source:  
    Install Golang, minimum Golang 1.10.3 required.
    In go workspace: go get -u github.com/deroproject/derosuite/...
    Check go workspace bin folder for binaries.
    For example on Linux machine following binaries will be created:
        derod-linux-amd64 -> DERO daemon.
        dero-wallet-cli-linux-amd64 -> DERO cmdline wallet.
        explorer-linux-amd64 -> DERO Explorer. Yes, DERO has prebuilt personal explorer also for advance privacy users.

Installation From Binary  
        Download DERO binaries for ARM, INTEL, MAC platform and Windows, Mac, FreeBSD, OpenBSD, Linux etc. operating systems.  
https://github.com/deroproject/derosuite/releases

2] ### Running DERO Daemon  
./derod-linux-amd64 

3] ### Running DERO Wallet (Use local or remote daemon) 
./dero-wallet-cli-linux-amd64 --remote  
https://wallet.dero.io [Web wallet]

4] ### DERO Mining Quickstart
Run miner with wallet address and no. of threads based on your CPU.  
./dero-miner --mining-threads 4 --daemon-rpc-address=http://explorer.dero.io:20206 --wallet-address dERoXHjNHFBabzBCQbBDSqbkLURQyzmPRCLfeFtzRQA3NgVfU4HDbRpZQUKBzq59QU2QLcoAviYQ59FG4bu8T9pZ1woERqciSL  

NOTE: Miners keep your system clock sync with NTP etc.  
Eg on linux machine: ntpdate pool.ntp.org 
For details visit http://wiki.dero.io
