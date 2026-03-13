# Chapter 7: Hashmaps

Hashmaps are associative arrays that map string keys to values. In Sloplang, a hashmap is a universal data structure for storing key-value pairs. Like all containers, hashmaps use pure symbolic operators — there are no method calls or dot notation.

## 7.1 Declaration Syntax

A hashmap is declared by specifying a name, a list of string keys in braces, and an array of initial values:

```
person{name, age} = ["alice", [30]]
|> person@name        // alice
|> "\n"
|> str(person@age)    // [30]
|> "\n"
```

The keys are listed inside curly braces; the values are provided as an array. Each value corresponds to the key in the same position. In the example above, `"alice"` is associated with the key `name`, and `[30]` is associated with the key `age`.

Keys are always strings. Values can be any SlopValue — numbers, strings, arrays, other hashmaps, or null.

An empty hashmap is declared with empty braces and an empty array:

```
counts{} = []
```

## 7.2 Reading and Writing Keys

To read a value by its literal key name, use the `@` operator:

```
point{x, y} = [[1], [2]]
|> str(point@x)    // [1]
|> "\n"
```

To write or update a value, assign to the key:

```
point{x, y} = [[1], [2]]
|> str(point@x)    // [1]
|> "\n"
point@x = [10]
|> str(point@x)    // [10]
|> "\n"
```

Assigning to a key that does not exist adds the key. Assigning to an existing key updates its value.

## 7.3 Dynamic Access with `$`

When the key is not known at compile time, use the `$` operator to access by a dynamic key:

```
person{name, age} = ["alice", [30]]
k = "name"
|> person$k        // alice
|> "\n"
```

The `$` operator dispatches at runtime:
- If the variable is a string (e.g., `"name"`), it performs a key lookup.
- If the variable is an int64 (e.g., `[0]`), it accesses by numeric index in insertion order (0 = first key, 1 = second key, etc.).

```
person{name, age} = ["alice", [30]]
i = [0]
|> person$i        // alice  (first key in insertion order)
|> "\n"
```

## 7.4 Keys and Values: `##` and `@@`

The `##` operator returns an array of all keys (as strings) in insertion order:

```
person{name, age} = ["alice", [30]]
keys = ##person
|> str(keys)        // [name, age]
|> "\n"
```

The `@@` operator returns an array of all values in insertion order:

```
person{name, age} = ["alice", [30]]
vals = @@person
|> str(vals)        // [alice, [30]]
|> "\n"
```

To check if a key exists, use the `??` operator on the keys array:

```
person{name, age} = ["alice", [30]]
if ##person ?? "name" {
    |> "has name\n"
}
```

The expression `##map ?? "key"` checks whether `"key"` appears in the keys array. If the key exists, it returns `[1]` (truthy); if not, it returns `[]` (falsy).

## 7.5 Iterating a Hashmap

To iterate over all keys in insertion order, use a `for` loop over the keys array:

```
scores{alice, bob, carol} = [[95], [87], [92]]
for k in ##scores {
    |> k
    |> ": "
    |> str(scores$k)
    |> "\n"
}
```

Output:
```
alice: [95]
bob: [87]
carol: [92]
```

This pattern is the standard way to iterate: get the keys with `##map`, loop over each key, and access its value with `map$k`.

## 7.6 Hashmaps as Structs

A hashmap with a fixed set of keys can represent a record or struct. Here is a simple example:

```
point{x, y, z} = [[1], [2], [3]]
point@x = point@x + [10]
|> str(point@x)    // [11]
|> "\n"
```

Hashmaps are useful as return values from functions. For example:

```
fn make_rect(w, h) {
    rect{width, height, area} = [w, h, w * h]
    <- rect
}

r = make_rect([4], [5])
|> str(r@width)     // [4]
|> "\n"
|> str(r@height)    // [5]
|> "\n"
|> str(r@area)      // [20]
|> "\n"
```

The function `make_rect` returns a hashmap with three keys: `width`, `height`, and `area`. The caller can then access each field using the `@` operator.

## 7.7 Limitations and Workarounds

**Hashmap equality respects insertion order:**

Two hashmaps are equal only if they have the same keys in the same insertion order with equal values. If the keys are in a different order, the hashmaps are not equal:

```
a{x, y} = [[1], [2]]
b{y, x} = [[2], [1]]
|> str(a == b)    // []  — not equal — different insertion order
|> "\n"
```

Even though `a` and `b` contain the same keys and values, they are not equal because the keys were inserted in a different order.

**No key deletion:**

There is no built-in operator to delete a key from a hashmap. To remove a key, rebuild the hashmap without it. Iterate over the keys, skip the unwanted key, and accumulate the rest:

```
fn remove_key(map, unwanted_key) {
    keys = ##map
    vals = @@map
    result{} = []
    i = [0]
    for k in keys {
        if k != unwanted_key {
            v = map$k
            result@k = v
        }
        i = i + [1]
    }
    <- result
}
```

**No nested hashmap literals:**

You cannot nest a hashmap literal inside another hashmap literal. Instead, create each hashmap separately and assign them:

```
inner{val} = [[42]]
outer{data} = [inner]
|> str(outer@data)
```

This creates an `inner` hashmap first, then stores it as the value for the `data` key in the `outer` hashmap.
