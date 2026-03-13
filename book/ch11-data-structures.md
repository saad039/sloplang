# Chapter 11: Data Structures

Sloplang's array operators give you everything you need to build classic data structures. This chapter walks through four practical patterns: stacks, queues, binary search trees, sets, and graphs. Each pattern maps cleanly onto the operators you already know.

---

## 11.1 Stacks and Queues with Arrays

Stacks and queues are both linear collections that differ only in which end you remove from. Sloplang arrays support both operations natively.

### Stack (LIFO)

A stack is last-in, first-out. The most recently pushed item is the first one popped.

- **Push:** `stack << item` — appends `item` to the end of `stack`
- **Pop:** `top = >>stack` — removes and returns the last element

```
stack = []
stack << [10]
stack << [20]
stack << [30]
|> str(>>stack)    // [30]
|> "\n"
|> str(>>stack)    // [20]
|> "\n"
|> str(stack)      // [10]
|> "\n"
```

After two pops, only `[10]` remains in `stack`. The `>>` prefix operator mutates the array in place — it is not a read-only peek.

### Queue (FIFO)

A queue is first-in, first-out. The element that has waited longest is served next.

- **Enqueue:** `queue << item` — appends `item` to the back
- **Dequeue:** `front = queue ~@ [0]` — removes and returns the element at index 0

```
queue = []
queue << [10]
queue << [20]
queue << [30]
front = queue ~@ [0]
|> str(front)     // [10]
|> "\n"
front = queue ~@ [0]
|> str(front)     // [20]
|> "\n"
|> str(queue)     // [30]
|> "\n"
```

`~@` is the remove-at operator. `queue ~@ [0]` removes the element at index `0` and returns it as a new single-element array, shifting all remaining elements left.

**Performance note:** `~@ [0]` on a large array shifts all elements left by one position on every dequeue. For small queues this is fine. For high-throughput queues with thousands of elements, keep a separate index pointer and advance it instead of removing from the front.

---

## 11.2 Binary Search Tree with Parallel Arrays

Sloplang has no pointer types and no struct types. To build a tree, represent it as three parallel arrays: one for values, one for left-child indices, one for right-child indices. An integer index into these arrays is the node reference. This is the same representation used by many embedded and arena-allocated tree implementations in lower-level languages.

### Sentinel value

Leaf nodes have no children. The value `-[1]` serves as the sentinel meaning "no child here." It is the unary negate operator applied to `[1]`, producing the integer `-1`. Any non-negative index is a valid node; `-1` is guaranteed to be invalid.

### Global state

The BST is stored in four global variables:

```
vals = []
lefts = []
rights = []
bst_size = [0]
```

Each call to `bst_insert` appends one element to each array. The node at logical index `i` has value `vals@i`, left child `lefts@i`, and right child `rights@i`. Because these are globals, all three functions access them directly without passing them as arguments.

### `bst_insert`

```
fn bst_insert(v) {
    vals << v
    lefts << [-1]
    rights << [-1]
    new_idx = bst_size
    bst_size = bst_size + [1]
    if new_idx == [0] {
        <- new_idx
    }
    cur = [0]
    for {
        cur_val = vals$cur
        if v < cur_val {
            left = lefts$cur
            if left == [-1] {
                lefts$cur = new_idx
                <- new_idx
            }
            cur = left
        } else {
            right = rights$cur
            if right == [-1] {
                rights$cur = new_idx
                <- new_idx
            }
            cur = right
        }
    }
    <- new_idx
}
```

The function first appends to all three arrays and records the new node's index in `new_idx`. If this is the first node (`new_idx == [0]`), it is the root — return immediately. Otherwise walk down from node `0`, going left when `v` is less than the current node's value and right otherwise. When a `-1` child slot is found, write `new_idx` into it and return.

`vals$cur` uses the `$` dynamic-access operator: `cur` holds an integer, so `$` dispatches to an array index lookup (equivalent to `vals@cur` but using a variable instead of a literal).

### `bst_search`

```
fn bst_search(v) {
    if bst_size == [0] { <- false }
    cur = [0]
    for {
        if cur == [-1] { <- false }
        cur_val = vals$cur
        if v == cur_val { <- true }
        if v < cur_val {
            cur = lefts$cur
        } else {
            cur = rights$cur
        }
    }
    <- false
}
```

Standard BST search. An empty tree returns `false` immediately. Each iteration checks whether the current node is `-1` (fell off the tree), compares the target value, and moves left or right. `true` returns `[1]` (truthy) and `false` returns `[]` (falsy), matching sloplang's boolean semantics.

### `bst_inorder_iter`

```
fn bst_inorder_iter() {
    result = []
    stack = []
    if bst_size == [0] { <- result }
    cur = [0]
    for {
        // Push left spine
        for {
            if cur == [-1] { break }
            stack << cur
            cur = lefts$cur
        }
        if #stack == [0] { break }
        cur = >>stack
        result << vals$cur
        cur = rights$cur
    }
    <- result
}
```

In-order traversal visits nodes in sorted order: left subtree, then current node, then right subtree. A recursive implementation would use the call stack to remember where to resume after visiting a subtree. Here, an explicit `stack` array does the same job.

The outer loop has two phases:

1. **Descend left:** push every node on the left spine onto `stack` until hitting a `-1` child.
2. **Visit and go right:** pop the top node with `>>stack`, append its value to `result`, then move to its right child. On the next outer iteration, the right child's left spine is descended.

When the stack empties and there is no current node to descend into, the traversal is complete.

### Usage

```
bst_insert([5])
bst_insert([3])
bst_insert([7])
bst_insert([1])
bst_insert([4])

inorder = bst_inorder_iter()
|> str(inorder)    // [1, 3, 4, 5, 7]
|> "\n"

|> str(bst_search([4]))    // [1]
|> "\n"
|> str(bst_search([9]))    // []
|> "\n"
```

---

## 11.3 Hashmaps as Sets

A set is a collection that supports fast membership testing and stores each element at most once. In sloplang, a hashmap with unit values (`[1]`) implements a set. The keys are the set members; the values are irrelevant (conventionally `[1]`).

### Key operations

- **Insert (literal key):** `set@key = [1]`
- **Insert (dynamic key in variable):** `set$var = [1]`
- **Membership check:** `##set ?? item` — `##set` is the keys array of the hashmap; `??` tests whether `item` is in that array
- **Iterate members:** `for k in ##set { ... }`

### Simple example

```
set{} = []
set@apple = [1]
set@banana = [1]

if ##set ?? "apple" {
    |> "has apple\n"
}
```

`set{} = []` declares `set` as a hashmap. `##set` produces the array of all keys currently in the map. `?? "apple"` checks whether the string `"apple"` is one of them.

### Iteration example

```
// Track visited nodes in a graph algorithm
visited{} = []

fn visit(node) {
    visited$node = [1]    // dynamic key-set with string key
}

visit("a")
visit("b")
visit("c")

for k in ##visited {
    |> k
    |> "\n"
}
// output:
// a
// b
// c
```

`visited$node` uses `$` for dynamic key assignment: `node` holds a string, so `$` dispatches to hashmap key-set. This is equivalent to `visited@a = [1]` when `node` equals `"a"`, but works for any string value at runtime.

---

## 11.4 Graphs with Adjacency Lists

A graph can be represented as a hashmap where each key is a node name (string) and each value is an array of neighbor names. This is the adjacency-list representation, which is memory-efficient for sparse graphs.

### Graph declaration

```
graph{} = []
graph@a = ["b", "c"]
graph@b = ["d"]
graph@c = ["d"]
graph@d = []
```

`graph@a = ["b", "c"]` stores a multi-element array as the value for key `"a"`. A hashmap value can be any `SlopValue`, including a multi-element array. `graph@d = []` stores an empty array for a node with no outgoing edges.

### Breadth-first search

BFS visits nodes level by level, starting from a source. A queue holds nodes waiting to be visited; a `visited` set prevents revisiting.

```
graph{} = []
graph@a = ["b", "c"]
graph@b = ["d"]
graph@c = ["d"]
graph@d = []

// BFS from "a"
queue = ["a"]
visited{} = []
for {
    if #queue == [0] { break }
    node = queue ~@ [0]
    if ##visited ?? node {
        // already visited
    } else {
        visited$node = [1]
        |> node
        |> "\n"
        neighbors = graph$node
        for n in neighbors {
            queue << n
        }
    }
}
// output:
// a
// b
// c
// d
```

Step by step:

- `queue = ["a"]` initializes the queue with the source node as a string element.
- `#queue == [0]` checks whether the queue is empty; if so, BFS is done.
- `queue ~@ [0]` dequeues the front element.
- `##visited ?? node` checks whether `node` is already in the visited set.
- `visited$node = [1]` marks `node` as visited using dynamic key-set (`node` is a string variable).
- `graph$node` looks up the neighbors array for `node` dynamically.
- `queue << n` pushes each neighbor string onto the back of the queue. `n` is a string coming from the neighbors array, so spreading it with `<<` appends the string to the queue.

The output is `a`, `b`, `c`, `d` in BFS order: `a` is visited first, its neighbors `b` and `c` are enqueued, then `b` is dequeued and its neighbor `d` is enqueued, then `c` is dequeued (its neighbor `d` is enqueued again but will be skipped as already visited), then `d` is visited.

### Depth-first search

For DFS, replace the queue with a stack: initialize with the source, pop from the back with `>>` instead of dequeuing from the front.

```
graph{} = []
graph@a = ["b", "c"]
graph@b = ["d"]
graph@c = ["d"]
graph@d = []

// DFS from "a"
dfs_stack = ["a"]
visited{} = []
for {
    if #dfs_stack == [0] { break }
    node = >>dfs_stack
    if ##visited ?? node {
        // already visited
    } else {
        visited$node = [1]
        |> node
        |> "\n"
        neighbors = graph$node
        for n in neighbors {
            dfs_stack << n
        }
    }
}
// output:
// a
// c
// d
// b
```

The only change from BFS is `node = >>dfs_stack` (pop from back) instead of `node = queue ~@ [0]` (dequeue from front). Because the stack reverses the order of exploration, `c` is visited before `b`.

---

The four patterns in this chapter — stack/queue, BST with parallel arrays, hashmap-as-set, and adjacency-list graph — cover the most common data structure needs. Each one maps directly onto sloplang's core operators without requiring any special syntax. The BST is the most involved: managing three parallel arrays and a sentinel value requires discipline, but the payoff is a fully functional ordered tree built from first principles.
