package sub

type (
	Sub interface {
		Process() error
		Close()
	}
)
