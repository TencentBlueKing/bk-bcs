package gw

// MockClient mock gw client
type MockClient struct {
	svcs []*Service
}

// Run implements interface
func (mc *MockClient) Run() error {
	return nil
}

// Update implements interface
func (mc *MockClient) Update(svcs []*Service) error {
	mc.svcs = svcs
	return nil
}

// Delete implements interface
func (mc *MockClient) Delete(svcs []*Service) error {
	for _, svcToDel := range svcs {
		for index, svc := range mc.svcs {
			if svc.Key() == svcToDel.Key() {
				mc.svcs = append(mc.svcs[index:], mc.svcs[:index+1]...)
				break
			}
		}
	}

	return nil
}

// List list services
func (mc *MockClient) List() ([]*Service, error) {
	return mc.svcs, nil
}
