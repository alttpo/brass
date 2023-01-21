// Package brass
//
// brass is a custom s-expressions encoding.
//
//	author: jsd1982
//	date:   2023-01-07
//
// a brass s-expression is recursively either an atom expression or a list of s-expressions surrounded by '(' ')'.
//
// encoding restrictions:
//  1. newline whitespace characters ('\r', '\n') MUST NOT appear in an encoded s-expression.
//  2. encoding is ASCII and 8-bit clean with exception of '\r' and '\n'.
//  3. each atom has a single encoding with no stylistic variation so as to simplify encoding and decoding logic.
//
// s-expression examples:
//
//	(test_exp abc d.e.f/gh snake_case nil true false #616263# ^3#616263#)
//	(^$a#0102030405060708090a# 1023 $3ff)
//	("abc\ndef\t\"123\"\x00\xff" () ^5"12345")
//	((a 1) (b 2) (c 3) (d nil) (e false))
//
// atom types:
//
//	nil
//	bool
//	integer
//	token-octets
//	hex-octets
//	quoted-octets
//	list
//
// nil atom type:
//
//	literal "nil" keyword
//
// bool atom type:
//
//	literal "true" and "false" keywords
//
// integer atom type:
//
//	a base-10 or base-16 integer value of 64-bit length
//	may start with optional '-' to indicate signed type and negative value
//	or may start with '+' suffix to indicate unsigned type and positive value
//	if starts with neither '-' nor '+' then type is signed and value is positive
//	integers contain only allowable digit characters depending on the base
//	no extra formatting-related ('_'), division (','), or white-space characters are allowed
//	any number of leading zeros are allowed and *do not* signify base-8
//	the default base is 10
//	base-10 integer must contain only digits '0'..'9'
//	base-16 integer must start with '$'
//	base-16 integer must contain only digits '0'..'9','a'..'f'
//	base-16 integer cannot contain upper-case 'A'..'F' so as to simplify encoding and decoding logic
//
//	examples:
//		  `-1024`  =   int64(-1024)
//		`0001023`  =   int64( 1023)
//		   `1023`  =   int64( 1023)
//		   `$3ff`  =   int64( 1023)
//		  `+$3ff`  =  uint64( 1023)
//		  `+1023`  =  uint64( 1023)
//
// token-octets atom type:
//
//	alpha-numeric identifier of arbitrary length without white-space
//	may begin with one optional '@' to escape reserved keywords like "nil", "true", "false"
//	cannot start with a decimal digit
//	may contain alpha characters 'a' .. 'z', 'A' .. 'Z'
//	may contain special punctuation chars "_" | "." | "/" | "?" | "!"
//	may contain non-ASCII characters 128 <= char <= 255
//	may contain decimal digits '0' .. '9'
//
//	examples:
//	  `test_exp`
//	  `abc!`
//	  `d.e.f/gh`
//	  `snake_case?`
//	  `@true`
//	  `@nil`
//
// hex-octets atom type:
//
//	leading '#' followed by <hex-digit>+ to specify the decoded data length
//	followed by '$' to separate length from data
//	encodes an array of octets with each octet described in hexadecimal
//	only hex-digits may appear after '$'
//	no white-space or other characters allowed to simplify encoding and decoding logic
//	octets are encoded as 2 hex digits in sequence, most significant digit first followed by least significant
//	only exact number of 2*length hex digits are expected after '$'
//
//	examples:
//	  `#3$616263`
//	  `#a$0102030405060708090a`
//
// quoted-octets atom type:
//
//	leading '"' and trailing '"'
//	may contain any ASCII and non-ASCII character except '"', '\r', '\n'
//	a '\' is treated as the start of an escape sequence followed by one of:
//		'\' = '\\'
//		'"' = '\"'
//		'r' = '\r'
//		'n' = '\n'
//		't' = '\t'
//		'x' <hex-digit> <hex-digit> = escape of 8-bit character encoded in hexadecimal
//
//	examples:
//	  "abc\ndef\t\"123\"\x00\xff"
//	  "12345"
//
// BNF:
//
//	<sexpr>           :: <nil> | <bool> | <integer> | <token> | <octets> | <list> ;
//
//	<list>            :: "(" ( <sexpr> | <whitespace> )* ")" ;
//
//	<nil>             :: "n" "i" "l" ;
//
//	<bool>            :: <bool-true> | <bool-false> ;
//	<bool-true>       :: "t" "r" "u" "e" ;
//	<bool-false>      :: "f" "a" "l" "s" "e" ;
//
//	<integer>         :: <decimal> | <hexadecimal> ;
//
//	<decimal>         :: <decimal-digit>+ ;
//	<decimal-digit>   :: "0" | ... | "9" ;
//
//	<hexadecimal>     :: "$" <hex-digit>+ ;
//	<hex-digit>       :: <decimal-digit> | "a" | ... | "f" ;
//
//	<octets>          :: <hex-octets> | <quoted-octets> ;
//
//	<hex-octets>      :: "#" <hex-digit>+ "$" <hex-digit>* ;
//
//	<quoted-octets>   :: "\"" ( <quoted-char> | <quoted-escape> )* "\"" ;
//	<quoted-char>     :: [any 8-bit char except "\"", "\\", "\r", "\n"] ;
//	<quoted-escape>   :: "\\" ( <escape-single> | <escape-hex> ) ;
//	<escape-single>   :: "\\" | "\"" | "n" | "r" | "t" ;
//	<escape-hex>      :: "x" <hex-digit> <hex-digit> ;
//
//	<whitespace>      :: <whitespace-char>* ;
//	<whitespace-char> :: " " | "\t" ;
//
//	<token>           :: ( "@" )? <token-start> <token-remainder>* ;
//	<token-start>     :: <alpha> | <simple-punc> | <non-ascii> ;
//	<token-remainder> :: <token-start> | <decimal-digit> ;
//	<alpha>           :: <upper-case> | <lower-case> ;
//	<lower-case>      :: "a" | ... | "z" ;
//	<upper-case>      :: "A" | ... | "Z" ;
//	<simple-punc>     :: "_" | "." | "/" | "?" | "!" ;
//	<non-ascii>       :: [128 <= char <= 255] ;
package brass
