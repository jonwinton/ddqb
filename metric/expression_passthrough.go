package metric

import (
	"fmt"

	"github.com/jonwinton/ddqp"
)

// expressionQueryBuilder enables limited editing of complex metric expressions.
// Currently supports adding filters which are applied to all metric queries
// within the expression. Other mutators are no-ops.
type expressionQueryBuilder struct {
	original     string
	addedFilters []FilterExpression
}

func newExpressionPassthroughBuilder(original string) QueryBuilder { // keep constructor name for minimal diff
	return &expressionQueryBuilder{original: original, addedFilters: []FilterExpression{}}
}

func (b *expressionQueryBuilder) Metric(_ string) QueryBuilder     { return b }
func (b *expressionQueryBuilder) Aggregator(_ string) QueryBuilder { return b }
func (b *expressionQueryBuilder) Filter(filter FilterExpression) QueryBuilder {
	b.addedFilters = append(b.addedFilters, filter)
	return b
}
func (b *expressionQueryBuilder) GetFilters() []FilterExpression { return nil }
func (b *expressionQueryBuilder) FindGroup(_ func(FilterGroupBuilder) bool) FilterGroupBuilder {
	return nil
}

func (b *expressionQueryBuilder) AddToGroup(_ FilterGroupBuilder, _ FilterExpression) QueryBuilder {
	// Not supported for expressions yet
	return b
}
func (b *expressionQueryBuilder) GroupBy(_ ...string) QueryBuilder             { return b }
func (b *expressionQueryBuilder) ApplyFunction(_ FunctionBuilder) QueryBuilder { return b }
func (b *expressionQueryBuilder) TimeWindow(_ string) QueryBuilder             { return b }

func (b *expressionQueryBuilder) Build() (string, error) {
	if len(b.addedFilters) == 0 {
		return b.original, nil
	}

	gp := ddqp.NewGenericParser()
	parsed, err := gp.Parse(b.original)
	if err != nil {
		return "", fmt.Errorf("failed to parse expression for editing: %w", err)
	}

	// Prepare params for all added filters
	params, err := buildParamsForFilters(b.addedFilters)
	if err != nil {
		return "", err
	}

	if parsed.MetricQuery != nil {
		if err := applyFiltersToMetricQuery(parsed.MetricQuery, params); err != nil {
			return "", err
		}
		return parsed.MetricQuery.String(), nil
	}

	if parsed.MetricExpression != nil {
		if err := applyFiltersToMetricExpression(parsed.MetricExpression, params); err != nil {
			return "", err
		}
		return parsed.MetricExpression.String(), nil
	}

	return b.original, nil
}

// buildParamsForFilters converts our FilterExpression list into ddqp.Param slices,
// including leading comma separators between appended filters.
func buildParamsForFilters(filters []FilterExpression) ([]*ddqp.Param, error) {
	out := []*ddqp.Param{}
	for _, fe := range filters {
		// Always separate with a comma from existing filters
		out = append(out, &ddqp.Param{Separator: &ddqp.FilterValueSeparator{Comma: true}})

		// Special-case negated groups to inject NOT
		if g, ok := fe.(*filterGroupBuilder); ok && g.negated {
			out = append(out, &ddqp.Param{Separator: &ddqp.FilterValueSeparator{Not: true}})
		}

		p, err := toDDQPParam(fe)
		if err != nil {
			return nil, err
		}
		if p != nil {
			out = append(out, p)
		}
	}
	return out, nil
}

func toDDQPParam(expr FilterExpression) (*ddqp.Param, error) {
	switch e := expr.(type) {
	case *filterBuilder:
		if e.key == "" {
			return nil, fmt.Errorf("filter key is required")
		}
		sf := &ddqp.SimpleFilter{FilterKey: e.key, FilterSeparator: &ddqp.FilterSeparator{}, FilterValue: &ddqp.FilterValue{}}
		switch e.operation {
		case Equal:
			sf.FilterSeparator.Colon = true
			sf.FilterValue.SimpleValue = &ddqp.Value{Identifier: &e.values[0]}
		case NotEqual:
			sf.Negative = true
			sf.FilterSeparator.Colon = true
			sf.FilterValue.SimpleValue = &ddqp.Value{Identifier: &e.values[0]}
		case Regex:
			sf.FilterSeparator.Regex = true
			sf.FilterValue.SimpleValue = &ddqp.Value{Identifier: &e.values[0]}
		case In, NotIn:
			if e.operation == In {
				sf.FilterSeparator.In = true
			} else {
				sf.FilterSeparator.NotIn = true
			}
			list := []*ddqp.Value{}
			for i, v := range e.values {
				// value
				val := v // ensure distinct address
				list = append(list, &ddqp.Value{Identifier: &val})
				// comma between values except after last
				if i < len(e.values)-1 {
					list = append(list, &ddqp.Value{Separator: &ddqp.FilterValueSeparator{Comma: true}})
				}
			}
			sf.FilterValue.ListValue = list
		default:
			return nil, fmt.Errorf("unknown filter operation")
		}
		return &ddqp.Param{SimpleFilter: sf}, nil

	case *filterGroupBuilder:
		// Build grouped filter recursively
		gf := &ddqp.GroupedFilter{Parameters: []*ddqp.Param{}}

		for idx, sub := range e.expressions {
			if idx > 0 {
				// Insert group operator separator between sub-expressions
				sep := &ddqp.FilterValueSeparator{}
				if e.operator == AndOperator {
					sep.And = true
				} else {
					sep.Or = true
				}
				gf.Parameters = append(gf.Parameters, &ddqp.Param{Separator: sep})
			}
			p, err := toDDQPParam(sub)
			if err != nil {
				return nil, err
			}
			gf.Parameters = append(gf.Parameters, p)
		}
		return &ddqp.Param{GroupedFilter: gf}, nil

	default:
		// Unknown expression type
		return nil, fmt.Errorf("unsupported filter expression type")
	}
}

func applyFiltersToMetricExpression(expr *ddqp.MetricExpression, params []*ddqp.Param) error {
	if expr == nil || expr.GroupedExpression == nil {
		return nil
	}
	return applyFiltersToGroupedExpression(expr.GroupedExpression, params)
}

func applyFiltersToGroupedExpression(ge *ddqp.GroupedExpression, params []*ddqp.Param) error {
	if ge == nil || ge.Left == nil {
		return nil
	}
	if err := applyFiltersToTerm(ge.Left, params); err != nil {
		return err
	}
	for _, rt := range ge.Right {
		if rt != nil && rt.Term != nil {
			if err := applyFiltersToTerm(rt.Term, params); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyFiltersToTerm(t *ddqp.Term, params []*ddqp.Param) error {
	if t == nil || t.Left == nil || t.Left.Base == nil {
		return nil
	}
	if err := applyFiltersToExprValue(t.Left.Base, params); err != nil {
		return err
	}
	for _, of := range t.Right {
		if of != nil && of.Factor != nil && of.Factor.Base != nil {
			if err := applyFiltersToExprValue(of.Factor.Base, params); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyFiltersToExprValue(v *ddqp.ExprValue, params []*ddqp.Param) error {
	if v.Subexpression != nil {
		return applyFiltersToMetricExpression(v.Subexpression, params)
	}
	if v.MetricQuery != nil {
		return applyFiltersToMetricQuery(v.MetricQuery, params)
	}
	if v.ExprAggregatorFuction != nil && v.ExprAggregatorFuction.Body != nil {
		return applyFiltersToGroupedExpression(v.ExprAggregatorFuction.Body, params)
	}
	return nil
}

func applyFiltersToMetricQuery(mq *ddqp.MetricQuery, params []*ddqp.Param) error {
	if mq == nil {
		return nil
	}
    if mq.Query != nil {
		q := mq.Query
		if q.Filters == nil {
			q.Filters = &ddqp.MetricFilter{Left: &ddqp.Param{Asterisk: true}}
		}
        q.Filters.Parameters = append(q.Filters.Parameters, params...)
        if hasExplicitOpsAndComma(q.Filters) {
            normalizeMetricFilterToExplicit(q.Filters)
        }
		return nil
	}
	if mq.AggregatorFuction != nil && mq.AggregatorFuction.Body != nil {
		return applyFiltersToMetricQuery(mq.AggregatorFuction.Body, params)
	}
	return nil
}

// normalizeMetricFilterToExplicit converts comma separators to AND and moves any
// simple filter negatives (!) to NOT separators. It also rewrites the entire
// filter into a single grouped filter to allow a leading NOT.
func normalizeMetricFilterToExplicit(mf *ddqp.MetricFilter) {
    if mf == nil || mf.Left == nil {
        return
    }

    gf := &ddqp.GroupedFilter{Parameters: []*ddqp.Param{}}

    // Helper to append a NOT before a simple filter if it was negated
    appendParamWithNotIfNeeded := func(p *ddqp.Param) {
        if p.SimpleFilter != nil && p.SimpleFilter.Negative {
            p.SimpleFilter.Negative = false
            gf.Parameters = append(gf.Parameters, &ddqp.Param{Separator: &ddqp.FilterValueSeparator{Not: true}})
        }
        gf.Parameters = append(gf.Parameters, p)
    }

    // Process Left
    left := cloneParam(mf.Left)
    normalizeParam(left)
    appendParamWithNotIfNeeded(left)

    // Process Parameters
    for _, p := range mf.Parameters {
        np := cloneParam(p)
        // Convert commas to AND
        if np.Separator != nil && np.Separator.Comma {
            np.Separator.Comma = false
            np.Separator.And = true
        }
        // Keep other separators as-is (AND/OR/NOT variants)
        // If this element is a negated simple filter, move negation to NOT separator
        if np.SimpleFilter != nil && np.SimpleFilter.Negative {
            np.SimpleFilter.Negative = false
            gf.Parameters = append(gf.Parameters, &ddqp.Param{Separator: &ddqp.FilterValueSeparator{Not: true}})
        }
        normalizeParam(np)
        gf.Parameters = append(gf.Parameters, np)
    }

    // Rewrite mf to a single grouped filter
    mf.Left = &ddqp.Param{GroupedFilter: gf}
    mf.Parameters = nil
}

// hasExplicitOpsAndComma returns true if the filter contains both any explicit
// boolean separators (AND/OR/NOT variants) and any comma separators.
func hasExplicitOpsAndComma(mf *ddqp.MetricFilter) bool {
    if mf == nil {
        return false
    }
    hasExplicit := false
    hasComma := false

    var scanParam func(p *ddqp.Param)
    scanParam = func(p *ddqp.Param) {
        if p == nil {
            return
        }
        if p.Separator != nil {
            if p.Separator.Comma {
                hasComma = true
            }
            if p.Separator.And || p.Separator.Or || p.Separator.AndNot || p.Separator.OrNot || p.Separator.Not {
                hasExplicit = true
            }
        }
        if p.GroupedFilter != nil {
            for _, sp := range p.GroupedFilter.Parameters {
                scanParam(sp)
            }
        }
    }

    scanParam(mf.Left)
    for _, p := range mf.Parameters {
        scanParam(p)
    }

    return hasExplicit && hasComma
}

func normalizeParam(p *ddqp.Param) {
    if p == nil {
        return
    }
    if p.GroupedFilter != nil {
        normalizeGroupedFilter(p.GroupedFilter)
    }
    if p.SimpleFilter != nil {
        // value stays; handled in placement to insert NOT when needed
        // nothing else to do here
        return
    }
}

func normalizeGroupedFilter(gf *ddqp.GroupedFilter) {
    if gf == nil {
        return
    }
    params := []*ddqp.Param{}

    // First element may need leading NOT if negated simple filter
    if len(gf.Parameters) > 0 {
        first := cloneParam(gf.Parameters[0])
        if first.SimpleFilter != nil && first.SimpleFilter.Negative {
            first.SimpleFilter.Negative = false
            params = append(params, &ddqp.Param{Separator: &ddqp.FilterValueSeparator{Not: true}})
        }
        normalizeParam(first)
        params = append(params, first)
    }

    // Remaining elements: convert commas to AND, normalize recursively
    for i := 1; i < len(gf.Parameters); i++ {
        np := cloneParam(gf.Parameters[i])
        if np.Separator != nil && np.Separator.Comma {
            np.Separator.Comma = false
            np.Separator.And = true
        }
        // If this element is a negated simple filter and separator isn't a NOT variant, insert NOT
        if np.SimpleFilter != nil && np.SimpleFilter.Negative {
            np.SimpleFilter.Negative = false
            // Insert NOT separator before the filter
            params = append(params, &ddqp.Param{Separator: &ddqp.FilterValueSeparator{Not: true}})
        }
        normalizeParam(np)
        params = append(params, np)
    }

    gf.Parameters = params
}

// cloneParam performs a shallow clone suitable for safe in-place normalization
func cloneParam(p *ddqp.Param) *ddqp.Param {
    if p == nil {
        return nil
    }
    cp := &ddqp.Param{}
    if p.Separator != nil {
        s := *p.Separator
        cp.Separator = &s
    }
    if p.GroupedFilter != nil {
        // deep-ish clone for nested structure
        ng := &ddqp.GroupedFilter{Parameters: []*ddqp.Param{}}
        for _, sub := range p.GroupedFilter.Parameters {
            ng.Parameters = append(ng.Parameters, cloneParam(sub))
        }
        cp.GroupedFilter = ng
    }
    if p.SimpleFilter != nil {
        sf := *p.SimpleFilter
        // FilterValue and FilterSeparator can be reused safely as we only mutate booleans
        cp.SimpleFilter = &sf
    }
    if p.Asterisk {
        cp.Asterisk = true
    }
    return cp
}
