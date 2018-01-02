package persist_lib

import "cloud.google.com/go/spanner"

func BobFromDeleteBobsQuery(req BobFromDeleteBobsQueryParams) *spanner.Mutation {
	return spanner.Delete("bob_table", spanner.KeyRange{
		Start: spanner.Key{
			"Bob",
		},
		End: spanner.Key{
			"Bob",
			req.GetStartTime(),
		},
		Kind: spanner.ClosedOpen,
	})
}
func BobFromPutBobsQuery(req BobFromPutBobsQueryParams) *spanner.Mutation {
	return spanner.InsertMap("bob_table", map[string]interface{}{
		"id":         req.GetId(),
		"name":       req.GetName(),
		"start_time": req.GetStartTime(),
	})
}
func EmptyFromGetBobsQuery(req EmptyFromGetBobsQueryParams) spanner.Statement {
	return spanner.Statement{
		SQL:    "SELECT * from bob_table",
		Params: map[string]interface{}{},
	}
}
func NamesFromGetPeopleFromNamesQuery(req NamesFromGetPeopleFromNamesQueryParams) spanner.Statement {
	return spanner.Statement{
		SQL: "SELECT * FROM bob_table WHERE name IN UNNEST(@names)",
		Params: map[string]interface{}{
			"@names": req.GetNames(),
		},
	}
}

type BobFromDeleteBobsQueryParams interface {
	GetStartTime() interface{}
}
type BobFromPutBobsQueryParams interface {
	GetId() int64
	GetName() string
	GetStartTime() interface{}
}
type EmptyFromGetBobsQueryParams interface {
}
type NamesFromGetPeopleFromNamesQueryParams interface {
	GetNames() []string
}