package snapshot

//go:generate mockery
type Able interface {
	Load() error
	Save() error
}
