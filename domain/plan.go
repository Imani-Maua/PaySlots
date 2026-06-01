package domain



type PlanName string

type Plan struct{
	ID int64
	Name PlanName
	PriceUSD int64
	MaxTalents *int
	MaxDepartments *int
	MLEnabled bool
}

const (
	FreePlan PlanName = "free"
	StarterPlan PlanName = "starter"
	ProPlan PlanName = "pro"
)


func (plan * Plan) WithinLimit(limit *int, count int) bool{
	if limit == nil{
		return true
	}

	return count <= *limit
}