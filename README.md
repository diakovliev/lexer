# lexer

General purpose lexer.

## Goals

* Minimal dependencies.
* Minimal memory usage. So, the library should not allocate memory if possible and do not hold data in memory if possible.
* Lexers should be able to be written in any language itself. So, no code generation like in ANTLR or Yacc.
* Functional programming style where it have sense.
* The library should be easy to use and understand.
* The library should be easy to extend with new features.
* The library should be able to work with any input stream, not only strings.
* The library should be able to process both text and binary data in similar way. The mixed input streams are also should be supported.
* Unicode, in particular UTF-8, should be fully supported.
* 100% test coverage.

## Non-goals

* Performance. The primary goal is to make the lexer easy to use, support and understand. If it against performance, so it be.
* No regexp based states. The library itself is almost regexp engine. So, no need in additional regexp engine.

## Overview

In essence the library is a specialized state machine with some additional features. It can be used to parse any kind of input stream and produce tokens.

A state machine is driven by a set of rules (states), and by the input stream. All rules are combined into a state transition table.

The matching rules itself are combined from primitive rules, such as: is last byte in range 0x30..0x39 (is digit), or is last byte equal to 0x2E (is dot), etc. Primitive rules are combined in chains, and at the end of the chain there is an action. The actions can be: emit token, or emit error.

See examples/[calculator/grammar.go](examples/calculator/grammar.go) for example.
