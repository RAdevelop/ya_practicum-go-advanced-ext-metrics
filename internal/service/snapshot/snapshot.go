package snapshot

type Able interface {
	Load() error
	Save() error
}
