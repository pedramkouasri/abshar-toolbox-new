package contracts

type Loader interface {
	Update(service_name string, percent int)
}
