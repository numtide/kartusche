package runtime_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/draganm/bolted/embedded"
	"github.com/draganm/kartusche/runtime"
	"github.com/stretchr/testify/require"
)

func TestRuntime(t *testing.T) {
	tf, err := os.CreateTemp("", "runtime-test")
	require.NoError(t, err)

	err = tf.Close()
	require.NoError(t, err)

	// os.Remove(tf.Name())

	db, err := embedded.Open(tf.Name(), 0700)
	require.NoError(t, err)

	err = bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
		handersPath := dbpath.ToPath("__handlers__")
		tx.CreateMap(handersPath)
		createUserPath := handersPath.Append("create_user")
		tx.CreateMap(createUserPath)
		meta := runtime.HandlerMeta{
			Method: "POST",
			Path:   "/users",
		}
		md, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("while marshalling meta data: %w", err)
		}

		tx.Put(createUserPath.Append("meta_data"), md)
		tx.Put(createUserPath.Append("source"), []byte(`
			w.Write("hello world!")
		`))
		return nil
	})

	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)
	defer func() {
		os.Remove(tf.Name())
	}()

	var rt runtime.Runtime
	t.Run("start runtime", func(t *testing.T) {
		rt, err = runtime.New(tf.Name(), "")
		require.NoError(t, err)
	})

	t.Run("execute handler", func(t *testing.T) {
		require.HTTPStatusCode(t, rt.ServeHTTP, "POST", "/users", nil, 200)
		require.HTTPBodyContains(t, rt.ServeHTTP, "POST", "/users", nil, "hello world!")
	})

	defer func() {
		if rt != nil {
			rt.Shutdown()
		}
	}()
}
