TODO:

- run with -race flag to check race conditions, add a mu sync.Mutex at the node struct and do a node.mu.Lock() defer node.mu.Unlock() in every operation involving either reading or writing to this.
