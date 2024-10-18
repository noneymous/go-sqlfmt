package parser

import (
	"fmt"
	"github.com/noneymous/go-sqlfmt/sqlfmt/lexer"
	"github.com/noneymous/go-sqlfmt/sqlfmt/parser/group"
)

const joinStartRange = 3

// Retriever initiates with a sequence of lexer tokens and groups them logically in the processing. Calls itself
// recursively if query subsegments are detected. Those subsegments are grouped and added to the final result.
type Retriever struct {
	indentLevel int

	tokens   []lexer.Token
	endTypes []lexer.TokenType

	result []group.Reindenter
}

// NewRetriever initializes a Retriever with a given sequence of lexer tokens representing an SQL query.
func NewRetriever(tokens []lexer.Token) (*Retriever, error) {
	typeFirstToken := tokens[0].Type
	switch typeFirstToken {
	case lexer.SELECT:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfSelect}, nil
	case lexer.FROM:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfFrom}, nil
	case lexer.CASE:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfCase}, nil
	case lexer.JOIN, lexer.INNER, lexer.OUTER, lexer.LEFT, lexer.RIGHT, lexer.NATURAL, lexer.CROSS:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfJoin}, nil
	case lexer.WHERE:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfWhere}, nil
	case lexer.ANDGROUP:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfAndGroup}, nil
	case lexer.ORGROUP:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfOrGroup}, nil
	case lexer.AND:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfAndGroup}, nil
	case lexer.OR:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfOrGroup}, nil
	case lexer.GROUP:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfGroupBy}, nil
	case lexer.HAVING:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfHaving}, nil
	case lexer.ORDER:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfOrderBy}, nil
	case lexer.LIMIT, lexer.FETCH, lexer.OFFSET:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfLimitClause}, nil
	case lexer.STARTPARENTHESIS:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfParenthesis}, nil
	case lexer.UNION, lexer.INTERSECT, lexer.EXCEPT:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfTieClause}, nil
	case lexer.UPDATE:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfUpdate}, nil
	case lexer.SET:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfSet}, nil
	case lexer.RETURNING:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfReturning}, nil
	case lexer.DELETE:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfDelete}, nil
	case lexer.INSERT:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfInsert}, nil
	case lexer.VALUES:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfValues}, nil
	case lexer.FUNCTION:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfFunction}, nil
	case lexer.TYPE:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfTypeCast}, nil
	case lexer.LOCK:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfLock}, nil
	case lexer.WITH:
		return &Retriever{tokens: tokens, endTypes: lexer.EndOfWith}, nil
	default:
		return nil, fmt.Errorf("invalid start token '%s'", tokens[0].Value)
	}
}

// Process wraps processSegment() and loops until EOF is reached. There might be multiple segments in a sequence, rather than
// nested. Without looping til EOF only the first segment would be processed.
// Furthermore, Process drops the end index from the return values, because it's only needed in recusrive calls.
func (r *Retriever) Process() ([]group.Reindenter, error) {

	// Prepare process variable
	var offset int

	// Iterate and process segments until EOF is reached. Sequential segments are processed by this loop.
	// Nested segments are recursively processed by processSegment().
	for {

		// Stop processing at EOF
		if r.tokens[offset].Type == lexer.EOF {
			break
		}

		// Prepare retriever for segment
		segmentRetriever, errSegmentRetriever := NewRetriever(r.tokens[offset:])
		if errSegmentRetriever != nil {
			return nil, errSegmentRetriever
		}

		// Process segment
		idxEndSegment, errSegment := segmentRetriever.processSegment()
		if errSegment != nil {
			return nil, errSegment
		}

		// Retrieve segment result
		segmentResult := segmentRetriever.buildSegmentResult()

		// Append segment result to total result
		r.result = append(r.result, segmentResult)

		// Increment offset counter to proceed with next segment, if available
		offset += idxEndSegment
	}

	// Return process result
	return r.result, nil
}

// processSegment iterates a token sequence and appends each token to the result.
// It iterates until a suitable endTokenType (depending on the segment's initial token) could be found.
// Whenever an intermediate subsequence (indicated by certain tokens) appears, a new Retriever is initialized
// (recursively) processing that subsequence and appending it as a result group. The same way subsequences
// can be nested  arbitrarily, Retrievers are be nested equally. A Retriever may fork a Retriever for a
// subsequence, and so on. The nested Retriever results are aggregated subsequently as they yield subsequence
// results.
func (r *Retriever) processSegment() (int, error) {

	// Prepare process variables
	var (
		idx          int
		tokenCurrent lexer.Token
	)

	// Iterate over all token
	for {

		// Abort if no end token could be found and to prevent out-of-bound panics
		// TODO: Maybe this can be improved to give more information about the actual faulty token segement
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

				// Create new retriever for subsegment
				segmentRetriever, errSegmentRetriever := NewRetriever(r.tokens[idx:])
				if errSegmentRetriever != nil {
					return 0, fmt.Errorf("invalid segment: %s", errSegmentRetriever)
				}

				// Sent indent for Retriever
				segmentRetriever.indentLevel = indent

				// Check if segment retriever actually contains a suitable end token
				if !segmentRetriever.hasEndType() {
					return 0, fmt.Errorf("'%s' segment has no end keyword", tokenCurrent.Value)
				}

				// Process subsegment
				idxEndSegment, errSegment := segmentRetriever.processSegment()
				if errSegment != nil {
					return 0, errSegment
				}

				// Append subsegment result to parent segment result
				segmentResult := segmentRetriever.buildSegmentResult()
				if segmentResult == nil {
					return 0, fmt.Errorf("invalid group result: %#v", segmentRetriever.result)
				}

				// Append subsegment result to parent segment result
				r.result = append(r.result, segmentResult)

				// Skip tokens that were processed as a subsegment retriever
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
		r.result = append(r.result, tokenCurrent)

		// Increase index to continue with next token
		idx++
	}
}

// isEndToken determines if the token at index idx is an end token
func (r *Retriever) isEndToken(idx int) bool {

	// Get tokens to work with
	tokenFirst := r.tokens[0]
	tokenCurrent := r.tokens[idx]

	// Ignore token end type if it appears in start of join clause such as LEFT OUTER JOIN, INNER JOIN etc ...
	if idx < joinStartRange {
		for _, ttype := range lexer.TokenTypesOfJoinMaker {
			if tokenFirst.Type == ttype {
				return false
			}
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

// hasEndType determines if the Retriever's token sequence includes a suitable and expected end token type
func (r *Retriever) hasEndType() bool {
	for _, token := range r.tokens {
		for _, ttype := range r.endTypes {
			if token.Type == ttype {
				return true
			}
		}
	}
	return false
}

// isNewSegment checks whether the token at index idx indicates a new subsegment that needs to be handled
func (r *Retriever) isNewSegment(idx int) (bool, int) {

	// Get tokens to work with
	tokenFirst := r.tokens[0]
	tokenCurrent := r.tokens[idx]
	tokenNext := r.tokens[idx+1]

	// Check if token is a misleading group marker, such as
	// 		- an ORDER BY in a function group
	// 		- a type cast that is a function
	// 		- an ORDER BY clause in window function
	// 		- an opening parenthesis which belongs to a function, rather than a sub query
	// 		- a type cast that is not a function
	if tokenFirst.Type == lexer.FUNCTION && tokenCurrent.Type == lexer.ORDER { // in order not to make ORDER BY subGroup in Function group. This is a solution of window function
		return false, 0
	} else if tokenFirst.Type == lexer.TYPE && tokenCurrent.Type == lexer.STARTPARENTHESIS { // in order to ignore "(" in TypeCast group
		return false, 0
	} else if tokenFirst.Type == lexer.STARTPARENTHESIS && tokenCurrent.Type == lexer.ORDER { // in order to ignore ORDER BY in window function
		return false, 0
	} else if tokenFirst.Type == lexer.FUNCTION && (tokenCurrent.Type == lexer.STARTPARENTHESIS || tokenCurrent.Type == lexer.FROM) {
		return false, 0
	} else if tokenCurrent.Type == lexer.TYPE && !(tokenNext.Type == lexer.STARTPARENTHESIS) {
		return false, 0
	}

	// Check if token is parenthesis opening sub query, indicating subsegment
	if tokenCurrent.Type == lexer.STARTPARENTHESIS && tokenNext.Type == lexer.SELECT {
		return true, r.indentLevel + 1
	}

	// Check if token is JOIN type, indicating subsegment
	if tokenCurrent.IsJoinStart() {

		// Ignore keyword if it appears in start of join group such as LEFT INNER JOIN.
		// In this case, "INNER" and "JOIN" are group keyword, but should not make subGroup.
		if idx >= joinStartRange {
			return true, r.indentLevel
		}
	}

	// Check if token is of any type commonly indicating subsegment
	for _, v := range lexer.TokenTypesOfGroupMaker {
		if tokenCurrent.Type == v {
			return true, r.indentLevel
		}
	}

	// Return false if no case for subsegment could be made
	return false, 0
}

// buildSegmentResult creates a Reindenter segment group from the intermediate Retriever subsegment, which can
// then be appended to the result sequence
func (r *Retriever) buildSegmentResult() group.Reindenter {

	// Get variables to work with
	elements := r.result
	firstElement, _ := elements[0].(lexer.Token)
	identLevel := r.indentLevel

	// Build suitable Reintender group and return it
	switch firstElement.Type {
	case lexer.SELECT:
		return &group.Select{Element: elements, IndentLevel: identLevel}
	case lexer.FROM:
		return &group.From{Element: elements, IndentLevel: identLevel}
	case lexer.JOIN, lexer.INNER, lexer.OUTER, lexer.LEFT, lexer.RIGHT, lexer.NATURAL, lexer.CROSS:
		return &group.Join{Element: elements, IndentLevel: identLevel}
	case lexer.WHERE:
		return &group.Where{Element: elements, IndentLevel: identLevel}
	case lexer.ANDGROUP:
		return &group.AndGroup{Element: elements, IndentLevel: identLevel}
	case lexer.ORGROUP:
		return &group.OrGroup{Element: elements, IndentLevel: identLevel}
	case lexer.AND:
		return &group.AndGroup{Element: elements, IndentLevel: identLevel}
	case lexer.OR:
		return &group.OrGroup{Element: elements, IndentLevel: identLevel}
	case lexer.GROUP:
		return &group.GroupBy{Element: elements, IndentLevel: identLevel}
	case lexer.ORDER:
		return &group.OrderBy{Element: elements, IndentLevel: identLevel}
	case lexer.HAVING:
		return &group.Having{Element: elements, IndentLevel: identLevel}
	case lexer.LIMIT, lexer.OFFSET, lexer.FETCH:
		return &group.LimitClause{Element: elements, IndentLevel: identLevel}
	case lexer.UNION, lexer.INTERSECT, lexer.EXCEPT:
		return &group.TieClause{Element: elements, IndentLevel: identLevel}
	case lexer.UPDATE:
		return &group.Update{Element: elements, IndentLevel: identLevel}
	case lexer.SET:
		return &group.Set{Element: elements, IndentLevel: identLevel}
	case lexer.RETURNING:
		return &group.Returning{Element: elements, IndentLevel: identLevel}
	case lexer.LOCK:
		return &group.Lock{Element: elements, IndentLevel: identLevel}
	case lexer.INSERT:
		return &group.Insert{Element: elements, IndentLevel: identLevel}
	case lexer.VALUES:
		return &group.Values{Element: elements, IndentLevel: identLevel}
	case lexer.DELETE:
		return &group.Delete{Element: elements, IndentLevel: identLevel}
	case lexer.WITH:
		return &group.With{Element: elements, IndentLevel: identLevel}
	case lexer.CASE:

		// End token of CASE group("END") has to be added to the group
		endToken := lexer.Token{Type: lexer.END, Value: "END"}
		elements = append(elements, endToken)
		return &group.Case{Element: elements, IndentLevel: identLevel}

	case lexer.STARTPARENTHESIS:

		// End token of sub query group (")") has to be added in the group
		endToken := lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}
		elements = append(elements, endToken)
		if _, isSubQuery := elements[1].(*group.Select); isSubQuery {
			return &group.Subquery{Element: elements, IndentLevel: identLevel}
		}
		return &group.Parenthesis{Element: elements, IndentLevel: identLevel}

	case lexer.FUNCTION:

		// End token of function group (")") has to be added in the group
		endToken := lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}
		elements = append(elements, endToken)
		return &group.Function{Element: elements, IndentLevel: identLevel}

	case lexer.TYPE:

		// End token of TYPE group (")") has to be added in the group
		endToken := lexer.Token{Type: lexer.ENDPARENTHESIS, Value: ")"}
		elements = append(elements, endToken)
		return &group.TypeCast{Element: elements, IndentLevel: identLevel}
	}

	// Return nil as no group could be built
	return nil
}
