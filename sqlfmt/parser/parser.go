package parser

import (
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"github.com/noneymous/go-sqlfmt/sqlfmt/reindenters"
)

const joinStartRange = 3

// Parse parses a sequence of tokens returning a logically grouped slice of Reindenter.
// Each Reindenter is a logical segment of an SQL query. It may also be a group of such.
func Parse(tokens []lexer.Token, options *reindenters.Options) ([]reindenters.Reindenter, error) {

	// Check if tokenized string is actually an SQL query
	t := tokens[0].Type
	if !(t == lexer.SELECT || t == lexer.UPDATE || t == lexer.DELETE ||
		t == lexer.INSERT || t == lexer.LOCK || t == lexer.WITH || t == lexer.SET) {
		return nil, fmt.Errorf("invalid sql statement")
	}

	// Prepare parser for segment
	parser, errParser := NewParser(tokens, options)
	if errParser != nil {
		return nil, errParser
	}

	// Parse tokens
	result, errParse := parser.Parse()
	if errParse != nil {
		return nil, errParse
	}

	// Return process result
	return result, nil
}

// Parser initiates with a sequence of lexer tokens to be processed by the Parse() function.
// Furthermore, it holds all necessary state variables and a set of options to be assigned to each Reindenters.
type Parser struct {
	options *reindenters.Options // options used later to format element

	indentLevel int

	tokens   []lexer.Token
	endTypes []lexer.TokenType

	result []reindenters.Reindenter
}

// NewParser initializes a Parser with a sequence of lexer tokens representing an SQL query.
func NewParser(tokens []lexer.Token, options *reindenters.Options) (*Parser, error) {

	// Use default options if none are passed
	if options == nil {
		options = reindenters.DefaultOptions()
	}

	// Create initial Parser with according type, tokens and end types
	firstTokenType := tokens[0].Type
	switch firstTokenType {
	case lexer.SELECT:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfSelect}, nil
	case lexer.FROM:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfFrom}, nil
	case lexer.CASE:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfCase}, nil
	case lexer.JOIN, lexer.INNER, lexer.OUTER, lexer.LEFT, lexer.RIGHT, lexer.NATURAL, lexer.CROSS:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfJoin}, nil
	case lexer.WHERE:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfWhere}, nil
	case lexer.ANDGROUP:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfAndGroup}, nil
	case lexer.ORGROUP:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfOrGroup}, nil
	case lexer.AND:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfAndGroup}, nil
	case lexer.OR:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfOrGroup}, nil
	case lexer.GROUP:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfGroupBy}, nil
	case lexer.HAVING:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfHaving}, nil
	case lexer.ORDER:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfOrderBy}, nil
	case lexer.LIMIT, lexer.FETCH, lexer.OFFSET:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfLimitClause}, nil
	case lexer.STARTPARENTHESIS:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfParenthesis}, nil
	case lexer.UNION, lexer.INTERSECT, lexer.EXCEPT:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfTieClause}, nil
	case lexer.UPDATE:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfUpdate}, nil
	case lexer.SET:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfSet}, nil
	case lexer.RETURNING:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfReturning}, nil
	case lexer.DELETE:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfDelete}, nil
	case lexer.INSERT:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfInsert}, nil
	case lexer.VALUES:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfValues}, nil
	case lexer.FUNCTIONKEYWORD:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfFunctionKeyword}, nil
	case lexer.FUNCTION:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfFunction}, nil
	case lexer.TYPE:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfTypeCast}, nil
	case lexer.LOCK:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfLock}, nil
	case lexer.WITH:
		return &Parser{options: options, tokens: tokens, endTypes: reindenters.EndOfWith}, nil
	default:
		return nil, fmt.Errorf("invalid start token '%s'", tokens[0].Value)
	}
}

// Parse wraps parseSegment() and loops until EOF is reached. The initial sequence of tokens is usually
// comprised out of multiple segments (SELECT, FROM, WHERE,...). Parse() loops until EOF to make sure all
// segments are processed.
func (r *Parser) Parse() ([]reindenters.Reindenter, error) {

	// Prepare process variable
	var offset int

	// Verify that there is an EOF token at the end
	if r.tokens[len(r.tokens)-1].Type != lexer.EOF {
		return nil, fmt.Errorf("missing EOF token")
	}

	// Iterate and process segments until EOF is reached. Sequential segments are processed by this loop.
	// Nested segments are recursively processed by parseSegment().
	for {

		// Stop processing at EOF
		if r.tokens[offset].Type == lexer.EOF {
			break
		}

		// Prepare parser for segment
		segmentParser, errSegmentParser := NewParser(r.tokens[offset:], r.options)
		if errSegmentParser != nil {
			return nil, errSegmentParser
		}

		// Parse segment
		idxEndSegment, errSegment := segmentParser.parseSegment()
		if errSegment != nil {
			return nil, errSegment
		}

		// Retrieve segment result
		segmentReindenter := segmentParser.buildReindenter()

		// Append segment result to total result
		r.result = append(r.result, segmentReindenter)

		// Increment offset counter to proceed with next segment, if available
		offset += idxEndSegment
	}

	// Return process result
	return r.result, nil
}

// parseSegment iterates a token segment, creates according Reindenters and appends them to the final result.
// It iterates until a suitable end token type (depending on the segment's initial token type) could be found.
// Whenever an intermediate subsequence (certain token type) is detected, a new Parser is initialized
// (recursively) processing that subsequence and appending it as a Reindenter group. The same way subsequences
// can be nested arbitrarily, Parsers will be nested equally. A Parser may fork a Parser for a
// subsequence, and so on. The nested Parser results are aggregated by their parent Parser as they are yielded.
func (r *Parser) parseSegment() (int, error) {

	// Prepare process variables
	var (
		idx          int
		tokenCurrent lexer.Token
	)

	// Iterate over all token
	for {

		// Abort if no end token could be found and to prevent out-of-bound panics. Query might not be valid SQL.
		if idx >= len(r.tokens) {
			return 0, fmt.Errorf("could not find end token for token sequence")
		}

		// Get reference of token to analyze
		tokenCurrent = r.tokens[idx]

		// Check for end token or new segment if current token is not first token
		if idx > 0 {

			// Check if token is end of segment
			if r.isEndToken(idx) {
				return idx, nil
			}

			// Check if token introduces query subsegment
			newSegment, indent := r.isNewSegment(idx)
			if newSegment {

				// Create new parser for subsegment
				segmentParser, errSegmentParser := NewParser(r.tokens[idx:], r.options)
				if errSegmentParser != nil {
					return 0, fmt.Errorf("invalid segment: %s", errSegmentParser)
				}

				// Sent indent for Parser
				segmentParser.indentLevel = indent

				// Check if segment parser actually contains a suitable end token
				if !segmentParser.hasEndType() {
					return 0, fmt.Errorf("'%s' segment has no end keyword", tokenCurrent.Value)
				}

				// Parse subsegment
				idxEndSegment, errSegment := segmentParser.parseSegment()
				if errSegment != nil {
					return 0, errSegment
				}

				// Append subsegment result to parent segment result
				segmentReindenter := segmentParser.buildReindenter()
				if segmentReindenter == nil {
					return 0, fmt.Errorf("invalid segment result: %#v", segmentParser.result)
				}

				// Append subsegment result to parent segment result
				r.result = append(r.result, segmentReindenter)

				// Skip tokens that were processed as a subsegment parser
				switch tokenCurrent.Type {
				case lexer.STARTPARENTHESIS, lexer.CASE, lexer.FUNCTION, lexer.TYPE:
					idx += idxEndSegment + 1 // Some types have end tags, e.g. "END" closing "CASE" or ")" closing "(". Next token starts after them.
				default:
					idx += idxEndSegment
				}

				// Continue with next token
				continue
			}
		}

		// Append token to result
		r.result = append(r.result, reindenters.Token{
			Options: r.options,
			Token: lexer.Token{
				Type:  tokenCurrent.Type,
				Value: tokenCurrent.Value,
			},
		})

		// Increase index to continue with next token
		idx++
	}
}

// hasEndType determines if the Parser's token sequence includes a suitable and expected end token type
func (r *Parser) hasEndType() bool {
	for _, token := range r.tokens {
		for _, ttype := range r.endTypes {
			if token.Type == ttype {
				return true
			}
		}
	}
	return false
}

// isEndToken determines if the token at index idx is an end token
func (r *Parser) isEndToken(idx int) bool {

	// Return true if there are no end types defined, meaning that anything is an end type
	if len(r.endTypes) == 0 {
		return true
	}

	// Get tokens to work with
	tokenFirst := r.tokens[0]
	tokenCurrent := r.tokens[idx]

	// Ignore usual token end type if within join clause
	for _, ttype := range reindenters.TokenTypesOfJoinMaker {
		if tokenFirst.Type == ttype && idx < joinStartRange {
			return false
		}
	}

	// Check if token is end token
	for _, tokenEndType := range r.endTypes {
		if tokenCurrent.Type == tokenEndType || tokenCurrent.Type == lexer.EOF {
			return true
		}
	}
	return false
}

// isNewSegment checks whether the token at index idx indicates a new subsegment that needs to be handled
func (r *Parser) isNewSegment(idx int) (bool, int) {

	// Get tokens to work with
	tokenFirst := r.tokens[0]
	tokenCurrent := r.tokens[idx]
	tokenPrevious := r.tokens[idx-1] // isNewSegment() is only called if idx > 0
	tokenNext := r.tokens[idx+1]     // There will always be an EOF token at the end

	//
	// Negative indicators:
	//

	// Not a new segment, if just the parenthesis of a function call
	if tokenCurrent.Type == lexer.STARTPARENTHESIS && tokenPrevious.Type == lexer.FUNCTION {
		return false, 0
	}

	// Not a new segment, if just the parenthesis of a type call
	if tokenCurrent.Type == lexer.STARTPARENTHESIS && tokenPrevious.Type == lexer.TYPE {
		return false, 0
	}

	// Not a new segment, if type call, as indicated by subsequent parenthesis
	if tokenCurrent.Type == lexer.TYPE && tokenNext.Type != lexer.STARTPARENTHESIS {
		return false, 0
	}

	// Not a new segment, if ORDER type contained within sub query
	if tokenCurrent.Type == lexer.ORDER && tokenFirst.Type == lexer.STARTPARENTHESIS {
		return false, 0
	}

	// Not a new segment, if ORDER type contained within FUNCTION subsegment
	if tokenCurrent.Type == lexer.ORDER && tokenFirst.Type == lexer.FUNCTION {
		return false, 0
	}

	// Not a new segment, if FROM type contained within a FUNCTION segment
	if tokenCurrent.Type == lexer.FROM && tokenFirst.Type == lexer.FUNCTION {
		return false, 0
	}

	//
	// Positive indicators:
	//

	// Check if token is indicating join clause.
	// However, ignore keyword if it appears in start of join group such as LEFT INNER JOIN.
	// In this case, "INNER" and "JOIN" are group keyword, but should not make subGroup
	for _, ttype := range reindenters.TokenTypesOfJoinMaker {
		if tokenCurrent.Type == ttype && idx >= joinStartRange {
			return true, r.indentLevel
		}
	}

	// Check if token is of any type commonly indicating subsegment
	for _, v := range reindenters.TokenTypesOfGroupMaker {
		if tokenCurrent.Type == v {
			return true, r.indentLevel
		}
	}

	// Return false as a fallback if no case for subsegment could be made
	return false, 0
}

// buildReindenter creates a Reindenter for the intermediate Parser subsegment, representing a
// segment of the SQL query, which can then be appended to the result sequence
func (r *Parser) buildReindenter() reindenters.Reindenter {

	// Get variables to work with
	elements := r.result
	firstElement, _ := elements[0].(reindenters.Token)
	identLevel := r.indentLevel

	// Build suitable Reintender group and return it
	switch firstElement.Type {
	case lexer.SELECT:
		return &reindenters.Select{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.FROM:
		return &reindenters.From{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.JOIN, lexer.INNER, lexer.OUTER, lexer.LEFT, lexer.RIGHT, lexer.NATURAL, lexer.CROSS:
		return &reindenters.Join{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.WHERE:
		return &reindenters.Where{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.ANDGROUP:
		return &reindenters.AndGroup{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.ORGROUP:
		return &reindenters.OrGroup{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.AND:
		return &reindenters.AndGroup{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.OR:
		return &reindenters.OrGroup{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.GROUP:
		return &reindenters.GroupBy{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.ORDER:
		return &reindenters.OrderBy{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.HAVING:
		return &reindenters.Having{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.LIMIT, lexer.OFFSET, lexer.FETCH:
		return &reindenters.Limit{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.UNION, lexer.INTERSECT, lexer.EXCEPT:
		return &reindenters.TieGroup{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.UPDATE:
		return &reindenters.Update{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.SET:
		return &reindenters.Set{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.RETURNING:
		return &reindenters.Returning{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.LOCK:
		return &reindenters.Lock{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.INSERT:
		return &reindenters.Insert{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.VALUES:
		return &reindenters.Values{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.DELETE:
		return &reindenters.Delete{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.WITH:
		return &reindenters.With{Options: r.options, Element: elements, IndentLevel: identLevel}
	case lexer.CASE:

		// End token of CASE group("END") has to be added to the group
		endToken := reindenters.Token{Options: r.options, Token: lexer.Token{Type: lexer.END, Value: "END"}}
		elements = append(elements, endToken)
		return &reindenters.Case{Options: r.options, Element: elements, IndentLevel: identLevel}

	case lexer.STARTPARENTHESIS:

		// End token of sub query group (")") has to be added in the group
		endToken := reindenters.Token{Options: r.options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}}
		elements = append(elements, endToken)

		// Create subquery indenter if first keyword is SELECT or related keyword. Subqueries are not a
		// lot different to parenthesis groups, but this gives us additional information and format control
		switch elements[1].(type) {
		case *reindenters.Select:
			return &reindenters.Subquery{Options: r.options, Element: elements, IndentLevel: identLevel}
		case *reindenters.With:
			return &reindenters.Subquery{Options: r.options, Element: elements, IndentLevel: identLevel}
		}

		// Return normal parenthesis group otherwise
		return &reindenters.Parenthesis{Options: r.options, Element: elements, IndentLevel: identLevel}

	case lexer.FUNCTION:

		// End token of function group (")") has to be added in the group
		endToken := reindenters.Token{Options: r.options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}}
		elements = append(elements, endToken)
		return &reindenters.Function{Options: r.options, Element: elements, IndentLevel: identLevel}

	case lexer.TYPE:

		// End token of TYPE group (")") has to be added in the group
		endToken := reindenters.Token{Options: r.options, Token: lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}}
		elements = append(elements, endToken)
		return &reindenters.TypeCast{Options: r.options, Element: elements, IndentLevel: identLevel}
	}

	// Return nil as no group could be built
	return nil
}
