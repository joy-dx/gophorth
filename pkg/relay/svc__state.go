package relay

import (
	"errors"
)

func (s *RelaySvc) Hydrate() error {
	if s.cfg == nil {
		return errors.New("cfg is nil")
	}
	return nil
}
