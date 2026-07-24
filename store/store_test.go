package store

import (
	"context"
	"crypto/rand"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vector-ops/mapil/database"
	"github.com/vector-ops/mapil/helpers"
)

type testSuite struct {
	call    func(t *testing.T, ctx context.Context, name string, st *Store)
	verify  func(t *testing.T, ctx context.Context, name string, st *Store)
	cleanup func(ctx context.Context, st *Store)
}

var testSet = map[string]testSuite{
	"AddListWithDefaultNamespace": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			expected := []database.ListType{
				{
					Key:       "test",
					Value:     []string{"val1", "val2", "val3"},
					Namespace: "default",
				},
			}
			allData := st.GetAllData(ctx)
			assert.Len(t, allData, 1, "%s: expected len to be 1 but got %d", name, len(allData))
			assert.ElementsMatch(t, expected, allData, name)
		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},

	"AddListWithNamespace": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			expected := []database.ListType{
				{
					Key:       "test",
					Value:     []string{"val1", "val2", "val3"},
					Namespace: "test-namespace",
				},
			}
			allData := st.GetAllData(ctx)
			assert.Len(t, allData, 1, "%s: expected len to be 1 but got %d", name, len(allData))
			assert.ElementsMatch(t, expected, allData, name)
		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},

	"UpdateList": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.UpdateList(ctx, "test", []string{"val4", "val5"}, "")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			expected := []database.ListType{
				{
					Key:       "test",
					Value:     []string{"val4", "val5"},
					Namespace: "default",
				},
			}
			allData := st.GetAllData(ctx)
			assert.Len(t, allData, 1, "%s: expected len to be 1 but got %d", name, len(allData))
			assert.ElementsMatch(t, expected, allData, name)
		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},

	"AppendList": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AppendList(ctx, "test", []string{"val4", "val5"}, false)
			assert.NoError(t, err, name)

		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			expected := []database.ListType{
				{
					Key:       "test",
					Value:     []string{"val1", "val2", "val3", "val4", "val5"},
					Namespace: "test-namespace",
				},
			}
			allData := st.GetAllData(ctx)
			assert.Len(t, allData, 1, "%s: expected len to be 1 but got %d", name, len(allData))
			assert.ElementsMatch(t, expected, allData, name)
		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},

	"AppendListWithDuplicates": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AppendList(ctx, "test", []string{"val4", "val2", "val5"}, true)
			assert.NoError(t, err, name)

		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			expected := []database.ListType{
				{
					Key:       "test",
					Value:     []string{"val1", "val2", "val3", "val4", "val2", "val5"},
					Namespace: "test-namespace",
				},
			}
			allData := st.GetAllData(ctx)
			assert.Len(t, allData, 1, "%s: expected len to be 1 but got %d", name, len(allData))
			assert.ElementsMatch(t, expected, allData, name)
		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},

	"AppendListWithoutDuplicatesThrowsError": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			err := st.AppendList(ctx, "test", []string{"val4", "val2", "val5"}, false)
			assert.ErrorIs(t, err, ErrDuplicateValue, name)
		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},

	"DeleteObject": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			err := st.DeleteObject(ctx, "test")
			assert.NoError(t, err, name)

			allData := st.GetAllData(ctx)
			assert.Len(t, allData, 0, "%s: expected len to be 0 but got %d", name, len(allData))
		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},

	"ReservedKeyMutationNotAllowed": {},

	"DeleteAll": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test1", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test2", []string{"val1", "val2", "val3"}, "")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()
			err := st.DeleteAll(ctx)
			assert.NoError(t, err, name)

			allData := st.GetAllData(ctx)
			assert.Len(t, allData, 0, "%s: expected len to be 0 but got %d", name, len(allData))

		},
		cleanup: func(ctx context.Context, st *Store) {},
	},
	"GetValue": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test1", []string{"val2", "val4", "val5"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test2", []string{"val5", "val6", "val7"}, "")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			keys := st.GetKeys(ctx)
			assert.Len(t, keys, 3, "%s: expected len to be 3 but got %d", name, len(keys))
			assert.ElementsMatch(t, []string{"test", "test1", "test2"}, keys, name)

			expectedValues := map[string][]string{
				"test":  {"val1", "val2", "val3"},
				"test1": {"val2", "val4", "val5"},
				"test2": {"val5", "val6", "val7"},
			}

			for k, ev := range expectedValues {
				values, err := st.GetValue(ctx, k)
				assert.NoError(t, err, name)
				assert.ElementsMatch(t, ev, values, name)
			}

		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},
	"GetNamespaceObjects": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test1", []string{"val2", "val4", "val5"}, "namespace2")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test2", []string{"val5", "val6", "val7"}, "")
			assert.NoError(t, err, name)

		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			st.GetNamespaceObjects(ctx, "test-namespace")

			expectedObjects := map[string]struct {
				key    string
				values []string
			}{
				"test-namespace": {key: "test", values: []string{"val1", "val2", "val3"}},
				"namespace2":     {key: "test1", values: []string{"val2", "val4", "val5"}},
				"default":        {key: "test2", values: []string{"val5", "val6", "val7"}},
			}

			for ns, eo := range expectedObjects {
				objs := st.GetNamespaceObjects(ctx, ns)
				assert.Len(t, objs, 1, "%s: expected len to be 1 but got %d", name, len(objs))
				assert.Equal(t, eo.key, objs[0].GetKey(), name)
				assert.ElementsMatch(t, eo.values, objs[0].GetValue(), name)
			}

		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},
	"GetNamespace": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test1", []string{"val2", "val4", "val5"}, "namespace2")
			assert.NoError(t, err, name)

		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			ns, err := st.GetNamespace(ctx, "test")
			assert.NoError(t, err, name)
			assert.Equal(t, "test-namespace", ns, name)
			_, err = st.GetNamespace(ctx, "test19")
			assert.Error(t, err, name)

		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},
	"GetAllData": {
		call: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			err := st.AddList(ctx, "test", []string{"val1", "val2", "val3"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test1", []string{"val2", "val4", "val5"}, "test-namespace")
			assert.NoError(t, err, name)
			err = st.AddList(ctx, "test2", []string{"val5", "val6", "val7"}, "")
			assert.NoError(t, err, name)
		},
		verify: func(t *testing.T, ctx context.Context, name string, st *Store) {
			t.Helper()

			keys := st.GetKeys(ctx)
			assert.Len(t, keys, 3, "%s: expected len to be 3 but got %d", name, len(keys))
			assert.ElementsMatch(t, []string{"test", "test1", "test2"}, keys, name)

			expectedValues := map[string]struct {
				ns     string
				values []string
			}{
				"test":  {ns: "test-namespace", values: []string{"val1", "val2", "val3"}},
				"test1": {ns: "test-namespace", values: []string{"val2", "val4", "val5"}},
				"test2": {ns: "default", values: []string{"val5", "val6", "val7"}},
			}

			objs := st.data.GetAllObjects(ctx)
			assert.Len(t, objs, 3, "%s: expected len to be 3 but got %d", name, len(objs))

			for _, kv := range objs {
				eo, ok := expectedValues[kv.GetKey()]
				assert.True(t, ok, "%s: found unexpected object '%s'", name, kv.GetKey())
				assert.Equal(t, eo.ns, kv.GetNamespace(), name)
				assert.ElementsMatch(t, eo.values, kv.GetValue(), name)
			}

		},
		cleanup: func(ctx context.Context, st *Store) {
			st.DeleteAll(ctx)
		},
	},
}

func TestLocalFile(t *testing.T) {

	filename := fmt.Sprintf("mapil-%s.json", rand.Text())

	cfg := helpers.Config{
		DataDir: filepath.Join(t.TempDir()),
		Databases: map[string]helpers.DBConfig{
			"file": {
				Filename: filename,
				Remote:   false,
				Primary:  true,
				Driver:   "fs",
			},
		},
	}

	st := NewStore(false, cfg)
	assert.NoError(t, st.Init(t.Context()))

	allData := st.GetAllData(context.Background())
	assert.Empty(t, allData, "expected len to be 0 but got %d", len(allData))

	for name, suit := range testSet {

		if suit.call == nil || suit.verify == nil {
			continue
		}

		suit.call(t, t.Context(), name, st)
		suit.verify(t, t.Context(), name, st)
		suit.cleanup(t.Context(), st)
	}
}

func TestSQLite(t *testing.T) {

	filename := fmt.Sprintf("mapil-%s.db", rand.Text())

	cfg := helpers.Config{
		DataDir: filepath.Join(t.TempDir()),
		Databases: map[string]helpers.DBConfig{
			"sqlite": {
				Filename: filename,
				Remote:   false,
				Primary:  true,
				Driver:   "sqlite",
			},
		},
	}

	st := NewStore(false, cfg)
	assert.NoError(t, st.Init(t.Context()))

	allData := st.GetAllData(context.Background())
	assert.Empty(t, allData, "expected len to be 0 but got %d", len(allData))

	for name, suit := range testSet {

		if suit.call == nil || suit.verify == nil {
			continue
		}

		suit.call(t, t.Context(), name, st)
		suit.verify(t, t.Context(), name, st)
		suit.cleanup(t.Context(), st)
	}
}
