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

func TestSchema_BuildQueryArgsForEndpoint(t *testing.T) {
	t.Parallel()

	// schema cached by 'make test-prep'
	s, err := Load("../../testdata/schema.json")
	require.NoError(t, err)

	cases := map[string]struct {
		Name        string
		Fields      []string
		IncludeArgs []string
		Results     []QueryArg
	}{
		"accountEntities": {
			Name:   "Actor",
			Fields: []string{"account", "entities"},
			Results: []QueryArg{
				{Key: "id", Value: "Int!"},
				{Key: "guids", Value: "[EntityGuid]!"},
			},
		},
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
			IncludeArgs: []string{
				"options",
				"query",
				"queryBuilder",
				"sortBy",
			},
			Results: []QueryArg{
				{Key: "options", Value: "EntitySearchOptions"},
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
			Name:        "AccountOutline",
			Fields:      []string{"reportingEventTypes"},
			IncludeArgs: []string{"filter", "timeWindow"},
			Results: []QueryArg{
				{Key: "filter", Value: "[String]"},
				{Key: "timeWindow", Value: "TimeWindowInput"},
			},
		},
		"linkedAccounts": {
			Name:        "CloudActorFields",
			Fields:      []string{"linkedAccounts"},
			IncludeArgs: []string{"provider"},
			Results: []QueryArg{
				{Key: "provider", Value: "String"},
			},
		},
		"linkedAccountsWithoutNullable": {
			Name:    "CloudActorFields",
			Fields:  []string{"linkedAccounts"},
			Results: []QueryArg{},
		},
		"linkedAccountsWithInvalidIncludeArgument": {
			Name:        "CloudActorFields",
			Fields:      []string{"linkedAccounts"},
			IncludeArgs: []string{"this-argument-does-not-exist"},
			Results:     []QueryArg{},
		},
	}

	for _, tc := range cases {
		x, err := s.LookupTypeByName(tc.Name)
		require.NoError(t, err)

		result := s.BuildQueryArgsForEndpoint(x, tc.Fields, tc.IncludeArgs)
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
		Path        []string
		Field       string
		Depth       int
		IncludeArgs []string
	}{
		"entitySearch": {
			Path:  []string{"actor"},
			Field: "entitySearch",
			Depth: 3,
		},
		"entitySearchArgs": {
			Path:        []string{"actor"},
			Field:       "entitySearch",
			Depth:       3,
			IncludeArgs: []string{"query"},
		},
		"entities": {
			Path:  []string{"actor"},
			Field: "entities",
			// Zero set here because we have the field coverage above with greater depth.  Here we want to ensure that required arguments on the entities endpoint has the correct syntax.
			Depth: 0,
		},
		"linkedAccounts": {
			Path:        []string{"actor", "cloud"},
			Field:       "linkedAccounts",
			Depth:       2,
			IncludeArgs: []string{"provider"},
		},
		"policy": {
			Path:  []string{"actor", "account", "alerts"},
			Field: "policy",
			Depth: 2,
		},
	}

	for n, tc := range cases {
		t.Logf("TestCase: %s", n)
		typePath, err := s.LookupQueryTypesByFieldPath(tc.Path)
		require.NoError(t, err)

		result := s.GetQueryStringForEndpoint(typePath, tc.Path, tc.Field, tc.Depth, tc.IncludeArgs)
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
		Override map[string]string
	}{
		"alertsMutingRuleCreate": {
			Mutation: "alertsMutingRuleCreate",
			Depth:    3,
			Override: map[string]string{},
		},
		"cloudRenameAccount": {
			Mutation: "cloudRenameAccount",
			Depth:    1,
			Override: map[string]string{
				"accountId": "Int!",
				"accounts":  "[CloudRenameAccountsInput!]!",
			},
		},
	}

	for n, tc := range cases {
		t.Logf("TestCase: %s", n)
		field, err := s.LookupMutationByName(tc.Mutation)
		require.NoError(t, err)

		result := s.GetQueryStringForMutation(field, tc.Depth, tc.Override)
		// saveFixture(t, n, result)
		expected := loadFixture(t, n)
		assert.Equal(t, expected, result)
	}
}

func TestSchema_GetInputFieldsForQueryPath(t *testing.T) {
	t.Parallel()

	// schema cached by 'make test-prep'
	s, err := Load("../../testdata/schema.json")
	require.NoError(t, err)

	cases := map[string]struct {
		QueryPath []string
		Fields    map[string][]string
	}{
		"accountCloud": {
			QueryPath: []string{"actor", "account", "cloud"},
			Fields: map[string][]string{
				"account": {"id"},
			},
		},
		"entities": {
			QueryPath: []string{"actor", "entities"},
			Fields: map[string][]string{
				"entities": {"guids"},
			},
		},
		"apiAccessKey": {
			QueryPath: []string{"actor", "apiAccess", "key"},
			Fields: map[string][]string{
				"key": {"id", "keyType"},
			},
		},
	}

	for _, tc := range cases {
		result := s.GetInputFieldsForQueryPath(tc.QueryPath)
		assert.Equal(t, len(tc.Fields), len(result))
		for pathName, fields := range tc.Fields {

			for i, name := range fields {
				assert.Equal(t, name, result[pathName][i].Name)
			}
		}
	}
}
