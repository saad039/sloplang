package codegen

import (
	"testing"
)

// ============================================================
// Semantic Hashmap E2E Tests
// Covers hashmap declaration, key access/set, ##/@@, dynamic
// access via $, iteration, functions, and equality.
// ============================================================

// --- Declaration & access (~5 tests) ---

func TestSem_Map_BasicDecl(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> person@name
|> str(person@age)`)
	if got != "bob[30]" {
		t.Fatalf("expected %q, got %q", "bob[30]", got)
	}
}

func TestSem_Map_EmptyMap(t *testing.T) {
	got := runE2E(t, `m{} = []
|> str(##m)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Map_SingleKey(t *testing.T) {
	got := runE2E(t, `m2{x} = [[5]]
|> str(m2@x)`)
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
	}
}

func TestSem_Map_KeyNotFound_Panics(t *testing.T) {
	runE2EExpectPanic(t, `m3{name} = ["bob"]
x = m3@missing`)
}

func TestSem_Map_StringValue(t *testing.T) {
	got := runE2E(t, `m4{a} = ["hello"]
|> m4@a`)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

// --- Key set (~5 tests) ---

func TestSem_Map_UpdateExisting(t *testing.T) {
	got := runE2E(t, `p1{name, age} = ["bob", [30]]
p1@age = [31]
|> str(p1@age)`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

func TestSem_Map_AddNew(t *testing.T) {
	got := runE2E(t, `p2{name} = ["bob"]
p2@email = "a@b"
|> p2@email`)
	if got != "a@b" {
		t.Fatalf("expected %q, got %q", "a@b", got)
	}
}

func TestSem_Map_SingleUnwraps(t *testing.T) {
	// IndexKeySetStr unwraps single-element: [31] stores raw int64(31)
	// Elements become ["bob", int64(31)]
	// @@p3 -> MapValues -> {Elements: ["bob", int64(31)]}
	// str() -> [bob, 31]
	got := runE2E(t, `p3{name, age} = ["bob", [30]]
p3@age = [31]
|> str(@@p3)`)
	if got != "[bob, 31]" {
		t.Fatalf("expected %q, got %q", "[bob, 31]", got)
	}
}

func TestSem_Map_MultiNests(t *testing.T) {
	// IndexKeySetStr with multi-element [1, 2] stores as *SlopValue
	got := runE2E(t, `p4{name} = ["bob"]
p4@scores = [1, 2]
|> str(p4@scores)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Map_AddNewInt(t *testing.T) {
	got := runE2E(t, `p5{name} = ["bob"]
p5@score = [100]
|> str(p5@score)`)
	if got != "[100]" {
		t.Fatalf("expected %q, got %q", "[100]", got)
	}
}

// --- ## and @@ (~6 tests) ---

func TestSem_Map_Keys(t *testing.T) {
	got := runE2E(t, `m5{name, age} = ["bob", [30]]
|> str(##m5)`)
	if got != "[name, age]" {
		t.Fatalf("expected %q, got %q", "[name, age]", got)
	}
}

func TestSem_Map_Values(t *testing.T) {
	// Elements: ["bob", *SlopValue{30}]
	// MapValues -> {Elements: ["bob", *SlopValue{30}]}
	// str() -> [bob, [30]]
	got := runE2E(t, `m6{name, age} = ["bob", [30]]
|> str(@@m6)`)
	if got != "[bob, [30]]" {
		t.Fatalf("expected %q, got %q", "[bob, [30]]", got)
	}
}

func TestSem_Map_KeysAfterAdd(t *testing.T) {
	got := runE2E(t, `m7{name} = ["bob"]
m7@age = [25]
|> str(##m7)`)
	if got != "[name, age]" {
		t.Fatalf("expected %q, got %q", "[name, age]", got)
	}
}

func TestSem_Map_ValuesAfterUpdate(t *testing.T) {
	// m8{a} = [[10]] -> Elements: [*SlopValue{10}]
	// m8@a = [20] -> unwraps single-element [20] to int64(20)
	// @@m8 -> MapValues -> {Elements: [int64(20)]}
	// str() -> [20]
	got := runE2E(t, `m8{a} = [[10]]
m8@a = [20]
|> str(@@m8)`)
	if got != "[20]" {
		t.Fatalf("expected %q, got %q", "[20]", got)
	}
}

func TestSem_Map_KeysOnNonMap(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
|> str(##arr)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Map_ValuesOnNonMap(t *testing.T) {
	got := runE2E(t, `arr2 = [1, 2, 3]
|> str(@@arr2)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Dynamic access (~4 tests) ---

func TestSem_Map_StringKeyRead(t *testing.T) {
	got := runE2E(t, `m9{name} = ["bob"]
key = "name"
|> m9$key`)
	if got != "bob" {
		t.Fatalf("expected %q, got %q", "bob", got)
	}
}

func TestSem_Map_IntKeyOnArray(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
i = [0]
|> str(arr$i)`)
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

func TestSem_Map_StringKeyWrite(t *testing.T) {
	got := runE2E(t, `m10{name} = ["bob"]
key = "new"
m10$key = [42]
|> str(m10@new)`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestSem_Map_IntKeyOnMapVia_Dollar(t *testing.T) {
	// DynAccess with int64 key calls Index (positional access on Elements).
	// m11{name} = ["bob"] has Elements: ["bob"] at index 0.
	// i = [0], m11$i -> Index(m11, [0]) -> Elements[0] = "bob" -> NewSlopValue("bob")
	// This does NOT panic: it treats the hashmap's elements as a positional array.
	got := runE2E(t, `m11{name} = ["bob"]
i = [0]
|> m11$i`)
	if got != "bob" {
		t.Fatalf("expected %q, got %q", "bob", got)
	}
}

// --- Hashmap iteration (~2 tests) ---

func TestSem_Map_IterateKeys(t *testing.T) {
	got := runE2E(t, `m12{a, b} = [[1], [2]]
for k in ##m12 {
	|> k
}`)
	if got != "ab" {
		t.Fatalf("expected %q, got %q", "ab", got)
	}
}

func TestSem_Map_IterateValues(t *testing.T) {
	// m13{a, b} = [[1], [2]] -> Elements: [*SlopValue{1}, *SlopValue{2}]
	// @@m13 -> MapValues -> {Elements: [*SlopValue{1}, *SlopValue{2}]}
	// Iterate -> each element is *SlopValue, returned directly
	// str(*SlopValue{1}) -> [1], str(*SlopValue{2}) -> [2]
	got := runE2E(t, `m13{a, b} = [[1], [2]]
for v in @@m13 {
	|> str(v)
}`)
	if got != "[1][2]" {
		t.Fatalf("expected %q, got %q", "[1][2]", got)
	}
}

// --- Hashmap in functions (~4 tests) ---

func TestSem_Map_PassToFn(t *testing.T) {
	got := runE2E(t, `fn getName(m) {
	<- m@name
}
p{name, age} = ["alice", [30]]
|> getName(p)`)
	if got != "alice" {
		t.Fatalf("expected %q, got %q", "alice", got)
	}
}

func TestSem_Map_ReturnFromFn(t *testing.T) {
	got := runE2E(t, `fn makeUser(n, a) {
	u{name, age} = [n, a]
	<- u
}
user = makeUser("charlie", [40])
|> user@name
|> str(user@age)`)
	if got != "charlie[40]" {
		t.Fatalf("expected %q, got %q", "charlie[40]", got)
	}
}

func TestSem_Map_MutateInFn(t *testing.T) {
	// Functions receive pointers — mutations inside functions are visible to caller
	got := runE2E(t, `fn setAge(m, a) {
	m@age = a
	<- m
}
person{name, age} = ["bob", [30]]
setAge(person, [31])
|> str(person@age)`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

func TestSem_Map_BuildInLoop(t *testing.T) {
	got := runE2E(t, `m14{} = []
keys = ["x", "y", "z"]
val = [0]
for k in keys {
	m14$k = val
	val = val + [1]
}
|> str(m14@x)
|> str(m14@y)
|> str(m14@z)`)
	if got != "[0][1][2]" {
		t.Fatalf("expected %q, got %q", "[0][1][2]", got)
	}
}

// --- Hashmap equality (~2 tests) ---

func TestSem_Map_SameStructure(t *testing.T) {
	got := runE2E(t, `a{x, y} = [[1], [2]]
b{x, y} = [[1], [2]]
|> str(a == b)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Map_DifferentValues(t *testing.T) {
	got := runE2E(t, `c{x, y} = [[1], [2]]
d{x, y} = [[1], [3]]
|> str(c == d)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}
