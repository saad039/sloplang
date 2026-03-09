package parser

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
}

// Stmt is a statement node.
type Stmt interface {
	Node
	stmtNode()
}

// Expr is an expression node.
type Expr interface {
	Node
	exprNode()
}

// Program is the root AST node containing all statements.
type Program struct {
	Statements []Stmt
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// AssignStmt represents: name = value
type AssignStmt struct {
	Name  string
	Value Expr
}

func (as *AssignStmt) stmtNode()            {}
func (as *AssignStmt) TokenLiteral() string { return as.Name }

// StdoutWriteStmt represents: |> value
type StdoutWriteStmt struct {
	Value Expr
}

func (sw *StdoutWriteStmt) stmtNode()            {}
func (sw *StdoutWriteStmt) TokenLiteral() string { return "|>" }

// ArrayLiteral represents: [elem1, elem2, ...]
type ArrayLiteral struct {
	Elements []Expr
}

func (al *ArrayLiteral) exprNode()            {}
func (al *ArrayLiteral) TokenLiteral() string { return "[" }

// NumberType distinguishes between int, uint, and float literals.
type NumberType int

const (
	NumInt NumberType = iota
	NumUint
	NumFloat
)

// NumberLiteral represents a numeric value: 42, 42u, 3.14
type NumberLiteral struct {
	Value   string
	NumType NumberType
}

func (nl *NumberLiteral) exprNode()            {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Value }

// StringLiteral represents: "hello world"
type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) exprNode()            {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Value }

// Identifier represents a variable reference.
type Identifier struct {
	Name string
}

func (id *Identifier) exprNode()            {}
func (id *Identifier) TokenLiteral() string { return id.Name }

// BinaryExpr represents: left op right
type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (be *BinaryExpr) exprNode()            {}
func (be *BinaryExpr) TokenLiteral() string { return be.Op }

// UnaryExpr represents: op operand (prefix)
type UnaryExpr struct {
	Op      string
	Operand Expr
}

func (ue *UnaryExpr) exprNode()            {}
func (ue *UnaryExpr) TokenLiteral() string { return ue.Op }

// CallExpr represents: name(args...)
type CallExpr struct {
	Name string
	Args []Expr
}

func (ce *CallExpr) exprNode()            {}
func (ce *CallExpr) TokenLiteral() string { return ce.Name }

// FnDeclStmt represents: fn name(params) { body }
type FnDeclStmt struct {
	Name   string
	Params []string
	Body   []Stmt
}

func (fd *FnDeclStmt) stmtNode()            {}
func (fd *FnDeclStmt) TokenLiteral() string { return "fn" }

// IfStmt represents: if condition { body } else { elseBody }
type IfStmt struct {
	Condition Expr
	Body      []Stmt
	Else      []Stmt // nil if no else branch
}

func (is *IfStmt) stmtNode()            {}
func (is *IfStmt) TokenLiteral() string { return "if" }

// ForInStmt represents: for varName in iterable { body }
type ForInStmt struct {
	VarName  string
	Iterable Expr
	Body     []Stmt
}

func (fi *ForInStmt) stmtNode()            {}
func (fi *ForInStmt) TokenLiteral() string { return "for" }

// ReturnStmt represents: <- value
type ReturnStmt struct {
	Value Expr // nil for bare return
}

func (rs *ReturnStmt) stmtNode()            {}
func (rs *ReturnStmt) TokenLiteral() string { return "<-" }

// MultiAssignStmt represents: a, b = expr
type MultiAssignStmt struct {
	Names []string
	Value Expr
}

func (ma *MultiAssignStmt) stmtNode()            {}
func (ma *MultiAssignStmt) TokenLiteral() string { return "=" }

// ForLoopStmt represents: for { body } (infinite loop)
type ForLoopStmt struct {
	Body []Stmt
}

func (fl *ForLoopStmt) stmtNode()            {}
func (fl *ForLoopStmt) TokenLiteral() string { return "for" }

// BreakStmt represents: break
type BreakStmt struct{}

func (bs *BreakStmt) stmtNode()            {}
func (bs *BreakStmt) TokenLiteral() string { return "break" }

// ExprStmt represents a bare expression as a statement (e.g., function call).
type ExprStmt struct {
	Expr Expr
}

func (es *ExprStmt) stmtNode()            {}
func (es *ExprStmt) TokenLiteral() string { return es.Expr.TokenLiteral() }
