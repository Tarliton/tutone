// +build unit

package schema

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

func saveFixture(t *testing.T, n, s string) {
	wd, _ := os.Getwd()
	fileName := fmt.Sprintf("testdata/%s_%s.txt", t.Name(), n)
	t.Logf("saving fixture to %s/%s", wd, fileName)

	os.Mkdir("testdata", 0750)

	f, err := os.Create(fileName)
	require.NoError(t, err)
	f.WriteString(s)
	defer f.Close()
}

func loadFixture(t *testing.T, n string) string {
	fileName := fmt.Sprintf("testdata/%s_%s.txt", t.Name(), n)
	t.Logf("loading fixture %s", strings.TrimPrefix(fileName, "testdata/"))

	content, err := ioutil.ReadFile(fileName)
	require.NoError(t, err)

	return string(content)
}

func TestSchema_QueryArgs(t *testing.T) {
	t.Parallel()

	// schema cached by 'make test-prep'
	s, err := Load("../../testdata/schema.json")
	require.NoError(t, err)

	cases := map[string]struct {
		Name    string
		Fields  []string
		Results []QueryArg
	}{
		"entities": {
			Name:   "Actor",
			Fields: []string{"entities"},
			Results: []QueryArg{
				{Key: "guids", Value: "[EntityGuid]!"},
			},
		},
		"account": {
			Name:   "Actor",
			Fields: []string{"account"},
			Results: []QueryArg{
				{Key: "id", Value: "Int!"},
			},
		},
		"entitySearch": {
			Name:   "Actor",
			Fields: []string{"entitySearch"},
			Results: []QueryArg{
				{Key: "query", Value: "String"},
				{Key: "queryBuilder", Value: "EntitySearchQueryBuilder"},
				{Key: "sortBy", Value: "[EntitySearchSortCriteria]"},
			},
		},
		"entity": {
			Name:   "Actor",
			Fields: []string{"entity"},
			Results: []QueryArg{
				{Key: "guid", Value: "EntityGuid!"},
			},
		},
		"accountOutline": {
			Name:   "AccountOutline",
			Fields: []string{"reportingEventTypes"},
			Results: []QueryArg{
				{Key: "filter", Value: "[String]"},
				{Key: "timeWindow", Value: "TimeWindowInput"},
			},
		},
		"linkedAccounts": {
			Name:   "CloudActorFields",
			Fields: []string{"linkedAccounts"},
			Results: []QueryArg{
				{Key: "provider", Value: "String"},
			},
		},
	}

	for _, tc := range cases {
		x, err := s.LookupTypeByName(tc.Name)
		require.NoError(t, err)

		result := s.QueryArgs(x, tc.Fields)
		assert.Equal(t, tc.Results, result)
	}
}

func TestSchema_LookupTypesByFieldPath(t *testing.T) {
	t.Parallel()

	// schema cached by 'make test-prep'
	s, err := Load("../../testdata/schema.json")
	require.NoError(t, err)

	actorType, err := s.LookupTypeByName("Actor")
	require.NoError(t, err)
	cloudType, err := s.LookupTypeByName("CloudActorFields")
	require.NoError(t, err)

	cases := map[string]struct {
		FieldPath []string
		Result    []*Type
	}{
		"cloud": {
			FieldPath: []string{"actor", "cloud"},
			Result:    []*Type{actorType, cloudType},
		},
	}

	for n, tc := range cases {
		t.Logf("TestCase: %s", n)

		result, err := s.LookupQueryTypesByFieldPath(tc.FieldPath)
		require.NoError(t, err)

		require.Equal(t, len(tc.Result), len(result))

		for i := range tc.Result {
			assert.Equal(t, tc.Result[i], result[i])
		}
	}

}

func TestSchema_GetQueryStringForEndpoint(t *testing.T) {
	t.Parallel()

	// schema cached by 'make test-prep'
	s, err := Load("../../testdata/schema.json")
	require.NoError(t, err)

	cases := map[string]struct {
		Path  []string
		Field string
		Depth int
	}{
		"entitySearch": {
			Path:  []string{"actor"},
			Field: "entitySearch",
			Depth: 3,
		},
		"linkedAccounts": {
			Path:  []string{"actor", "cloud"},
			Field: "linkedAccounts",
			Depth: 2,
		},
	}

	for n, tc := range cases {
		t.Logf("TestCase: %s", n)
		typePath, err := s.LookupQueryTypesByFieldPath(tc.Path)
		require.NoError(t, err)

		result := s.GetQueryStringForEndpoint(typePath, tc.Path, tc.Field, tc.Depth)
		// saveFixture(t, n, result)
		expected := loadFixture(t, n)
		assert.Equal(t, expected, result)
	}
}

func TestSchema_GetQueryStringForMutation(t *testing.T) {
	t.Parallel()

	// schema cached by 'make test-prep'
	s, err := Load("../../testdata/schema.json")
	require.NoError(t, err)

	cases := map[string]struct {
		Mutation string
		Depth    int
	}{
		"alertsMutingRuleCreate": {
			Mutation: "alertsMutingRuleCreate",
			Depth:    3,
		},
	}

	for n, tc := range cases {
		t.Logf("TestCase: %s", n)
		field, err := s.LookupMutationByName(tc.Mutation)
		require.NoError(t, err)

		result := s.GetQueryStringForMutation(field, tc.Depth)
		// saveFixture(t, n, result)
		expected := loadFixture(t, n)
		assert.Equal(t, expected, result)
	}
}