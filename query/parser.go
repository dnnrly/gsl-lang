package query

import (
	"fmt"
	"strconv"
	"strings"
)

// parseQuery is the main entry point for parsing GSL-QL queries using participle
func parseQuery(input string) (*Query, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return &Query{Expressions: []Expression{&IdentityExpr{}}}, nil
	}

	ast, err := queryParser.ParseString("", input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	return convertQuery(ast)
}

// convertQuery converts QueryAST to Query
func convertQuery(ast *QueryAST) (*Query, error) {
	if ast == nil {
		return &Query{Expressions: []Expression{&IdentityExpr{}}}, nil
	}

	expressions := make([]Expression, 0)
	for _, exprAST := range ast.Expressions {
		expr, err := convertExpression(exprAST)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expr)
	}

	if len(expressions) == 0 {
		expressions = append(expressions, &IdentityExpr{})
	}

	return &Query{Expressions: expressions}, nil
}

// convertExpression converts ExpressionAST to Expression
func convertExpression(ast *ExpressionAST) (Expression, error) {
	if ast == nil {
		return &IdentityExpr{}, nil
	}

	if ast.Binding != nil {
		return convertBinding(ast.Binding)
	}
	if ast.Subgraph != nil {
		return convertSubgraph(ast.Subgraph)
	}
	if ast.Make != nil {
		return convertMake(ast.Make)
	}
	if ast.Remove != nil {
		return convertRemove(ast.Remove)
	}
	if ast.Collapse != nil {
		return convertCollapse(ast.Collapse)
	}
	if ast.From != nil {
		return convertFrom(ast.From)
	}
	if ast.GraphAlgebra != nil {
		return convertGraphAlgebra(ast.GraphAlgebra)
	}

	return &IdentityExpr{}, nil
}

// convertBinding converts BindingAST to BindExpr
func convertBinding(ast *BindingAST) (*BindExpr, error) {
	if ast == nil {
		return nil, fmt.Errorf("binding AST is nil")
	}

	// Validate graph name
	if !isValidGraphName(ast.Name) {
		return nil, fmt.Errorf("invalid graph name: %s", ast.Name)
	}

	// Convert the pipeline
	pipeline, err := convertQuery(ast.Pipeline)
	if err != nil {
		return nil, err
	}

	return &BindExpr{
		Pipeline: pipeline,
		Name:     ast.Name,
	}, nil
}

// convertSubgraph converts SubgraphAST to SubgraphExpr
func convertSubgraph(ast *SubgraphAST) (*SubgraphExpr, error) {
	if ast == nil {
		return nil, fmt.Errorf("subgraph AST is nil")
	}

	// Convert predicate
	pred, err := convertPredicate(ast.Predicate)
	if err != nil {
		return nil, err
	}

	var traversal *TraversalConfig
	if ast.Traverse != nil {
		depth := 999999
		if ast.Traverse.Depth != "all" {
			d, err := strconv.Atoi(ast.Traverse.Depth)
			if err != nil {
				return nil, fmt.Errorf("invalid traversal depth: %s", ast.Traverse.Depth)
			}
			// Validate depth is positive
			if d <= 0 {
				return nil, fmt.Errorf("traversal depth must be positive, got %d", d)
			}
			depth = d
		}
		traversal = &TraversalConfig{
			Direction: ast.Traverse.Direction,
			Depth:     depth,
		}
	}

	return &SubgraphExpr{
		Pred:      pred,
		Traversal: traversal,
	}, nil
}

// convertMake converts MakeAST to MakeExpr
func convertMake(ast *MakeAST) (*MakeExpr, error) {
	if ast == nil {
		return nil, fmt.Errorf("make AST is nil")
	}

	// Convert attribute path
	element := ast.Path.Element
	attr := ast.Path.Attr

	// Convert value
	value := convertValue(ast.Value)

	// Convert predicate (where clause)
	pred, err := convertPredicate(ast.Predicate)
	if err != nil {
		return nil, err
	}

	return &MakeExpr{
		Target: element,
		Attr:   attr,
		Value:  value,
		Pred:   pred,
	}, nil
}

// convertRemove converts RemoveAST to Expression
func convertRemove(ast *RemoveAST) (Expression, error) {
	if ast == nil {
		return nil, fmt.Errorf("remove AST is nil")
	}

	// Handle "remove orphans"
	if ast.Orphans {
		return &RemoveOrphansExpr{}, nil
	}

	// Handle "remove edge where predicate"
	if ast.EdgePredicate != nil {
		pred, err := convertPredicate(ast.EdgePredicate)
		if err != nil {
			return nil, err
		}
		return &RemoveEdgeExpr{
			Pred: pred,
		}, nil
	}

	// Handle "remove attr_path" or "remove attr_path where predicate"
	if ast.AttrPath != nil {
		pred, err := convertPredicate(ast.AttrPredicate)
		if err != nil {
			return nil, err
		}
		return &RemoveAttributeExpr{
			Target: ast.AttrPath.Element,
			Attr:   ast.AttrPath.Attr,
			Pred:   pred,
		}, nil
	}

	return nil, fmt.Errorf("invalid remove expression")
}

// convertCollapse converts CollapseAST to CollapseExpr
func convertCollapse(ast *CollapseAST) (*CollapseExpr, error) {
	if ast == nil {
		return nil, fmt.Errorf("collapse AST is nil")
	}

	// Convert predicate
	pred, err := convertPredicate(ast.Predicate)
	if err != nil {
		return nil, err
	}

	return &CollapseExpr{
		NodeID: ast.NodeID,
		Pred:   pred,
	}, nil
}

// convertFrom converts FromAST to FromExpr
func convertFrom(ast *FromAST) (*FromExpr, error) {
	if ast == nil {
		return nil, fmt.Errorf("from AST is nil")
	}

	if ast.Wildcard {
		return &FromExpr{
			IsWildcard: true,
		}, nil
	}

	// Validate graph name
	if !isValidGraphName(ast.Name) {
		return nil, fmt.Errorf("invalid graph name: %s", ast.Name)
	}

	return &FromExpr{
		IsWildcard: false,
		Name:       ast.Name,
	}, nil
}

// convertGraphAlgebra converts GraphAlgebraAST to GraphAlgebraExpr
func convertGraphAlgebra(ast *GraphAlgebraAST) (*GraphAlgebraExpr, error) {
	if ast == nil {
		return nil, fmt.Errorf("graph algebra AST is nil")
	}

	return &GraphAlgebraExpr{
		LeftRef:  ast.Left,
		Operator: ast.Operator,
		RightRef: ast.Right,
	}, nil
}

// convertPredicate converts PredicateAST to Predicate
func convertPredicate(ast *PredicateAST) (Predicate, error) {
	if ast == nil {
		return &ExistsPredicate{}, nil
	}

	if len(ast.Terms) == 0 {
		return &ExistsPredicate{}, nil
	}

	// Convert first term
	pred, err := convertPredicateTerm(ast.Terms[0])
	if err != nil {
		return nil, err
	}

	// Chain remaining terms with AND
	for i := 1; i < len(ast.Terms); i++ {
		right, err := convertPredicateTerm(ast.Terms[i])
		if err != nil {
			return nil, err
		}
		pred = &AndPredicate{Left: pred, Right: right}
	}

	return pred, nil
}

// convertPredicateTerm converts PredicateTermAST to Predicate
func convertPredicateTerm(ast *PredicateTermAST) (Predicate, error) {
	if ast == nil {
		return &ExistsPredicate{}, nil
	}

	if ast.Exists {
		return &ExistsPredicate{}, nil
	}
	// Handle bare set membership (no element type specified)
	if ast.BareSetName != "" {
		return &SetMembershipPredicate{
			Target: "",
			SetID:  ast.BareSetName,
		}, nil
	}
	if ast.BareNotSetName != "" {
		return &SetNotMembershipPredicate{
			Target: "",
			SetID:  ast.BareNotSetName,
		}, nil
	}
	if ast.BareNot != nil {
		innerPred, err := convertPredicate(ast.BareNot)
		if err != nil {
			return nil, err
		}
		return &NotPredicate{Inner: innerPred}, nil
	}
	if ast.SetPredicate != nil {
		return convertSetPredicate(ast.SetPredicate)
	}
	if ast.AttrPredicate != nil {
		return convertAttrPredicate(ast.AttrPredicate)
	}

	return &ExistsPredicate{}, nil
}

// convertAttrPredicate converts AttrPredicateAST to Predicate
func convertAttrPredicate(ast *AttrPredicateAST) (Predicate, error) {
	if ast == nil {
		return nil, fmt.Errorf("attribute predicate AST is nil")
	}

	element := ast.Path.Element
	attr := ast.Path.Attr

	// Handle exists / not exists
	if ast.Exists {
		return &AttributeExistsPredicate{
			Target: element,
			Name:   attr,
		}, nil
	}
	if ast.NotExists {
		return &AttributeNotExistsPredicate{
			Target: element,
			Name:   attr,
		}, nil
	}

	// Handle equality / inequality
	if ast.Operator != "" {
		value := convertValue(ast.Value)
		// Support both "=" and "==" for equality
		if ast.Operator == "=" || ast.Operator == "==" {
			return &AttributeEqualsPredicate{
				Target: element,
				Name:   attr,
				Value:  value,
			}, nil
		}
		if ast.Operator == "!=" {
			return &AttributeNotEqualsPredicate{
				Target: element,
				Name:   attr,
				Value:  value,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid attribute predicate")
}

// convertSetPredicate converts SetPredicateAST to Predicate
func convertSetPredicate(ast *SetPredicateAST) (Predicate, error) {
	if ast == nil {
		return nil, fmt.Errorf("set predicate AST is nil")
	}

	if ast.Not {
		return &SetNotMembershipPredicate{
			Target: ast.Element,
			SetID:  ast.SetName,
		}, nil
	}

	return &SetMembershipPredicate{
		Target: ast.Element,
		SetID:  ast.SetName,
	}, nil
}

// convertValue converts ValueAST to interface{}
func convertValue(v *ValueAST) interface{} {
	if v == nil {
		return ""
	}

	if v.String != nil {
		// Strip surrounding quotes
		s := *v.String
		return s[1 : len(s)-1]
	}

	if v.Bool != nil {
		return bool(*v.Bool)
	}

	if v.Number != nil {
		return *v.Number
	}

	if v.Ident != nil {
		return *v.Ident
	}

	return ""
}

// isValidGraphName validates graph names: [A-Z][A-Z0-9_]*
func isValidGraphName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}
	for i := 1; i < len(name); i++ {
		ch := name[i]
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return false
		}
	}
	return true
}

// isValidGraphRef checks if a reference is valid (either "*" or a named graph name)
func isValidGraphRef(ref string) bool {
	if ref == "*" {
		return true
	}
	// Must match named graph naming: [A-Z][A-Z0-9_]*
	return isValidGraphName(ref)
}

