// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config // import "robpike.io/ivy/config"

import (
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// A Config holds information about the configuration of the system.
// The zero value of a Config, or a nil Config pointer, represents the default
// values for all settings.
type Config struct {
	prompt      string
	format      string
	ratFormat   string
	formatVerb  byte // The verb if format is floating-point.
	formatPrec  int  // The precision if format is floating-point.
	formatFloat bool // Whether format is floating-point.
	origin      int
	bigOrigin   *big.Int
	debug       map[string]bool
	source      rand.Source
	random      *rand.Rand
	// Bases: 0 means C-like, base 10 with 07 for octal and 0xa for hex.
	inputBase  int
	outputBase int
}

func (c *Config) init() {
	if c.random == nil {
		c.source = rand.NewSource(time.Now().Unix())
		c.random = rand.New(c.source)
	}
}

// Format returns the formatting string. If empty, the default
// formatting is used, as defined by the bases.
func (c *Config) Format() string {
	if c == nil {
		return ""
	}
	return c.format
}

// Format returns the formatting string for rationals.
func (c *Config) RatFormat() string {
	if c == nil {
		return "%v/%v"
	}
	return c.ratFormat
}

// SetFormat sets the formatting string. Rational formatting
// is just this format applied twice with a / in between.
func (c *Config) SetFormat(s string) {
	c.formatVerb = 0
	c.formatPrec = 0
	c.formatFloat = false
	c.format = s
	if s == "" {
		c.ratFormat = "%v/%v"
		return
	}
	c.ratFormat = s + "/" + s
	// Is it a floating-point format?
	switch s[len(s)-1] {
	case 'f', 'F', 'g', 'G', 'e', 'E':
		// Yes
	default:
		return
	}
	c.formatFloat = true
	c.formatVerb = s[len(s)-1]
	c.formatPrec = 6 // The default
	point := strings.LastIndex(s, ".")
	if point > 0 {
		prec, err := strconv.ParseInt(s[point+1:len(s)-1], 10, 32)
		if err == nil && prec >= 0 {
			c.formatPrec = int(prec)
		}

	}
}

// FloatFormat returns the parsed information about the format,
// if it's a floating-point format.
func (c *Config) FloatFormat() (verb byte, prec int, ok bool) {
	return c.formatVerb, c.formatPrec, c.formatFloat
}

// Debug returns the value of the specified boolean debugging flag.
func (c *Config) Debug(flag string) bool {
	if c == nil {
		return false
	}
	return c.debug[flag]
}

// SetDebug sets the value of the specified boolean debugging flag.
func (c *Config) SetDebug(flag string, state bool) {
	if c.debug == nil {
		c.debug = make(map[string]bool)
	}
	c.debug[flag] = state
}

// Origin returns the index origin, default 1.
func (c *Config) Origin() int {
	if c == nil {
		return 1
	}
	return c.origin
}

// BigOrigin returns the index origin as a *big.Int.
func (c *Config) BigOrigin() *big.Int {
	if c == nil {
		return big.NewInt(1)
	}
	return c.bigOrigin
}

// SetOrigin sets the index origin.
func (c *Config) SetOrigin(origin int) {
	c.origin = origin
	c.bigOrigin = big.NewInt(int64(origin))
}

// Prompt returns the interactive prompt.
func (c *Config) Prompt() string {
	if c == nil {
		return ""
	}
	return c.prompt
}

// SetPrompt sets the interactive prompt.
func (c *Config) SetPrompt(prompt string) {
	c.prompt = prompt
}

// Random returns the generator for random numbers.
func (c *Config) Random() *rand.Rand {
	c.init()
	return c.random
}

// RandomSeed sets the seed for the random number generator.
func (c *Config) RandomSeed(seed int64) {
	c.init()
	c.source.Seed(seed)
}

// Base returns the input and output bases.
func (c *Config) Base() (inputBase, outputBase int) {
	if c == nil {
		return 0, 0
	}
	return c.inputBase, c.outputBase
}

// InputBase returns the input base.
func (c *Config) InputBase() int {
	if c == nil {
		return 0
	}
	return c.inputBase
}

// OutputBase returns the output base.
func (c *Config) OutputBase() int {
	if c == nil {
		return 0
	}
	return c.outputBase
}

// SetBase sets the input and output bases.
func (c *Config) SetBase(inputBase, outputBase int) {
	c.inputBase = inputBase
	c.outputBase = outputBase
}
