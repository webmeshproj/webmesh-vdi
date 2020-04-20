package rethinkdb

// Just scaffolding - will probably stick with using the k8s api as the source
// of truth for this.

type DesktopSession struct {
	Endpoint  string `rethinkdb:"id" json:"endpoint"`
	Name      string `rethinkdb:"name" json:"name"`
	Namespace string `rethinkdb:"namespace" json:"namespace"`
	User      *User  `rethinkdb:"user_id,reference" rethinkdb_ref:"id" json:"user"`
}

func (d *rethinkDBSession) GetDesktopSession(id string) (*DesktopSession, error) {
	return nil, nil
}

func (d *rethinkDBSession) CreateDesktopSession(session *DesktopSession) (*DesktopSession, error) {
	return nil, nil
}

func (d *rethinkDBSession) DeleteDesktopSession(session *DesktopSession) error {
	return nil
}
