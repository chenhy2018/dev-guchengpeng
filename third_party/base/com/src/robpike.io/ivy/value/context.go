// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package value

// Expr and Context are defined here to avoid import cycles
// between parse and value.

// Expr is the interface for a parsed expression.
type Expr interface {
	String() string

	Eval(Context) Value
}

// Context is the execution context for evaluation.
type Context interface {
	// Lookup returns the value of a symbol.
	Lookup(name string) Value

	// AssignLocal binds a value to the name in the current function.
	AssignLocal(name string, value Value)

	// Assign assigns the variable the value. The variable must
	// be defined either in the current function or globally.
	// Inside a function, new variables become locals.
	Assign(name string, value Value)

	// Push pushes a new frame onto the context stack.
	Push()

	// Pop pops the top frame from the stack.
	Pop()

	// Eval evaluates a list of expressions.
	Eval(exprs []Expr) []Value
}
