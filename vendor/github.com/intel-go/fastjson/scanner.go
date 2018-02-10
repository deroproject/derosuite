// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fastjson

// JSON value parser state machine.
// Just about at the limit of what is reasonable to write by hand.
// Some parts are a bit tedious, but overall it nicely factors out the
// otherwise common code from the multiple scanning functions
// in this package (Compact, Indent, checkValid, nextValue, etc).
//
// This file starts with two simple examples using the scanner
// before diving into the scanner itself.

import (
	"bytes"
	"strconv"
)

// checkValid verifies that data is valid JSON-encoded data.
// scan is passed in for use by checkValid to avoid an allocation.
func checkValid(data []byte, scan *scanner) error {
	scan.length_data = len(data)
	scan.reset()
	scan.endTop = true
	stream := streamByte{data: &data, pos: 0}
	op := scan.parseValue(&stream)

	if op == scanError {
		return scan.err
	}

	return nil
}

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

type Record struct {
	state int
	pos   int
}

func (e *SyntaxError) Error() string { return e.msg }

// A scanner is a JSON scanning state machine.
// Callers call scan.reset() and then pass bytes in one at a time
// by calling scan.step(&scan, c) for each byte.
// The return value, referred to as an opcode, tells the
// caller about significant parsing events like beginning
// and ending literals, objects, and arrays, so that the
// caller can follow along if it wishes.
// The return value scanEnd indicates that a single top-level
// JSON value has been completed, *before* the byte that
// just got passed in.  (The indication must be delayed in order
// to recognize the end of numbers: is 123 a whole value or
// the beginning of 12345e+6?).
type scanner struct {
	// The step is a func to be called to execute the next transition.
	// Also tried using an integer constant and a single func
	// with a switch, but using the func directly was 10% faster
	// on a 64-bit Mac Mini, and it's nicer to read.
	step func(*scanner, byte) int

	// Reached end of top-level value.
	endTop bool

	// Stack of what we're in the middle of - array values, object keys, object values.
	parseState []int

	// Error that happened, if any.
	err error

	stateRecord          []Record//array of records of labels(position in array and state on this position)
	cacheRecord	     Record
	cached               bool
	readPos              int  //position in array stateRecord during filling
	length_data          int  //length of data to read, initialized in unmarshal. Helps to set correct capacity of stateRecord
	inNumber             bool // flag of parsing figure
	endLiteral           bool //flag of finishing literal

	bytes int64 // total bytes consumed, updated by decoder.Decode
}

// These values are returned by the state transition functions
// assigned to scanner.state and the method scanner.eof.
// They give details about the current state of the scan that
// callers might be interested to know about.
// It is okay to ignore the return value of any particular
// call to scanner.state: if one call returns scanError,
// every subsequent call will return scanError too.
const (
	scanBeginLiteral = iota // end implied by next result != scanContinue
	scanEndLiteral          // not returned by scanner, but clearer for state recording
	scanBeginObject         // begin object
	scanEndObject           // end object (implies scanObjectValue if possible)
	scanBeginArray          // begin array
	scanEndArray            // end array (implies scanArrayValue if possible)
	scanObjectKey           // just finished object key (string)
	scanObjectValue         // just finished non-last object value
	scanContinue            // uninteresting byte
	scanArrayValue          // just finished array value

	scanSkipSpace // space byte; can skip; known to be last "continue" result

	// Stop.
	scanEnd   // top-level value ended *before* this byte; known to be first "stop" result
	scanError // hit an error, scanner.err.
)

// These values are stored in the parseState stack.
// They give the current state of a composite value
// being scanned. If the parser is inside a nested value
// the parseState describes the nested state, outermost at entry 0.
const (
	parseObjectKey   = iota // parsing object key (before colon)
	parseObjectValue        // parsing object value (after colon)
	parseArrayValue         // parsing array value
)

type streamByte struct {
	data *[]byte
	pos  int
}

func (s *streamByte) isEnd() bool {
	return s.pos >= len(*s.data)
}

func (s *streamByte) Take() byte {
	result := s.Peek()
	s.pos++
	return result
}

func (s *streamByte) Peek() byte {
	if !s.isEnd() {
		return (*s.data)[s.pos]
	} else {
		return 0//I have to leave this case, because I can call Peek, when the stream is over. I won't use value it returns, but it should be protected 
	}
}

func (sb *streamByte) skipSpaces() {
	for c := sb.Peek(); c <= ' ' && isSpace(c); {
		sb.pos++
		c = sb.Peek()
	}
}

const AVERAGE_LENGTH = 10000

// reset prepares the scanner for use.
// It must be called before calling s.step.
func (s *scanner) reset() {
	s.step = stateBeginValue
	s.parseState = s.parseState[0:0]
	s.err = nil
	if s.isRecordEmpty() {
		if s.length_data >= AVERAGE_LENGTH {
			s.stateRecord = make([]Record, 0, s.length_data/4) //capacity doesn't depends on the length whole value, but on the length of nested values. But predictively the large values have large nested values.
		} else {
			s.stateRecord = make([]Record, 0, s.length_data/2)
		}
	}
	s.inNumber = false
	s.endLiteral = false
	s.cached = false
	s.readPos = 0
	s.endTop = false
}

// eof tells the scanner that the end of input has been reached.
// It returns a scan status just as s.step does.
func (s *scanner) eof() int {
	if s.err != nil {
		return scanError
	}
	if s.endTop {
		return scanEnd
	}
	s.step(s, ' ')
	if s.endTop {
		return scanEnd
	}
	if s.err == nil {
		s.err = &SyntaxError{"unexpected end of JSON input", s.bytes}
	}
	return scanError
}

// pushParseState pushes a new parse state p onto the parse stack.
func (s *scanner) pushParseState(p int) {
	s.parseState = append(s.parseState, p)
}

// popParseState pops a parse state (already obtained) off the stack
// and updates s.step accordingly.
func (s *scanner) popParseState() {
	n := len(s.parseState) - 1
	s.parseState = s.parseState[0:n]
	if n == 0 {
		s.step = stateEndTop
		s.endTop = true
	} else {
		s.step = stateEndValue
	}
}

//checks if array of records is empty
func (s *scanner) isRecordEmpty() bool {
	return len(s.stateRecord) == 0
}

//pushes Record into array
func (s *scanner) pushRecord(state, pos int) {
	s.stateRecord = append(s.stateRecord, Record{state:state, pos:pos}) //state are at even positions, pos are at odd positions in stateRecord array
}

//peeks current state for filling object. Doesn't change position. Returns state, pos
func (s *scanner) peekPos() int {
	if s.readPos >= len(s.stateRecord){
		return  s.cacheRecord.pos// peek can be called when the array is over , only if unmarshal error occured, so return last read position 
	}
	if !s.cached {
		s.cached = true
		s.cacheRecord = s.stateRecord[s.readPos]
	}
	return s.cacheRecord.pos
}

func (s *scanner) peekState() int {
	if s.readPos >= len(s.stateRecord) {
	    return s.cacheRecord.state  // the same as Peek 
	}
	if !s.cached {
		s.cached = true
		s.cacheRecord = s.stateRecord[s.readPos]
	}
	return s.cacheRecord.state
}

//takes current state and increments reading position.
func (s *scanner) takeState() int {
	if s.cached {
		s.cached = false
	}else{
	    s.peekState()
	}
	s.readPos += 1
	return s.cacheRecord.state
}

func (s *scanner) takePos() int {
	if s.cached {
		s.cached = false
	}else{
	    s.peekState()
	}
	s.readPos += 1
	return s.cacheRecord.pos
}

func (s *scanner) skipRecord() {
	s.readPos += 1
	s.cached = false
}

//checks if we need this state to be recorded
func (s *scanner) isNeededState(state int) bool {
	if s.endLiteral {
		return true
	}
	if state > scanEndArray || state < scanBeginLiteral {
		return false
	}
	return true
}

func (s *scanner) fillRecord(pos, state int) {

	if s.isNeededState(state) {
		if s.inNumber && s.endLiteral { // in case 2] , 2} or 2,
			s.inNumber = false
			s.endLiteral = false
			s.pushRecord(scanEndLiteral, pos-1)
			if s.isNeededState(state) { // in case 2] or 2}
				s.pushRecord(state, pos)
			}
			return
		}

		if s.endLiteral {
			s.endLiteral = false
			state = scanEndLiteral
		}
		s.pushRecord(state, pos)
	}

}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

// stateBeginValueOrEmpty is the state after reading `[`.
func stateBeginValueOrEmpty(s *scanner, c byte) int {
	if c <= ' ' && isSpace(c) {
		return scanSkipSpace
	}
	if c == ']' {
		return stateEndValue(s, c)
	}
	return stateBeginValue(s, c)
}

// stateBeginValue is the state at the beginning of the input.
func stateBeginValue(s *scanner, c byte) int {
	if c <= ' ' && isSpace(c) {
		return scanSkipSpace
	}
	switch c {
	case '{':
		s.step = stateBeginStringOrEmpty
		s.pushParseState(parseObjectKey)
		return scanBeginObject
	case '[':
		s.step = stateBeginValueOrEmpty
		s.pushParseState(parseArrayValue)
		return scanBeginArray
	case '"':
		s.step = stateInString
		return scanBeginLiteral
	case '-':
		s.step = stateNeg
		s.inNumber = true
		return scanBeginLiteral
	case '0': // beginning of 0.123
		s.step = state0
		s.inNumber = true
		return scanBeginLiteral
	case 't': // beginning of true
		s.step = stateT
		return scanBeginLiteral
	case 'f': // beginning of false
		s.step = stateF
		return scanBeginLiteral
	case 'n': // beginning of null
		s.step = stateN
		return scanBeginLiteral
	}
	if '1' <= c && c <= '9' { // beginning of 1234.5
		s.step = state1
		s.inNumber = true
		return scanBeginLiteral
	}
	return s.error(c, "looking for beginning of value")
}

// stateBeginStringOrEmpty is the state after reading `{`.
func stateBeginStringOrEmpty(s *scanner, c byte) int {
	if c <= ' ' && isSpace(c) {
		return scanSkipSpace
	}
	if c == '}' {
		n := len(s.parseState)
		s.parseState[n-1] = parseObjectValue
		return stateEndValue(s, c)
	}
	return stateBeginString(s, c)
}

// stateBeginString is the state after reading `{"key": value,`.
func stateBeginString(s *scanner, c byte) int {
	if c <= ' ' && isSpace(c) {
		return scanSkipSpace
	}
	if c == '"' {
		s.step = stateInString
		return scanBeginLiteral
	}
	return s.error(c, "looking for beginning of object key string")
}

// stateEndValue is the state after completing a value,
// such as after reading `{}` or `true` or `["x"`.
func stateEndValue(s *scanner, c byte) int {
	n := len(s.parseState)
	if n == 0 {
		// Completed top-level before the current byte.
		s.step = stateEndTop
		s.endTop = true
		return stateEndTop(s, c)
	}
	if c <= ' ' && isSpace(c) {
		s.step = stateEndValue
		return scanSkipSpace
	}
	ps := s.parseState[n-1]
	switch ps {
	case parseObjectKey:
		if c == ':' {
			s.parseState[n-1] = parseObjectValue
			s.step = stateBeginValue
			return scanObjectKey
		}
		return s.error(c, "after object key")
	case parseObjectValue:
		if c == ',' {
			s.parseState[n-1] = parseObjectKey
			s.step = stateBeginString
			return scanObjectValue
		}
		if c == '}' {
			s.popParseState()
			return scanEndObject
		}
		return s.error(c, "after object key:value pair")
	case parseArrayValue:
		if c == ',' {
			s.step = stateBeginValue
			return scanArrayValue
		}
		if c == ']' {
			s.popParseState()
			return scanEndArray
		}
		return s.error(c, "after array element")
	}
	return s.error(c, "")
}

// stateEndTop is the state after finishing the top-level value,
// such as after reading `{}` or `[1,2,3]`.
// Only space characters should be seen now.
func stateEndTop(s *scanner, c byte) int {
	if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
		// Complain about non-space byte on next call.
		s.error(c, "after top-level value")
	}
	return scanEnd
}

// stateInString is the state after reading `"`.
func stateInString(s *scanner, c byte) int {
	if c == '"' {
		s.step = stateEndValue
		s.endLiteral = true
		return scanContinue
	}
	if c == '\\' {
		s.step = stateInStringEsc
		return scanContinue
	}
	if c < 0x20 {
		return s.error(c, "in string literal")
	}
	return scanContinue
}

// stateInStringEsc is the state after reading `"\` during a quoted string.
func stateInStringEsc(s *scanner, c byte) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		s.step = stateInString
		return scanContinue
	case 'u':
		s.step = stateInStringEscU
		return scanContinue
	}
	return s.error(c, "in string escape code")
}

// stateInStringEscU is the state after reading `"\u` during a quoted string.
func stateInStringEscU(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU1
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU1 is the state after reading `"\u1` during a quoted string.
func stateInStringEscU1(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU12
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU12 is the state after reading `"\u12` during a quoted string.
func stateInStringEscU12(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU123
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU123 is the state after reading `"\u123` during a quoted string.
func stateInStringEscU123(s *scanner, c byte) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInString
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateNeg is the state after reading `-` during a number.
func stateNeg(s *scanner, c byte) int {
	if c == '0' {
		s.step = state0
		return scanContinue
	}
	if '1' <= c && c <= '9' {
		s.step = state1
		return scanContinue
	}
	return s.error(c, "in numeric literal")
}

// state1 is the state after reading a non-zero integer during a number,
// such as after reading `1` or `100` but not `0`.
func state1(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		s.step = state1
		return scanContinue
	}
	return state0(s, c)
}

// state0 is the state after reading `0` during a number.
func state0(s *scanner, c byte) int {
	if c == '.' {
		s.step = stateDot
		return scanContinue
	}
	if c == 'e' || c == 'E' {
		s.step = stateE
		return scanContinue
	}
	s.endLiteral = true
	return stateEndValue(s, c)
}

// stateDot is the state after reading the integer and decimal point in a number,
// such as after reading `1.`.
func stateDot(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0
		return scanContinue
	}
	return s.error(c, "after decimal point in numeric literal")
}

// stateDot0 is the state after reading the integer, decimal point, and subsequent
// digits of a number, such as after reading `3.14`.
func stateDot0(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		return scanContinue
	}
	if c == 'e' || c == 'E' {
		s.step = stateE
		return scanContinue
	}
	s.endLiteral = true
	return stateEndValue(s, c)
}

// stateE is the state after reading the mantissa and e in a number,
// such as after reading `314e` or `0.314e`.
func stateE(s *scanner, c byte) int {
	if c == '+' || c == '-' {
		s.step = stateESign
		return scanContinue
	}
	return stateESign(s, c)
}

// stateESign is the state after reading the mantissa, e, and sign in a number,
// such as after reading `314e-` or `0.314e+`.
func stateESign(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		s.step = stateE0
		return scanContinue
	}
	return s.error(c, "in exponent of numeric literal")
}

// stateE0 is the state after reading the mantissa, e, optional sign,
// and at least one digit of the exponent in a number,
// such as after reading `314e-2` or `0.314e+1` or `3.14e0`.
func stateE0(s *scanner, c byte) int {
	if '0' <= c && c <= '9' {
		return scanContinue
	}
	s.endLiteral = true
	return stateEndValue(s, c)
}

// stateT is the state after reading `t`.
func stateT(s *scanner, c byte) int {
	if c == 'r' {
		s.step = stateTr
		return scanContinue
	}
	return s.error(c, "in literal true (expecting 'r')")
}

// stateTr is the state after reading `tr`.
func stateTr(s *scanner, c byte) int {
	if c == 'u' {
		s.step = stateTru
		return scanContinue
	}
	return s.error(c, "in literal true (expecting 'u')")
}

// stateTru is the state after reading `tru`.
func stateTru(s *scanner, c byte) int {
	if c == 'e' {
		s.step = stateEndValue
		s.endLiteral = true
		return scanContinue
	}
	return s.error(c, "in literal true (expecting 'e')")
}

// stateF is the state after reading `f`.
func stateF(s *scanner, c byte) int {
	if c == 'a' {
		s.step = stateFa
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 'a')")
}

// stateFa is the state after reading `fa`.
func stateFa(s *scanner, c byte) int {
	if c == 'l' {
		s.step = stateFal
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 'l')")
}

// stateFal is the state after reading `fal`.
func stateFal(s *scanner, c byte) int {
	if c == 's' {
		s.step = stateFals
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 's')")
}

// stateFals is the state after reading `fals`.
func stateFals(s *scanner, c byte) int {
	if c == 'e' {
		s.step = stateEndValue
		s.endLiteral = true
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 'e')")
}

// stateN is the state after reading `n`.
func stateN(s *scanner, c byte) int {
	if c == 'u' {
		s.step = stateNu
		return scanContinue
	}
	return s.error(c, "in literal null (expecting 'u')")
}

// stateNu is the state after reading `nu`.
func stateNu(s *scanner, c byte) int {
	if c == 'l' {
		s.step = stateNul
		return scanContinue
	}
	return s.error(c, "in literal null (expecting 'l')")
}

// stateNul is the state after reading `nul`.
func stateNul(s *scanner, c byte) int {
	if c == 'l' {
		s.step = stateEndValue
		s.endLiteral = true
		return scanContinue
	}
	return s.error(c, "in literal null (expecting 'l')")
}

// stateError is the state after reaching a syntax error,
// such as after reading `[1}` or `5.1.2`.
func stateError(s *scanner, c byte) int {
	return scanError
}

// error records an error and switches to the error state.
func (s *scanner) error(c byte, context string) int {
	s.step = stateError
	s.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, s.bytes}
	return scanError
}

// quoteChar formats c as a quoted character literal
func quoteChar(c byte) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}

func (sb *streamByte) error(s *scanner, context string) int {
	s.err = &SyntaxError{"invalid character " + quoteChar(sb.Peek()) + " " + context, int64(sb.pos + 1)}
	return scanError
}

func (s *scanner) parseSimpleLiteral(sb *streamByte, length int) int {
	if len(*sb.data) < sb.pos+length {
		s.err = &SyntaxError{"unexpected end of JSON input", int64(len(*sb.data))}
		return scanError
	}
	s.pushRecord(scanBeginLiteral, sb.pos)
	sb.Take()
	s.bytes = int64(sb.pos)
	for i := 0; i < length-1; i++ {
		s.bytes++
		op := s.step(s, sb.Take())
		if op == scanError {
			return op
		}
	}
	s.pushRecord(scanEndLiteral, sb.pos-1)
	return scanContinue
}

func (s *scanner) parseValue(sb *streamByte) int {
	sb.skipSpaces()
	topValue := s.endTop
	if len(*sb.data) <= sb.pos {
		s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
		return scanError
	}
	cur := sb.Peek()
	s.endTop = false
	op := scanContinue
	switch cur {
	case '"':
		op = s.parseString(sb)
	case '{':
		op = s.parseObject(sb)
	case '[':
		op = s.parseArray(sb)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		op = s.parseNumber(sb)
	case 't':
		s.step = stateT
		op = s.parseSimpleLiteral(sb, 4)

	case 'f':
		s.step = stateF
		op = s.parseSimpleLiteral(sb, 5)
	case 'n':
		s.step = stateN
		op = s.parseSimpleLiteral(sb, 4)
	default:
		return sb.error(s, "looking for beginning of value")
	}

	if topValue && op != scanError {
		sb.skipSpaces()
		if !sb.isEnd() {
			return sb.error(s, "after top-level value")
		}
	}

	return op
}

func (s *scanner) parseString(sb *streamByte) int {
	s.pushRecord(scanBeginLiteral, sb.pos)
	sb.pos++ //skip "
	quotePos := bytes.IndexByte((*sb.data)[sb.pos:], '"')
	if quotePos < 0 {
		s.err = &SyntaxError{"unexpected end of JSON input", int64(len(*sb.data))}
		return scanError

	}

	// in case without escape symbol \". Errors inside string will be handled during object filling, with function unquote
	// it's done in sake of speed
	sb.pos += quotePos - 1
	for sb.Peek() == '\\' { //pos on the symbol before "
		//it may escape symbol "
		sb.pos--
		sum := 1

		//checking multiple symbols \, kind of "...\\"..."
		for sb.Peek() == '\\' {
			sum++
			sb.pos--
		}
		if sum%2 == 0 { //even number of \, last of them doesn't escape "; it means that current qoute pos is end of string
			sb.pos += sum
			break
		}
		//otherwise odd number of \ escapes ". Looking for the next "
		sb.pos += sum + 1 // pos on "
		n := bytes.IndexByte((*sb.data)[sb.pos+1:], '"')
		if n < 0 {
			s.err = &SyntaxError{"unexpected end of JSON input", int64(len(*sb.data))}
			return scanError
		}
		sb.pos += n
	}
	//here pos is on the symbol before "
	sb.pos += 2
	s.pushRecord(scanEndLiteral, sb.pos-1)
	return scanEndLiteral
}

func (s *scanner) parseNumber(sb *streamByte) int {
	s.pushRecord(scanBeginLiteral, sb.pos)
	cur := sb.Take()
	if cur == '-' {
		if sb.isEnd() {
			s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
			return scanError
		}
		cur = sb.Take()
	}
	if sb.isEnd() {
		if '0' <= cur && cur <= '9' {
			s.pushRecord(scanEndLiteral, sb.pos-1)
			return scanEndLiteral
		} else {
			sb.pos--
			return sb.error(s, "in numeric literal")
		}
	}
	if !sb.isEnd() && '1' <= cur && cur <= '9' {
		sb.parseFigures()
	} else {
		if cur != '0' {
			sb.pos--
			return sb.error(s, "in numeric literal")
		}
	}
	cur = sb.Take() //pos on the next after cur
	if cur == '.' {
		if op := sb.Peek(); op > '9' || op < '0' {
			if sb.isEnd() {
				s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
				return scanError
			}
			return sb.error(s, "after decimal point in numeric literal")
		}
		sb.parseFigures()
		cur = sb.Take()
	}
	if cur == 'e' || cur == 'E' {
		op := sb.Peek()
		if op != '+' && op != '-' && (op < '0' || op > '9') {
			if sb.isEnd() {
				s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
				return scanError
			}
			return sb.error(s, "in exponent of numeric literal")
		}
		op = sb.Take()
		if op == '-' || op == '+' {
			op = sb.Peek()
			if op < '0' || op > '9' {
				if sb.isEnd() {
					s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
					return scanError
				}
				return sb.error(s, "in exponent of numeric literal")
			}
		}

		sb.parseFigures()

	} else { //pos on the second after unknown symbol. like 123ua. pos now at a
		sb.pos--
	}
	s.pushRecord(scanEndLiteral, sb.pos-1)
	return scanEndLiteral
}

func (sb *streamByte) parseFigures() {
	c := sb.Take()

	for '0' <= c && c <= '9' {
		c = sb.Take()
	}
	sb.pos--
}

func (s *scanner) parseObject(sb *streamByte) int {
	s.pushRecord(scanBeginObject, sb.pos)
	sb.pos++ // skip {
	sb.skipSpaces()
	cur := sb.Peek()
	if sb.isEnd() {
		s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
		return scanError
	}

	if cur != '"' && cur != '}' {
		return sb.error(s, "looking for beginning of object key string")
	}
	for !sb.isEnd() {
		sb.skipSpaces()

		switch cur {
		case '}':
			s.pushRecord(scanEndObject, sb.pos)
			sb.pos++
			return scanEndObject
		case '"':
			op := s.parseString(sb)
			if op == scanError {
				return op
			}
			sb.skipSpaces()
			if sb.isEnd() {
				s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
				return scanError
			}
			cur = sb.Peek()

			if cur == ':' {
				sb.pos++
			} else {
				return sb.error(s, "after object key")
			}
			op = s.parseValue(sb)
			if op == scanError {
				return op
			}
			sb.skipSpaces()
			if sb.isEnd() {
				s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
				return scanError
			}
			cur = sb.Peek()
			if cur == ',' {
				sb.pos++
				if sb.isEnd() {
					s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
					return scanError
				}
				sb.skipSpaces()
				cur = sb.Peek()
			}
		default:
			return sb.error(s, "after object key:value pair")
		}
	}

	s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
	return scanError
}

func (s *scanner) parseArray(sb *streamByte) int {
	s.pushRecord(scanBeginArray, sb.pos)
	sb.pos++
	sb.skipSpaces()
	if sb.isEnd() {
		s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
		return scanError
	}
	cur := sb.Peek()
	if cur == ']' {
		s.pushRecord(scanEndArray, sb.pos)
		sb.pos++
		return scanEndArray
	}
	op := s.parseValue(sb)
	if op == scanError {
		return op
	}
	sb.skipSpaces()

	cur = sb.Peek()
	for !sb.isEnd() {
		switch cur {
		case ']':
			s.pushRecord(scanEndArray, sb.pos)
			sb.pos++
			return scanEndArray
		case ',':
			sb.pos++
			sb.skipSpaces()
		default:
			return sb.error(s, "after array element")
		}

		op = s.parseValue(sb)
		if op == scanError {
			return op
		}

		sb.skipSpaces()
		cur = sb.Peek()
	}

	//here is incomplete array
	s.err = &SyntaxError{"unexpected end of JSON input", int64(sb.pos)}
	return scanError

}
