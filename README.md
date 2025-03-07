# Prolly tree projected onto key-value storage

- [_] node
- [_] tree
- [_] iterator
- [_] encoder

TODO
- currently we read all values on level0, but that's problematic:
  - if we store several trees / generations, all of them will be read
  - need to store root for each generation: `root:generation`
  - then Deserialize starting from that root and read only the relevant level0 nodes
    - problem: the tree currently doesn't store children hashes in the nodes
    - we store hashes and keys (boundaries), but if we store multiple generations, that's not enough
- build trees for the latest N generations of a library
  - see how storage size grows vs naive approach (saving JSONs)
- make a db to index changes between generations
- kv iterator so that I can use it in Diff, compare on KV level without loading the whole tree
- partition storage by adding prefix, e.g. based on generation / timestamp
