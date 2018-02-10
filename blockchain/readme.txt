The blockchain is stored as follows, assuming genesis block is stored

each block has a single child to store main chain
each block can have multiple children to store numerous blocks as alternative


each block stores the following extra info
  1) height
  2) single parent
  3) cumulative difficulty
  4) size of block (  size of all transactions included within block + size of block )
  5) timestamp
  6) cumulative coins
  7) block blob itself
  8) child
  9) multiple children ( alternative or orphan blocks )

storing block mean storing all attributes except child/children
connecting a block means setting child/children attributes properly
  
  
imagine the chain as a doubly linked list, with traversal possible using blockids as ptrs

the psedo code is as follows

1) verify incoming blockor semantic errors
2) verify block pow
3) verify each and every transaction for correctnes
4) store the transactions
5) store the block, height, cumulative difficulty, coins emission etc
6) check whether block is the new top, 
7) if yes,update main chain
8) if not we have to do chain reorganisation at the top
9) choose the block with higher poW as new top
10) push alt chain txs to mem pool after verification
11) if block is being added somewhere in the middle, find the chain with higher Pow as main chain
12) push the orphan block txs to mempool after verification
