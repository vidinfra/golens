package filter

type Clause string

const (
	Equals          Clause = "eq"
	NotEquals       Clause = "ne"
	Contains        Clause = "like"
	NotContains     Clause = "not-like"
	StartsWith      Clause = "starts-with"
	EndsWith        Clause = "ends-with"
	GreaterThan     Clause = "gt"
	GreaterThanOrEq Clause = "gte"
	LessThan        Clause = "lt"
	LessThanOrEq    Clause = "lte"
	In              Clause = "in"
	NotIn           Clause = "not-in"
	IsNull          Clause = "null"
	IsNotNull       Clause = "not-null"
	Between         Clause = "between"
	NotBetween      Clause = "not-between"
)

func (c Clause) IsValid() bool {
	switch c {
	case Equals, NotEquals, Contains, NotContains, StartsWith, EndsWith,
		GreaterThan, GreaterThanOrEq, LessThan, LessThanOrEq,
		In, NotIn, IsNull, IsNotNull, Between, NotBetween:
		return true
	default:
		return false
	}
}

func (c Clause) String() string {
	return string(c)
}
