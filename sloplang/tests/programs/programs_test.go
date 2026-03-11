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

// ──────────────────────────────────────────────
// Infrastructure
// ──────────────────────────────────────────────

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

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

func runSlopProgram(t *testing.T, slopBin, source string) string {
	t.Helper()
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "program.slop")
	if err := os.WriteFile(srcPath, []byte(source), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	cmd := exec.Command(slopBin, srcPath)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "SLOP_MODULE_ROOT="+projectRoot())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("slop run failed: %v\n%s", err, string(out))
	}
	return tmpDir
}

func buildSource(t *testing.T, programFile string, size int) string {
	t.Helper()
	_, callerFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(callerFile)
	path := filepath.Join(dir, programFile)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read program: %v", err)
	}
	return strings.Replace(string(data), "SIZE_PLACEHOLDER", fmt.Sprintf("%d", size), -1)
}

func readResults(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "results.txt"))
	if err != nil {
		t.Fatalf("no results.txt: %v", err)
	}
	return string(data)
}

// compareLineByLine compares got vs expected line-by-line, reporting the first
// mismatched line with context. Works correctly even for 100k+ line outputs.
func compareLineByLine(t *testing.T, label string, got, expected string) {
	t.Helper()
	gotLines := strings.Split(got, "\n")
	expLines := strings.Split(expected, "\n")

	// Trim trailing empty line from split (trailing \n produces an empty element)
	if len(gotLines) > 0 && gotLines[len(gotLines)-1] == "" {
		gotLines = gotLines[:len(gotLines)-1]
	}
	if len(expLines) > 0 && expLines[len(expLines)-1] == "" {
		expLines = expLines[:len(expLines)-1]
	}

	maxLines := len(gotLines)
	if len(expLines) > maxLines {
		maxLines = len(expLines)
	}

	for i := 0; i < maxLines; i++ {
		var g, e string
		if i < len(gotLines) {
			g = gotLines[i]
		} else {
			t.Errorf("%s: got has %d lines, expected has %d lines (missing line %d: %q)",
				label, len(gotLines), len(expLines), i+1, expLines[i])
			return
		}
		if i < len(expLines) {
			e = expLines[i]
		} else {
			t.Errorf("%s: got has %d lines, expected has %d lines (extra line %d: %q)",
				label, len(gotLines), len(expLines), i+1, gotLines[i])
			return
		}
		if g != e {
			t.Errorf("%s: line %d mismatch\n  got:  %q\n  want: %q", label, i+1, g, e)
			return
		}
	}
}

// ──────────────────────────────────────────────
// LCG — mirrors sloplang's next_rand
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Formatting — mirrors sloplang's FormatValue / write_summary
// ──────────────────────────────────────────────

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

// ──────────────────────────────────────────────
// Test: Fibonacci
// ──────────────────────────────────────────────

func goFib(n int) int64 {
	if n <= 1 {
		return int64(n)
	}
	return goFib(n-1) + goFib(n-2)
}

func TestFibonacci(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 20}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "fibonacci.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)

			var sb strings.Builder
			for i := 0; i < n; i++ {
				sb.WriteString(fmt.Sprintf("[%d]", goFib(i)))
				sb.WriteString("\n")
			}
			expected := sb.String()

			compareLineByLine(t, fmt.Sprintf("fibonacci N=%d", n), got, expected)
		})
	}
}

// ──────────────────────────────────────────────
// Test: Word Count
// ──────────────────────────────────────────────

var vocab = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "it",
	"for", "not", "on", "with", "he", "as", "you", "do", "at", "this",
	"but", "his", "by", "from", "they", "we", "say", "her", "she", "or",
	"an", "will", "my", "one", "all", "would", "there", "their", "what", "so",
	"up", "out", "if", "about", "who", "get", "which", "go", "me", "when",
	"make", "can", "like", "time", "no", "just", "him", "know", "take", "people",
	"into", "year", "your", "good", "some", "could", "them", "see", "other", "than",
	"then", "now", "look", "only", "come", "its", "over", "think", "also", "back",
	"after", "use", "two", "how", "our", "work", "first", "well", "way", "even",
	"new", "want", "because", "any", "these", "give", "day", "most", "us", "much",
}

func TestWordCount(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "wordcount.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)

			// Replicate LCG to pick same words
			seed := int64(42)
			words := make([]string, 0, n)
			for i := 0; i < n; i++ {
				seed = lcgNext(seed)
				idx := seed % 100
				words = append(words, vocab[idx])
			}

			// Count frequencies preserving insertion order
			type entry struct {
				word  string
				count int64
			}
			seen := make(map[string]int) // word -> index in entries
			var entries []entry
			for _, w := range words {
				if idx, ok := seen[w]; ok {
					entries[idx].count++
				} else {
					seen[w] = len(entries)
					entries = append(entries, entry{w, 1})
				}
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("total: [%d]\n", n))
			for _, e := range entries {
				sb.WriteString(fmt.Sprintf("%s: [%d]\n", e.word, e.count))
			}
			expected := sb.String()

			compareLineByLine(t, fmt.Sprintf("wordcount N=%d", n), got, expected)
		})
	}
}

// ──────────────────────────────────────────────
// Test: Array Ops Demo
// ──────────────────────────────────────────────

func TestArrayOps(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "array_ops_demo.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedArrayOps(n)
			compareLineByLine(t, fmt.Sprintf("array_ops N=%d", n), got, expected)
		})
	}
}

func computeExpectedArrayOps(n int) string {
	var sb strings.Builder
	arr, seed := lcgBuildArray(n, 42)
	sb.WriteString(formatSummary("built", arr))

	// Insert 5 more
	for i := 0; i < 5; i++ {
		seed = lcgNext(seed)
		arr = append(arr, seed%1000)
	}
	sb.WriteString(formatSummary("after_insert", arr))

	// Search
	if len(arr) > 0 {
		target := arr[0]
		found := false
		for _, v := range arr {
			if v == target {
				found = true
				break
			}
		}
		if found {
			sb.WriteString("search_first: [1]\n")
		} else {
			sb.WriteString("search_first: []\n")
		}
	}
	// Search missing 9999
	sb.WriteString("search_missing: []\n")

	// Delete index 0
	var removed int64
	if len(arr) > 0 {
		removed = arr[0]
		arr = arr[1:]
		sb.WriteString(fmt.Sprintf("removed: [%d]\n", removed))
	}
	sb.WriteString(formatSummary("after_delete", arr))

	// Update index 0
	if len(arr) > 0 {
		arr[0] = 777
		sb.WriteString("updated_0: [777]\n")
	}
	sb.WriteString(formatSummary("after_update", arr))

	return sb.String()
}

// ──────────────────────────────────────────────
// Test: Linear Search
// ──────────────────────────────────────────────

func TestLinearSearch(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "linear_search.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedLinearSearch(n)
			compareLineByLine(t, fmt.Sprintf("linear_search N=%d", n), got, expected)
		})
	}
}

func goLinearSearch(arr []int64, target int64) (string, string) {
	for i, v := range arr {
		if v == target {
			return fmt.Sprintf("[%d]", i), "[0]"
		}
	}
	return "[-1]", "[1]"
}

func computeExpectedLinearSearch(n int) string {
	var sb strings.Builder
	arr, _ := lcgBuildArray(n, 42)
	sb.WriteString(formatSummary("built", arr))

	// Search existing
	if len(arr) > 0 {
		target := arr[0]
		idx, errStr := goLinearSearch(arr, target)
		sb.WriteString(fmt.Sprintf("search_existing: idx=%s err=%s\n", idx, errStr))
	}

	// Search missing
	idx2, err2 := goLinearSearch(arr, 9999)
	sb.WriteString(fmt.Sprintf("search_missing: idx=%s err=%s\n", idx2, err2))

	// Delete index 0
	if len(arr) > 0 {
		arr = arr[1:]
		sb.WriteString(formatSummary("after_delete", arr))
	}

	// Update index 0, search for 777
	if len(arr) > 0 {
		arr[0] = 777
		idx3, err3 := goLinearSearch(arr, 777)
		sb.WriteString(fmt.Sprintf("search_updated: idx=%s err=%s\n", idx3, err3))
	}
	sb.WriteString(formatSummary("final", arr))

	return sb.String()
}

// ──────────────────────────────────────────────
// Test: Binary Search
// ──────────────────────────────────────────────

func TestBinarySearch(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "binary_search.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedBinarySearch(n)
			compareLineByLine(t, fmt.Sprintf("binary_search N=%d", n), got, expected)
		})
	}
}

func goBinarySearch(arr []int64, target int64) (int64, int64) {
	lo, hi := int64(0), int64(len(arr)-1)
	for lo <= hi {
		mid := (lo + hi) / 2
		if arr[mid] == target {
			return mid, 0
		}
		if arr[mid] < target {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return -1, 1
}

func computeExpectedBinarySearch(n int) string {
	var sb strings.Builder
	arr, _ := lcgBuildArray(n, 42)
	sort.Slice(arr, func(i, j int) bool { return arr[i] < arr[j] })
	sb.WriteString(formatSummary("sorted", arr))

	// Search existing (middle element)
	if len(arr) > 0 {
		midIdx := len(arr) / 2
		target := arr[midIdx]
		idx, err := goBinarySearch(arr, target)
		sb.WriteString(fmt.Sprintf("search_existing: idx=[%d] err=[%d]\n", idx, err))
	}

	// Search missing
	idx2, err2 := goBinarySearch(arr, 9999)
	sb.WriteString(fmt.Sprintf("search_missing: idx=[%d] err=[%d]\n", idx2, err2))

	// Delete middle
	if len(arr) > 0 {
		delIdx := len(arr) / 2
		arr = append(arr[:delIdx], arr[delIdx+1:]...)
		sb.WriteString(formatSummary("after_delete", arr))
	}

	// Update index 0 to 0, re-sort
	if len(arr) > 0 {
		arr[0] = 0
		sort.Slice(arr, func(i, j int) bool { return arr[i] < arr[j] })
		idx3, err3 := goBinarySearch(arr, 0)
		sb.WriteString(fmt.Sprintf("search_updated: idx=[%d] err=[%d]\n", idx3, err3))
	}
	sb.WriteString(formatSummary("final", arr))

	return sb.String()
}

// ──────────────────────────────────────────────
// Sorting test helpers
// ──────────────────────────────────────────────

func computeExpectedSortProgram(n int, sortFn func([]int64) []int64) string {
	var sb strings.Builder
	arr, _ := lcgBuildArray(n, 42)
	sb.WriteString(formatSummary("built", arr))

	// Sort
	arr = sortFn(arr)
	sb.WriteString(formatSummary("sorted", arr))
	sb.WriteString("is_sorted: [1]\n")

	// Search first element
	if len(arr) > 0 {
		sb.WriteString("search_first: [1]\n")
	}

	// Delete index 0 + re-sort
	if len(arr) > 0 {
		arr = arr[1:]
		arr = sortFn(arr)
		sb.WriteString(formatSummary("after_delete", arr))
		sb.WriteString("sorted_after_delete: [1]\n")
	}

	// Update index 0 + re-sort
	if len(arr) > 0 {
		arr[0] = 999
		arr = sortFn(arr)
		sb.WriteString(formatSummary("after_update", arr))
		sb.WriteString("sorted_after_update: [1]\n")
	}

	return sb.String()
}

func goStableSort(arr []int64) []int64 {
	sort.SliceStable(arr, func(i, j int) bool { return arr[i] < arr[j] })
	return arr
}

// ──────────────────────────────────────────────
// Test: Bubble Sort
// ──────────────────────────────────────────────

func TestBubbleSort(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100}
	// Skip 10000 and 100000 for bubble sort — O(n²) is too slow
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "bubble_sort.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedSortProgram(n, goStableSort)
			compareLineByLine(t, fmt.Sprintf("bubble_sort N=%d", n), got, expected)
		})
	}
}

// ──────────────────────────────────────────────
// Test: Merge Sort
// ──────────────────────────────────────────────

func TestMergeSort(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "merge_sort.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedSortProgram(n, goStableSort)
			compareLineByLine(t, fmt.Sprintf("merge_sort N=%d", n), got, expected)
		})
	}
}

// ──────────────────────────────────────────────
// Test: Quick Sort
// ──────────────────────────────────────────────

func TestQuickSort(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "quick_sort.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedSortProgram(n, goStableSort)
			compareLineByLine(t, fmt.Sprintf("quick_sort N=%d", n), got, expected)
		})
	}
}

// ──────────────────────────────────────────────
// Test: Heap Sort
// ──────────────────────────────────────────────

func TestHeapSort(t *testing.T) {
	slopBin := buildSlopCLI(t)
	sizes := []int{0, 1, 5, 10, 100, 10000, 100000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "heap_sort.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedSortProgram(n, goStableSort)
			compareLineByLine(t, fmt.Sprintf("heap_sort N=%d", n), got, expected)
		})
	}
}

// ──────────────────────────────────────────────
// Test: BST
// ──────────────────────────────────────────────

func TestBST(t *testing.T) {
	slopBin := buildSlopCLI(t)
	// Skip 100000 — BST with LCG data may be deeply unbalanced
	sizes := []int{0, 1, 5, 10, 100, 10000}
	for _, n := range sizes {
		t.Run(fmt.Sprintf("N=%d", n), func(t *testing.T) {
			source := buildSource(t, "bst.slop", n)
			dir := runSlopProgram(t, slopBin, source)
			got := readResults(t, dir)
			expected := computeExpectedBST(n)
			compareLineByLine(t, fmt.Sprintf("bst N=%d", n), got, expected)
		})
	}
}

// Go-side BST replication using parallel arrays
type goBST struct {
	vals   []int64
	lefts  []int64
	rights []int64
	size   int
}

func newGoBST() *goBST {
	return &goBST{}
}

func (b *goBST) insert(v int64) {
	b.vals = append(b.vals, v)
	b.lefts = append(b.lefts, -1)
	b.rights = append(b.rights, -1)
	newIdx := int64(b.size)
	b.size++
	if newIdx == 0 {
		return
	}
	cur := int64(0)
	for {
		if v < b.vals[cur] {
			if b.lefts[cur] == -1 {
				b.lefts[cur] = newIdx
				return
			}
			cur = b.lefts[cur]
		} else {
			if b.rights[cur] == -1 {
				b.rights[cur] = newIdx
				return
			}
			cur = b.rights[cur]
		}
	}
}

func (b *goBST) inorderIter() []int64 {
	result := []int64{}
	if b.size == 0 {
		return result
	}
	stack := []int64{}
	cur := int64(0)
	for {
		for cur != -1 {
			stack = append(stack, cur)
			cur = b.lefts[cur]
		}
		if len(stack) == 0 {
			break
		}
		cur = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		result = append(result, b.vals[cur])
		cur = b.rights[cur]
	}
	return result
}

func (b *goBST) search(v int64) bool {
	if b.size == 0 {
		return false
	}
	cur := int64(0)
	for cur != -1 {
		if v == b.vals[cur] {
			return true
		}
		if v < b.vals[cur] {
			cur = b.lefts[cur]
		} else {
			cur = b.rights[cur]
		}
	}
	return false
}

func computeExpectedBST(n int) string {
	var sb strings.Builder
	bst := newGoBST()
	seed := int64(42)
	for i := 0; i < n; i++ {
		seed = lcgNext(seed)
		bst.insert(seed % 1000)
	}

	inorder := bst.inorderIter()
	sb.WriteString(formatSummary("inorder", inorder))

	// Search existing
	if n > 0 {
		firstVal := bst.vals[0]
		if bst.search(firstVal) {
			sb.WriteString("search_existing: [1]\n")
		} else {
			sb.WriteString("search_existing: []\n")
		}
	}

	// Search missing
	if bst.search(9999) {
		sb.WriteString("search_missing: [1]\n")
	} else {
		sb.WriteString("search_missing: []\n")
	}

	// Update root
	if n > 0 {
		bst.vals[0] = 500
		sb.WriteString("updated_root: [500]\n")
	}

	// Re-traverse
	inorder2 := bst.inorderIter()
	sb.WriteString(formatSummary("after_update", inorder2))

	return sb.String()
}
