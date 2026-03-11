# Phase 8: Real Programs — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Write 10 self-contained sloplang programs exercising every language feature, each tested at 7 sizes (0, 1, 5, 10, 100, 10000, 100000) with insert/search/delete/update operations.

**Architecture:** Each `.slop` program generates its own data using a deterministic LCG, performs all operations, and **writes results to `.txt` files** (using `.>`). The Go test builds the `slop` CLI, runs it on each `.slop` program (with `SIZE_PLACEHOLDER` replaced per size), reads the output files, and verifies correctness by replicating the same LCG in Go. No `runE2E` — tests use the real `./slop` CLI pipeline.

**Tech Stack:** Sloplang (source), Go (test harness + verification)

---

## Key Design Decisions (from previous session)

### Self-Contained Programs
Each program generates its own data using an LCG pseudo-random number generator. No stdin, no external input files.

### File-Based Output
Programs write results to `.txt` files using `.>` (file write). The Go test reads these files and compares against Go-computed expected values. This exercises the full `./slop` CLI pipeline (transpile → compile → run) and the `.>` file I/O operator.

### Template Substitution
The Go test reads each `.slop` source, replaces `SIZE_PLACEHOLDER` with the actual size (0, 1, 5, ..., 100000), writes the modified source to a temp dir, and runs `./slop` on it.

### Test Sizes
Every program runs at: **0, 1, 5, 10, 100, 10000, 100000**

### Operations
Every program exercises: **insertion, search, deletion, updating**

### Output Strategy
Can't write 1M elements to a file. Each program writes summary data: **length, first 5 elements, last 5 elements, is_sorted check** (for sorting algorithms).

### Idiom Fixes
1. **No `fn main()` wrapper** — conflicts with codegen's auto-generated `main()`
2. **No `++` for string concatenation** — it's array Concat; use separate `|>` / `.>` calls or build strings with `str()`
3. **`??` checks elements, not keys** — for hashmap key existence use `##map ?? key`
4. **All numbers in brackets** — `arr ~@ [0]` not `arr ~@ 0`

---

## File Structure

| File | Purpose |
|------|---------|
| `sloplang/examples/fibonacci.slop` | 8a: Recursive Fibonacci |
| `sloplang/examples/wordcount.slop` | 8b: Word frequency counter |
| `sloplang/examples/array_ops_demo.slop` | 8c: Array manipulation showcase |
| `sloplang/examples/linear_search.slop` | 8d: Linear search |
| `sloplang/examples/binary_search.slop` | 8e: Binary search on sorted array |
| `sloplang/examples/bubble_sort.slop` | 8f: Bubble sort |
| `sloplang/examples/merge_sort.slop` | 8g: Merge sort |
| `sloplang/examples/quick_sort.slop` | 8h: Quick sort |
| `sloplang/examples/heap_sort.slop` | 8i: Heap sort |
| `sloplang/examples/bst.slop` | 8j: Binary search tree (array-based) |
| `sloplang/tests/programs/programs_test.go` | Integration tests for all 10 programs |

---

## Shared Patterns Across Programs

### LCG Random Number Generator
Every program (except fibonacci and wordcount) includes:
```slop
fn next_rand(seed) {
    <- (seed * [1103515245] + [12345]) % [2147483648]
}
```
Deterministic with fixed seed `[42]`, so Go tests replicate identical sequences.

### Build Array
```slop
fn build_array(n, seed) {
    arr = []
    i = [0]
    for {
        if i == n { break }
        seed = next_rand(seed)
        arr << (seed % [1000])
        i = i + [1]
    }
    <- [arr, seed]
}
```

### File Output Helper
Since programs write results to files, each uses a helper that appends summary lines to an output file:
```slop
fn write_summary(label, arr) {
    line = label
    n = #arr
    if n > [5] {
        lo = n - [5]
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " first5="
        .>> "results.txt" str(arr::0::5)
        .>> "results.txt" " last5="
        .>> "results.txt" str(arr::lo::n)
        .>> "results.txt" "\n"
    } else {
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " all="
        .>> "results.txt" str(arr)
        .>> "results.txt" "\n"
    }
    <- arr
}
```
Note: Using `.>>` (append) so multiple operations build up the result file. The first write uses `.>` (truncate) to clear any previous results.

### Is-Sorted Check (sorting programs)
```slop
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

### Go Test Structure
```go
func TestPhase8_BubbleSort(t *testing.T) {
    slopBin := buildSlopCLI(t)  // build ./slop binary once
    sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
    for _, n := range sizes {
        t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
            tmpDir := t.TempDir()
            // Read template, replace SIZE_PLACEHOLDER, write to tmpDir
            source := buildSource(t, "bubble_sort.slop", n)
            srcPath := filepath.Join(tmpDir, "program.slop")
            os.WriteFile(srcPath, []byte(source), 0644)
            // Run: ./slop program.slop
            cmd := exec.Command(slopBin, srcPath)
            cmd.Dir = tmpDir
            out, err := cmd.CombinedOutput()
            if err != nil {
                t.Fatalf("slop failed: %v\n%s", err, out)
            }
            // Read results.txt written by the program
            got, err := os.ReadFile(filepath.Join(tmpDir, "results.txt"))
            if err != nil {
                t.Fatalf("no results.txt: %v", err)
            }
            // Compare against Go-computed expected
            expected := computeExpectedBubbleSort(n)
            if string(got) != expected {
                t.Errorf("mismatch for N=%d\ngot:\n%s\nwant:\n%s", n, string(got), expected)
            }
        })
    }
}
```

Each `computeExpected*` function replays the same LCG in Go, performs the same operations, and builds the expected file content string.

---

## Chunk 1: Infrastructure + Fibonacci + Word Count (Tasks 1–3)

### Task 1: Go Test Harness

**Files:**
- Create: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write the test file skeleton with CLI runner + LCG replication**

The harness:
- `buildSlopCLI(t)` — runs `go build -o slop ./cmd/slop/...` once, returns path to binary
- `buildSource(t, filename, size)` — reads `.slop` template from `examples/`, replaces `SIZE_PLACEHOLDER`
- `lcgNext(seed)` / `lcgBuildArray(n, seed)` — Go-side LCG replicating sloplang's `next_rand`
- `formatSummary(label, arr)` — formats output lines like sloplang would
- `formatIntArray(arr)` — formats `[1, 2, 3]` notation

```go
package programs

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "sort"
    "strings"
    "testing"
)

// projectRoot returns the sloplang module root (two levels up from tests/programs/).
func projectRoot() string {
    _, filename, _, _ := runtime.Caller(0)
    return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// buildSlopCLI compiles the slop CLI binary and returns its path.
func buildSlopCLI(t *testing.T) string {
    t.Helper()
    root := projectRoot()
    binPath := filepath.Join(t.TempDir(), "slop")
    cmd := exec.Command("go", "build", "-o", binPath, "./cmd/slop/...")
    cmd.Dir = root
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("build slop CLI: %v\n%s", err, out)
    }
    return binPath
}

// runSlopProgram runs a .slop source via the slop CLI in a temp dir.
// Returns the working directory where the program ran (for reading output files).
func runSlopProgram(t *testing.T, slopBin, source string) string {
    t.Helper()
    tmpDir := t.TempDir()
    srcPath := filepath.Join(tmpDir, "program.slop")
    if err := os.WriteFile(srcPath, []byte(source), 0644); err != nil {
        t.Fatalf("write source: %v", err)
    }
    cmd := exec.Command(slopBin, srcPath)
    cmd.Dir = tmpDir
    out, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("slop run failed: %v\n%s", err, string(out))
    }
    return tmpDir
}

// buildSource reads a .slop template and replaces SIZE_PLACEHOLDER.
func buildSource(t *testing.T, exampleFile string, size int) string {
    t.Helper()
    root := projectRoot()
    path := filepath.Join(root, "examples", exampleFile)
    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("read example: %v", err)
    }
    return strings.Replace(string(data), "SIZE_PLACEHOLDER", fmt.Sprintf("%d", size), -1)
}

// Go-side LCG replicating sloplang's next_rand
func lcgNext(seed int64) int64 {
    return (seed*1103515245 + 12345) % 2147483648
}

func lcgBuildArray(n int, seed int64) ([]int64, int64) {
    arr := make([]int64, 0, n)
    for i := 0; i < n; i++ {
        seed = lcgNext(seed)
        arr = append(arr, seed%1000)
    }
    return arr, seed
}

func formatSummary(label string, arr []int64) string {
    var sb strings.Builder
    sb.WriteString(label)
    sb.WriteString(": len=")
    sb.WriteString(fmt.Sprintf("[%d]", len(arr)))
    if len(arr) > 5 {
        sb.WriteString(" first5=")
        sb.WriteString(formatIntArray(arr[:5]))
        sb.WriteString(" last5=")
        sb.WriteString(formatIntArray(arr[len(arr)-5:]))
    } else {
        sb.WriteString(" all=")
        sb.WriteString(formatIntArray(arr))
    }
    sb.WriteString("\n")
    return sb.String()
}

func formatIntArray(arr []int64) string {
    if len(arr) == 0 {
        return "[]"
    }
    parts := make([]string, len(arr))
    for i, v := range arr {
        parts[i] = fmt.Sprintf("%d", v)
    }
    return "[" + strings.Join(parts, ", ") + "]"
}
```

- [ ] **Step 2: Verify file compiles**

Run: `cd sloplang && go vet ./tests/programs/...`

- [ ] **Step 3: Commit**

```bash
git add sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): test harness with CLI runner + Go LCG"
```

---

### Task 2: Fibonacci (8a)

**Files:**
- Create: `sloplang/examples/fibonacci.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

Fibonacci uses recursive fib, capped at 20 (too slow beyond that). Uses `SIZE_PLACEHOLDER` like other programs. Sizes: **0, 1, 5, 10, 20**.

- [ ] **Step 1: Write `fibonacci.slop`**

```slop
fn fib(n) {
    if n <= [1] {
        <- n
    }
    <- fib(n - [1]) + fib(n - [2])
}

.> "results.txt" ""
n = [SIZE_PLACEHOLDER]
i = [0]
for {
    if i == n { break }
    .>> "results.txt" str(fib(i))
    .>> "results.txt" "\n"
    i = i + [1]
}
```

- [ ] **Step 2: Write E2E test `TestPhase8_Fibonacci`**

Parameterized by sizes `[0, 1, 5, 10, 20]`. Go test computes fib(0..N-1) and builds expected file content.

- [ ] **Step 3: Run test**

Run: `cd sloplang && go test -run TestPhase8_Fibonacci ./tests/programs/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/fibonacci.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): fibonacci program + E2E test"
```

---

### Task 3: Word Frequency Counter (8b)

**Files:**
- Create: `sloplang/examples/wordcount.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

Word counter uses `SIZE_PLACEHOLDER` like other programs. Sizes: **0, 1, 5, 10, 100, 10000, 100000**. Generates N words by LCG-sampling from a fixed 100-word English vocabulary array, writes them to `wc_input.txt`, then reads back and counts frequencies. Exercises hashmap insert/search/update.

- [ ] **Step 1: Write `wordcount.slop`**

Key idiom fixes:
- No `fn main()` — top-level statements
- `##counts ?? w` for key existence (not `counts ?? w`)
- Uses `.>` / `.>>` to write results
- Fixed 100-word vocabulary as an array literal
- LCG picks `vocab$idx` where `idx = seed % [100]`

The program:
1. Builds a 100-word vocab array
2. LCG generates N word indices, writes space-separated words to `wc_input.txt`
3. Reads `wc_input.txt`, splits, counts frequencies
4. Writes sorted-by-insertion-order results to `results.txt`

The Go test replicates the same LCG + vocab to compute expected frequencies.

- [ ] **Step 2: Write E2E test `TestPhase8_WordCount`**

Parameterized by sizes. Go test replicates LCG, picks same words from same vocab, counts frequencies, builds expected output.

- [ ] **Step 3: Run test**

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/wordcount.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): word frequency counter + E2E test"
```

---

## Chunk 2: Array + Search Programs (Tasks 4–6)

### Task 4: Array Manipulation Demo (8c)

**Files:**
- Create: `sloplang/examples/array_ops_demo.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `array_ops_demo.slop`**

Self-contained. Builds array via LCG. Exercises: insertion (push), search (contains), deletion (remove-at), updating (index-set). Writes all results to `results.txt`.

```slop
fn next_rand(seed) {
    <- (seed * [1103515245] + [12345]) % [2147483648]
}

fn write_summary(label, arr) {
    n = #arr
    if n > [5] {
        lo = n - [5]
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " first5="
        .>> "results.txt" str(arr::0::5)
        .>> "results.txt" " last5="
        .>> "results.txt" str(arr::lo::n)
        .>> "results.txt" "\n"
    } else {
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " all="
        .>> "results.txt" str(arr)
        .>> "results.txt" "\n"
    }
    <- arr
}

// Clear output file
.> "results.txt" ""

// Build array of N elements
n = [SIZE_PLACEHOLDER]
arr = []
seed = [42]
i = [0]
for {
    if i == n { break }
    seed = next_rand(seed)
    arr << (seed % [1000])
    i = i + [1]
}
write_summary("built", arr)

// Insert: push 5 elements
j = [0]
for {
    if j == [5] { break }
    seed = next_rand(seed)
    arr << (seed % [1000])
    j = j + [1]
}
write_summary("after_insert", arr)

// Search: contains check
if #arr > [0] {
    target = arr@0
    .>> "results.txt" "search_first: "
    .>> "results.txt" str(arr ?? target)
    .>> "results.txt" "\n"
}
.>> "results.txt" "search_missing: "
.>> "results.txt" str(arr ?? [9999])
.>> "results.txt" "\n"

// Delete: remove first element if non-empty
if #arr > [0] {
    removed = arr ~@ [0]
    .>> "results.txt" "removed: "
    .>> "results.txt" str(removed)
    .>> "results.txt" "\n"
}
write_summary("after_delete", arr)

// Update: set index 0 if non-empty
if #arr > [0] {
    arr@0 = [777]
    .>> "results.txt" "updated_0: [777]\n"
}
write_summary("after_update", arr)
```

- [ ] **Step 2: Write E2E test `TestPhase8_ArrayOps`**

Parameterized by size. Go test replicates LCG, performs same operations, builds expected `results.txt` content.

- [ ] **Step 3: Run tests**

Run: `cd sloplang && go test -run TestPhase8_ArrayOps ./tests/programs/... -timeout 600s`

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/array_ops_demo.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): array ops demo + E2E tests at 7 sizes"
```

---

### Task 5: Linear Search (8d)

**Files:**
- Create: `sloplang/examples/linear_search.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `linear_search.slop`**

Builds array via LCG. Exercises insert/search/delete/update. All output to `results.txt`.

```slop
fn next_rand(seed) {
    <- (seed * [1103515245] + [12345]) % [2147483648]
}

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

fn write_summary(label, arr) {
    n = #arr
    if n > [5] {
        lo = n - [5]
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " first5="
        .>> "results.txt" str(arr::0::5)
        .>> "results.txt" " last5="
        .>> "results.txt" str(arr::lo::n)
        .>> "results.txt" "\n"
    } else {
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " all="
        .>> "results.txt" str(arr)
        .>> "results.txt" "\n"
    }
    <- arr
}

.> "results.txt" ""

// Build
n = [SIZE_PLACEHOLDER]
arr = []
seed = [42]
i = [0]
for {
    if i == n { break }
    seed = next_rand(seed)
    arr << (seed % [1000])
    i = i + [1]
}
write_summary("built", arr)

// Search existing (first element)
if #arr > [0] {
    target = arr@0
    idx, err = linear_search(arr, target)
    .>> "results.txt" "search_existing: idx="
    .>> "results.txt" str(idx)
    .>> "results.txt" " err="
    .>> "results.txt" str(err)
    .>> "results.txt" "\n"
}

// Search missing
idx2, err2 = linear_search(arr, [9999])
.>> "results.txt" "search_missing: idx="
.>> "results.txt" str(idx2)
.>> "results.txt" " err="
.>> "results.txt" str(err2)
.>> "results.txt" "\n"

// Delete element at index 0
if #arr > [0] {
    arr ~@ [0]
    write_summary("after_delete", arr)
}

// Update index 0
if #arr > [0] {
    arr@0 = [777]
    idx3, err3 = linear_search(arr, [777])
    .>> "results.txt" "search_updated: idx="
    .>> "results.txt" str(idx3)
    .>> "results.txt" " err="
    .>> "results.txt" str(err3)
    .>> "results.txt" "\n"
}
write_summary("final", arr)
```

- [ ] **Step 2: Write E2E test `TestPhase8_LinearSearch`**

- [ ] **Step 3: Run tests**

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/linear_search.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): linear search + E2E tests at 7 sizes"
```

---

### Task 6: Binary Search (8e)

**Files:**
- Create: `sloplang/examples/binary_search.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `binary_search.slop`**

Includes its own merge sort to sort the generated array. Exercises insert/search/delete/update. Output to `results.txt`.

```slop
fn next_rand(seed) {
    <- (seed * [1103515245] + [12345]) % [2147483648]
}

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
    if #arr <= [1] { <- arr }
    mid = #arr / [2]
    left = merge_sort(arr::0::mid)
    right = merge_sort(arr::mid::#arr)
    <- merge(left, right)
}

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

fn write_summary(label, arr) {
    n = #arr
    if n > [5] {
        lo = n - [5]
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " first5="
        .>> "results.txt" str(arr::0::5)
        .>> "results.txt" " last5="
        .>> "results.txt" str(arr::lo::n)
        .>> "results.txt" "\n"
    } else {
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " all="
        .>> "results.txt" str(arr)
        .>> "results.txt" "\n"
    }
    <- arr
}

.> "results.txt" ""

// Build + sort
n = [SIZE_PLACEHOLDER]
arr = []
seed = [42]
i = [0]
for {
    if i == n { break }
    seed = next_rand(seed)
    arr << (seed % [1000])
    i = i + [1]
}
arr = merge_sort(arr)
write_summary("sorted", arr)

// Search existing (middle element)
if #arr > [0] {
    mid_idx = #arr / [2]
    target = arr$mid_idx
    idx, err = binary_search(arr, target)
    .>> "results.txt" "search_existing: idx="
    .>> "results.txt" str(idx)
    .>> "results.txt" " err="
    .>> "results.txt" str(err)
    .>> "results.txt" "\n"
}

// Search missing
idx2, err2 = binary_search(arr, [9999])
.>> "results.txt" "search_missing: idx="
.>> "results.txt" str(idx2)
.>> "results.txt" " err="
.>> "results.txt" str(err2)
.>> "results.txt" "\n"

// Delete middle element
if #arr > [0] {
    del_idx = #arr / [2]
    arr ~@ del_idx
    write_summary("after_delete", arr)
}

// Update: set index 0 to 0, re-sort
if #arr > [0] {
    arr@0 = [0]
    arr = merge_sort(arr)
    idx3, err3 = binary_search(arr, [0])
    .>> "results.txt" "search_updated: idx="
    .>> "results.txt" str(idx3)
    .>> "results.txt" " err="
    .>> "results.txt" str(err3)
    .>> "results.txt" "\n"
}
write_summary("final", arr)
```

- [ ] **Step 2: Write E2E test `TestPhase8_BinarySearch`**

- [ ] **Step 3: Run tests**

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/binary_search.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): binary search + E2E tests at 7 sizes"
```

---

## Chunk 3: Sorting Programs (Tasks 7–10)

All sorting programs follow the same structure: build array via LCG, sort, write summary + is_sorted check, then exercise search/delete/update. All output to `results.txt`.

### Task 7: Bubble Sort (8f)

**Files:**
- Create: `sloplang/examples/bubble_sort.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `bubble_sort.slop`**

In-place sort using `$` dynamic access for swaps.

```slop
fn next_rand(seed) {
    <- (seed * [1103515245] + [12345]) % [2147483648]
}

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

fn write_summary(label, arr) {
    n = #arr
    if n > [5] {
        lo = n - [5]
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " first5="
        .>> "results.txt" str(arr::0::5)
        .>> "results.txt" " last5="
        .>> "results.txt" str(arr::lo::n)
        .>> "results.txt" "\n"
    } else {
        .>> "results.txt" label
        .>> "results.txt" ": len="
        .>> "results.txt" str(#arr)
        .>> "results.txt" " all="
        .>> "results.txt" str(arr)
        .>> "results.txt" "\n"
    }
    <- arr
}

.> "results.txt" ""

// Build
n = [SIZE_PLACEHOLDER]
arr = []
seed = [42]
i = [0]
for {
    if i == n { break }
    seed = next_rand(seed)
    arr << (seed % [1000])
    i = i + [1]
}
write_summary("built", arr)

// Sort
arr = bubble_sort(arr)
write_summary("sorted", arr)
.>> "results.txt" "is_sorted: "
.>> "results.txt" str(is_sorted(arr))
.>> "results.txt" "\n"

// Search
if #arr > [0] {
    target = arr@0
    found = false
    for elem in arr {
        if elem == target {
            found = true
            break
        }
    }
    .>> "results.txt" "search_first: "
    .>> "results.txt" str(found)
    .>> "results.txt" "\n"
}

// Delete + re-sort
if #arr > [0] {
    arr ~@ [0]
    arr = bubble_sort(arr)
    write_summary("after_delete", arr)
    .>> "results.txt" "sorted_after_delete: "
    .>> "results.txt" str(is_sorted(arr))
    .>> "results.txt" "\n"
}

// Update + re-sort
if #arr > [0] {
    arr@0 = [999]
    arr = bubble_sort(arr)
    write_summary("after_update", arr)
    .>> "results.txt" "sorted_after_update: "
    .>> "results.txt" str(is_sorted(arr))
    .>> "results.txt" "\n"
}
```

- [ ] **Step 2: Write E2E test `TestPhase8_BubbleSort`**

- [ ] **Step 3: Run tests** (may be slow for N=100000 — O(n²))

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/bubble_sort.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): bubble sort + E2E tests at 7 sizes"
```

---

### Task 8: Merge Sort (8g)

**Files:**
- Create: `sloplang/examples/merge_sort.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `merge_sort.slop`**

Recursive merge sort. Uses `::` (slice) to split, `<<` (push) to merge.
Key: `result << lv` (not `result << [lv]`) — Push spreads elements.

Same structure as bubble sort but with merge sort. Identical `write_summary`, `is_sorted`, `next_rand`. All output to `results.txt`.

- [ ] **Step 2: Write E2E test `TestPhase8_MergeSort`**

- [ ] **Step 3: Run tests**

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/merge_sort.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): merge sort + E2E tests at 7 sizes"
```

---

### Task 9: Quick Sort (8h)

**Files:**
- Create: `sloplang/examples/quick_sort.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `quick_sort.slop`**

Functional approach: pick pivot, partition into less/equal/greater, recurse, concat with `++`.
Same insert/search/delete/update structure. All output to `results.txt`.

- [ ] **Step 2: Write E2E test `TestPhase8_QuickSort`**

- [ ] **Step 3: Run tests**

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/quick_sort.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): quick sort + E2E tests at 7 sizes"
```

---

### Task 10: Heap Sort (8i)

**Files:**
- Create: `sloplang/examples/heap_sort.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `heap_sort.slop`**

In-place heap sort using `$` dynamic access for index-based mutations.
Same structure. All output to `results.txt`.

- [ ] **Step 2: Write E2E test `TestPhase8_HeapSort`**

- [ ] **Step 3: Run tests**

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/heap_sort.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): heap sort + E2E tests at 7 sizes"
```

---

## Chunk 4: BST + Finalization (Tasks 11–12)

### Task 11: Binary Search Tree (8j)

**Files:**
- Create: `sloplang/examples/bst.slop`
- Modify: `sloplang/tests/programs/programs_test.go`

- [ ] **Step 1: Write `bst.slop`**

Array-based BST using parallel arrays: `vals`, `lefts`, `rights`. Global mutable state. Exercises insert, search, update. All output to `results.txt`.

For large N, recursive `bst_inorder` may hit stack limits — skip 100000 for BST or limit traversal depth.

- [ ] **Step 2: Write E2E test `TestPhase8_BST`**

- [ ] **Step 3: Run tests**

- [ ] **Step 4: Commit**

```bash
git add sloplang/examples/bst.slop sloplang/tests/programs/programs_test.go
git commit -m "feat(phase8): BST + E2E tests at 7 sizes"
```

---

### Task 12: Final Verification + Docs Update

- [ ] **Step 1: Run full test suite**

```bash
cd sloplang && go test ./... -timeout 600s
```

Expected: ALL tests pass.

- [ ] **Step 2: Update `docs/architecture.md`**

Add Phase 8 row: `| 8 | Real Programs | Done |`

- [ ] **Step 3: Update `docs/plans/phase8-real-programs.json`** — flip all `passes` to `true`

- [ ] **Step 4: Commit**

```bash
git add docs/architecture.md docs/plans/phase8-real-programs.json
git commit -m "docs: mark phase 8 complete"
```
