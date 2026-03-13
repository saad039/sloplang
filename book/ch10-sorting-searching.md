# Chapter 10: Sorting and Searching

This chapter presents six classic algorithms implemented entirely in sloplang.
Each implementation is complete and production-tested — these are the exact
functions used in the Phase 8 real-program test suite.

Reading them together reveals a consistent style: index arithmetic uses bracket-
wrapped literals, all dynamic access goes through `$`, in-place mutation is done
with three-variable swaps, and the absence of `else if` is handled by nesting
`if` inside `else` blocks.

---

## 10.1 Bubble Sort

Bubble sort repeatedly scans the array, swapping adjacent elements that are out
of order. After each full pass, the largest unsorted element has "bubbled" to
its correct position at the right end. The inner loop shrinks by one on every
outer iteration, giving O(n²) worst-case time.

```
fn bubble_sort(arr) {
    n = #arr
    i = [0]
    for {
        if i >= n - [1] { break }
        j = [0]
        for {
            if j >= n - [1] - i { break }
            next_j = j + [1]
            if arr$j > arr$next_j {
                tmp = arr$j
                arr$j = arr$next_j
                arr$next_j = tmp
            }
            j = j + [1]
        }
        i = i + [1]
    }
    <- arr
}

fn is_sorted(arr) {
    if #arr <= [1] { <- true }
    i = [0]
    limit = #arr - [1]
    for {
        if i >= limit { break }
        next = i + [1]
        if arr$i > arr$next { <- false }
        i = i + [1]
    }
    <- true
}
```

### Walkthrough

`n = #arr` captures the length once at the start. The outer loop counter `i`
tracks how many elements have been sorted and placed at the right end. The inner
loop bound `n - [1] - i` shrinks with each outer pass: after the first pass, the
rightmost element is in its final position; after the second pass, the last two
are, and so on. There is no need to re-examine already-sorted tail elements.

The swap uses three assignments. Sloplang has no destructuring, so:

```
tmp = arr$j
arr$j = arr$next_j
arr$next_j = tmp
```

`next_j` is precomputed as `j + [1]` because dynamic index-set (`arr$expr =
val`) requires a variable on the right side of `$` — you cannot write
`arr$(j+[1]) = val` inline.

`is_sorted` performs a single linear scan, returning `false` the moment it finds
a pair out of order. Both functions use `for { if condition { break } }` because
sloplang has no `while` keyword.

### Sloplang notes

- `n = #arr` must be stored before use in arithmetic. `#arr - [1] - i` is
  computed from the variable, not the operator applied inline.
- There is no `++i` shorthand. Increment is always written `i = i + [1]`.
- `arr$j` and `arr$next_j` both use the `$` (dynamic) operator because the
  index is held in a variable. Literal numeric indices use `@`: `arr@0`.

---

## 10.2 Merge Sort

Merge sort divides the array in half, recursively sorts each half, then merges
the two sorted halves into a single sorted array. It runs in O(n log n) time and
does not sort in place — each recursive call returns a fresh array.

```
fn merge(left, right) {
    result = []
    li = [0]
    ri = [0]
    for {
        if li >= #left { break }
        if ri >= #right { break }
        lv = left$li
        rv = right$ri
        if lv <= rv {
            result << lv
            li = li + [1]
        } else {
            result << rv
            ri = ri + [1]
        }
    }
    for {
        if li >= #left { break }
        result << left$li
        li = li + [1]
    }
    for {
        if ri >= #right { break }
        result << right$ri
        ri = ri + [1]
    }
    <- result
}

fn merge_sort(arr) {
    len = #arr
    if len <= [1] { <- arr }
    mid = len / [2]
    left = merge_sort(arr::0::mid)
    right = merge_sort(arr::mid::len)
    <- merge(left, right)
}
```

### Walkthrough

`merge_sort` stores the length in `len` before slicing. This is required because
sloplang does not allow `#arr` to appear directly inside a slice postfix
expression — `arr::mid::#arr` would be parsed incorrectly. Storing the value
first gives `arr::mid::len`, which works.

`mid = len / [2]` uses integer division. The left half is `arr::0::mid` (indices
0 through mid-1) and the right half is `arr::mid::len` (indices mid through
end), following sloplang's half-open slice convention `[lo, hi)`.

`merge` maintains two index variables `li` and `ri`. The first loop runs as long
as both halves have remaining elements, always taking the smaller of the two
front elements. When one half is exhausted, the loop exits and two drain loops
copy the remainder of the non-empty half into `result`. Elements are appended
with `<<` (push).

### Sloplang notes

- `arr::lo::hi` is half-open: includes `lo`, excludes `hi`.
- `len = #arr` must precede the slice. Inlining `#arr` in the slice position is
  not supported.
- `<<` spreads the pushed element into the array. Since each element in these
  arrays is a single-element `SlopValue`, `result << lv` appends exactly one
  value.
- Recursion is unbounded. For very large arrays the call stack will grow
  O(log n) frames deep.

---

## 10.3 Quick Sort

Quick sort selects a pivot (here, the first element), then partitions every
other element into three buckets: elements less than the pivot, elements equal
to the pivot, and elements greater than the pivot. The less and greater buckets
are sorted recursively, then all three are concatenated. Average O(n log n),
worst-case O(n²) when the input is already sorted.

```
fn quick_sort(arr) {
    if #arr <= [1] { <- arr }
    pivot = arr@0
    less = []
    equal = []
    greater = []
    for elem in arr {
        if elem < pivot {
            less << elem
        } else {
            if elem == pivot {
                equal << elem
            } else {
                greater << elem
            }
        }
    }
    <- quick_sort(less) ++ equal ++ quick_sort(greater)
}

fn is_sorted(arr) {
    if #arr <= [1] { <- true }
    i = [0]
    limit = #arr - [1]
    for {
        if i >= limit { break }
        next = i + [1]
        if arr$i > arr$next { <- false }
        i = i + [1]
    }
    <- true
}
```

### Walkthrough

The base case returns `arr` immediately when its length is 0 or 1. `pivot =
arr@0` uses the literal numeric index operator `@0` to read the first element.

The `for elem in arr` loop iterates over every element including the pivot
itself (the pivot ends up in `equal`). Each element is routed to one of the
three staging arrays using nested `if`/`else` — there is no `else if` in
sloplang, so the second condition lives inside the `else` block of the first.

The return line `quick_sort(less) ++ equal ++ quick_sort(greater)` is the
assembly step. `++` is array concat, not string join. It produces a single flat
array by appending the elements of `equal` and `quick_sort(greater)` to the
result of `quick_sort(less)`.

### Sloplang notes

- Quick sort is exceptionally concise in sloplang because `++` (concat) and `<<`
  (push) make the partition-and-reassemble pattern natural. No in-place index
  bookkeeping is needed.
- `arr@0` uses literal index `@0`. If the index were held in a variable `i`, it
  would be `arr$i`.
- `++` does not join strings — it concatenates the element sequences of two
  arrays. `["a"] ++ ["b"]` produces `["a", "b"]`.
- Sloplang has no tail-call optimization. The recursion depth is bounded by the
  height of the recursion tree, which degrades to O(n) on sorted input with
  a first-element pivot.

---

## 10.4 Heap Sort

Heap sort works in two phases: first it rearranges the array into a max-heap,
then it repeatedly swaps the root (the largest element) with the last unsorted
element and restores the heap property. O(n log n) time, in-place.

```
fn sift_down(arr, start, end_idx) {
    root = start
    for {
        child = root * [2] + [1]
        if child > end_idx { break }
        swap_idx = root
        if arr$swap_idx < arr$child {
            swap_idx = child
        }
        right_child = child + [1]
        if right_child <= end_idx {
            if arr$swap_idx < arr$right_child {
                swap_idx = right_child
            }
        }
        if swap_idx == root { break }
        tmp = arr$root
        arr$root = arr$swap_idx
        arr$swap_idx = tmp
        root = swap_idx
    }
    <- arr
}

fn heap_sort(arr) {
    n = #arr
    if n <= [1] { <- arr }
    // Build max heap
    start = n / [2] - [1]
    for {
        if start < [0] { break }
        arr = sift_down(arr, start, n - [1])
        start = start - [1]
    }
    // Extract elements
    end_idx = n - [1]
    for {
        if end_idx <= [0] { break }
        tmp = arr@0
        arr@0 = arr$end_idx
        arr$end_idx = tmp
        end_idx = end_idx - [1]
        arr = sift_down(arr, [0], end_idx)
    }
    <- arr
}
```

### Walkthrough

`sift_down(arr, start, end_idx)` enforces the max-heap property on the subtree
rooted at `start`, treating only indices up to `end_idx` as part of the heap.
It computes the left child index as `root * [2] + [1]` and the right child as
`child + [1]`. It then finds the index of the largest value among the root, left
child, and (if it exists) right child. If the root is already the largest,
`swap_idx == root` and the loop breaks. Otherwise the root and largest child are
swapped via a three-variable swap, and the loop continues downward from the
swapped-into position.

`heap_sort` first builds the max-heap by sifting down every non-leaf node,
starting from `n / [2] - [1]` down to index 0. After the heap is built, the
extraction loop runs `n - 1` times: swap `arr@0` (the max) with `arr$end_idx`
(the last heap element), shrink the heap boundary by one, and sift down the new
root.

Because `sift_down` mutates and returns the array, `heap_sort` reassigns `arr`
on each call: `arr = sift_down(arr, ...)`. This is idiomatic sloplang — arrays
are values and mutations are returned, not applied in place behind a reference.

### Sloplang notes

- Index arithmetic uses bracket-wrapped integer literals throughout:
  `root * [2] + [1]`, `child + [1]`, `n / [2] - [1]`.
- `arr@0` uses a literal index; `arr$end_idx` uses a variable index via `$`.
  Both operators work on the same array.
- `sift_down` returns the mutated array and the caller must reassign:
  `arr = sift_down(arr, start, n - [1])`. Forgetting the reassignment would
  silently discard all mutations.

---

## 10.5 Linear Search

Linear search scans each element in sequence until it finds a match, or
exhausts the array. O(n). Works on any array regardless of ordering.

```
fn linear_search(arr, target) {
    i = [0]
    for elem in arr {
        if elem == target {
            <- [i, [0]]
        }
        i = i + [1]
    }
    <- [[-1], [1]]
}
```

### Walkthrough

The function maintains a manual index counter `i` alongside the `for elem in
arr` loop. Sloplang's for-in loop provides the element value directly but does
not expose the index, so `i` is incremented by hand on each iteration.

When `elem == target`, the function returns immediately with a two-element array
`[i, [0]]`. The first element is the index where the target was found; the
second is an error code of `[0]` (success). On failure — the loop finishes
without a match — the function returns `[[-1], [1]]`: index `-1` and error code
`[1]` (not found).

This dual-return convention mirrors the I/O operations in sloplang and allows
callers to use multi-assign syntax:

```
idx, err = linear_search(arr, target)
```

### Sloplang notes

- `for elem in arr` wraps each element in a single-element SlopValue.
  Comparisons with `target` use `==`, which performs deep structural equality.
- The success return `[i, [0]]` bundles two single-element arrays into a
  two-element array. The failure return `[[-1], [1]]` uses `[-1]` (a negative
  one-element array) as the sentinel index.
- Multi-assign `idx, err = linear_search(arr, target)` unpacks the two elements
  via the `UnpackTwo` runtime path (not the builtin dual-return path, which is
  reserved for I/O builtins).

---

## 10.6 Binary Search

Binary search requires a sorted input array and runs in O(log n). It repeatedly
halves the search interval by comparing the target to the middle element,
narrowing in on the target or determining it is absent.

```
fn binary_search(arr, target) {
    lo = [0]
    hi = #arr - [1]
    for {
        if lo > hi { break }
        mid = (lo + hi) / [2]
        val = arr$mid
        if val == target { <- [mid, [0]] }
        if val < target {
            lo = mid + [1]
        } else {
            hi = mid - [1]
        }
    }
    <- [[-1], [1]]
}
```

### Walkthrough

`lo` and `hi` are the current search bounds (both inclusive). On each iteration,
`mid = (lo + hi) / [2]` computes the midpoint using integer division. `val =
arr$mid` reads the element at that position via dynamic index.

If `val == target`, the function returns immediately with the found index and
error code 0. If `val < target`, the target must lie in the upper half, so `lo`
advances to `mid + [1]`. Otherwise the target lies in the lower half and `hi`
retreats to `mid - [1]`. When `lo > hi`, the search interval is empty and the
function returns the not-found sentinel `[[-1], [1]]`.

The binary_search source file also includes a `merge` and `merge_sort`
function — the test harness pre-sorts the array with merge sort before running
binary search against it.

### Sloplang notes

- `mid = (lo + hi) / [2]` — parentheses are necessary. Without them, operator
  precedence would parse this as `lo + (hi / [2])`.
- `arr$mid` uses the `$` operator because `mid` is a variable. `arr@0` would be
  used for the literal index 0.
- The dual-return convention `[mid, [0]]` / `[[-1], [1]]` is identical to
  `linear_search`. Callers use `idx, err = binary_search(arr, target)`.
- Binary search silently produces incorrect results on unsorted input. The
  caller is responsible for sorting first.

---

## Summary

| Algorithm     | Time (avg)  | Time (worst) | In-place | Stable |
|---------------|-------------|--------------|----------|--------|
| Bubble sort   | O(n²)       | O(n²)        | yes      | yes    |
| Merge sort    | O(n log n)  | O(n log n)   | no       | yes    |
| Quick sort    | O(n log n)  | O(n²)        | no*      | no     |
| Heap sort     | O(n log n)  | O(n log n)   | yes      | no     |
| Linear search | O(n)        | O(n)         | —        | —      |
| Binary search | O(log n)    | O(log n)     | —        | —      |

*The sloplang quick sort allocates three partition arrays per call and is
therefore not in-place. A true in-place quick sort would require direct index
mutation without the partition-and-concat pattern.

### Common patterns across all six

**Length capture.** Every algorithm begins with `n = #arr` or `len = #arr`.
This is necessary before arithmetic or slicing because `#arr` cannot appear
inside a slice postfix position.

**Dynamic vs literal index.** `arr$i` when the index is a variable; `arr@0`
when the index is a literal number. Mixing them up is a compile error.

**Three-variable swap.** There is no swap shorthand. All in-place swaps are:

```
tmp = arr$a
arr$a = arr$b
arr$b = tmp
```

**Infinite loop with break.** `for { if cond { break } }` replaces `while`.
The condition check sits at the top of the body.

**Dual-return for search results.** Both search functions return `[index, err]`
as a two-element array and support multi-assign at the call site.
