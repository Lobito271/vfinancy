//go:build !(linux || darwin)

package bindings

func freeDiskBytes(dir string) (uint64, error) {
	return 0, nil
}
