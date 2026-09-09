package hydrate

import "fmt"

type ServiceHydrateError struct {
	Service string
	Problem string
}

func (e *ServiceHydrateError) Error() string {
	return fmt.Sprintf("service hydration error (%s): %v", e.Service, e.Problem)
}
