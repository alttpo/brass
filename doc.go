/*
Package brass

brass is a custom s-expressions encoding.

	author: jsd1982
	date:   2023-01-07

a brass s-expression is recursively either an atom expression or a list of s-expressions surrounded by '(' ')'.

encoding restrictions:
 1. newline whitespace characters ('\r', '\n') MUST NOT appear in an encoded s-expression.
 2. encoding is ASCII and 8-bit clean with exception of '\r' and '\n'.
 3. each atom has a single encoding with no stylistic variation so as to simplify encoding and decoding logic.

s-expression examples:

	("test_exp" "abc" "d.e.f/gh" nil true false #3$616263 $1000)
	(#a$0102030405060708090a $3ff -$7f)
	("abc\ndef\t\"123\"\x00\xff" () "12345")
	{("a" 1) ("b" 2) ("c" 3) ("d" nil) ("e" false)}

atom types:

	nil
	bool
	integer
	octets
	string
	list
	map

nil atom type:

	literal "nil" keyword

bool atom type:

	literal "true" and "false" keywords

integer atom type:

	a base-16 signed integer value of at most 52-bit length
	may start with optional '-' to indicate negative value
	no extra formatting-related ('_'), division (','), or white-space characters are allowed
	any number of leading zeros are allowed and *do not* signify base-8
	must start with '$' (after optional '-')
	must contain only digits '0'..'9','a'..'f'
	cannot contain upper-case 'A'..'F' so as to simplify encoding and decoding logic

	examples:
		 `$3ff`  =   ( 1023)
		`-$3ff`  =   (-1023)

octets atom type:

	leading '#' followed by <hex-digit>+ to specify the decoded data length
	followed by '$' and then <hex-digit>* to separate length from data
	length records number of octets encoded, not the number of hex-digits
	if size is zero then must end in only '$' with no hex-digits after
	only hex-digits may appear after '$'
	only exact number of 2*length hex digits are expected after '$'
	octets are encoded as 2 hex digits in sequence, most significant digit first followed by least significant
	no white-space or other characters allowed to simplify encoding and decoding logic

	examples:
	  `#3$616263`
	  `#a$0102030405060708090a`

string atom type:

	required leading '"' and trailing '"'
	may directly contain any 7-bit ASCII and 8-bit non-ASCII character except '"', '\r', '\n'
	a '\' is treated as the start of an escape sequence followed by one of:
		'\' = '\\'
		'"' = '\"'
		'r' = '\r'
		'n' = '\n'
		't' = '\t'
		'x' <hex-digit> <hex-digit> = escape of 8-bit character encoded in hexadecimal

	examples:
	  "abc\ndef\t\"123\"\x00\xff"
	  "12345"

list atom type:

	list of s-expressions surrounded by '(' ')'
	each s-expression separated by whitespace

map atom type:

	map of key-value pairs surrounded by '{' '}'
	each key-value pair separated by whitespace
	key-value pairs are represented as two-element lists surrounded by '(' ')'
	key and value are separated by whitespace
	keys can only be a primitive atom type (not list or map)
	values can be any atom type

	examples:
	  {("a" $1) ("b" $2)}
	  {(#1$0a $1) (#1$0b $2)}

BNF:

	<sexpr>           :: <sexpr-primitive> | <sexpr-complex> ;
	<sexpr-primitive> :: <nil> | <bool> | <integer> | <string> | <octets> ;
	<sexpr-complex>   :: <list> | <map> ;

	<list>            :: '(' ( <sexpr> | <whitespace> )* ')' ;

	<map>             :: '{' ( <map-entry> | <whitespace> )* '}' ;
	<map-entry>       :: '(' <sexpr-primitive> <whitespace> <sexpr> ')' ;

	<nil>             :: 'n' 'i' 'l' ;

	<bool>            :: <bool-true> | <bool-false> ;
	<bool-true>       :: 't' 'r' 'u' 'e' ;
	<bool-false>      :: 'f' 'a' 'l' 's' 'e' ;

	<integer>         :: ( '-' )? <hexadecimal> ;

	<hexadecimal>     :: '$' <hex-digit>+ ;
	<hex-digit>       :: '0' | ... | '9' | 'a' | ... | 'f' ;

	<octets>          :: '#' <hex-digit>+ '$' <hex-digit>* ;

	<string>          :: '\"' ( <quoted-char> | <quoted-escape> )* '\"' ;
	<quoted-char>     :: [any 8-bit char except '\"', '\\', '\r', '\n'] ;
	<quoted-escape>   :: '\\' ( <escape-single> | <escape-hex> ) ;
	<escape-single>   :: '\\' | '\"' | 'n' | 'r' | 't' ;
	<escape-hex>      :: 'x' <hex-digit> <hex-digit> ;

	<whitespace>      :: ' ' ;
*/
package brass
